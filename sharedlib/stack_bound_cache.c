#include "stack_bound_cache.h"

#include <stdatomic.h>
#include <stddef.h>
#include <stdint.h>

#if \
    defined(__APPLE__) || defined(__GLIBC__) || \
    defined(__sun)
#define LIGHTER_HAS_EXACT_STACK_BOUNDS 1
#endif

#if defined(LIGHTER_HAS_EXACT_STACK_BOUNDS)
static _Atomic(LighterStackBoundFunction) original_stack_bound_function = NULL;
static _Thread_local uintptr_t cached_stack_low = 0;
static _Thread_local uintptr_t cached_stack_high = 0;
static atomic_flag setup_lock = ATOMIC_FLAG_INIT;

static void lock_setup(void) {
    while (atomic_flag_test_and_set_explicit(
        &setup_lock, memory_order_acquire
    )) {
        atomic_signal_fence(memory_order_seq_cst);
    }
}

static void unlock_setup(void) {
    atomic_flag_clear_explicit(&setup_lock, memory_order_release);
}

static void lighter_cached_stack_bounds(uintptr_t bounds[2]) {
    char stack_marker;
    uintptr_t stack_pointer = (uintptr_t)&stack_marker;

    if (
        cached_stack_low != 0 &&
        stack_pointer > cached_stack_low &&
        stack_pointer <= cached_stack_high
    ) {
        bounds[0] = cached_stack_low;
        bounds[1] = cached_stack_high;
        return;
    }

    /*
     * Go deliberately refreshes these bounds because a C coroutine may have
     * switched stacks. Reuse only bounds that were returned by the runtime's
     * exact pthread lookup and that still contain the current stack pointer.
     * A switched stack falls through to the original lookup and is not cached
     * unless those exact bounds contain it.
     */
    LighterStackBoundFunction original = atomic_load_explicit(
        &original_stack_bound_function, memory_order_acquire
    );
    if (original == NULL) {
        bounds[0] = 0;
        bounds[1] = 0;
        return;
    }
    original(bounds);
    if (
        bounds[0] != 0 &&
        stack_pointer > bounds[0] &&
        stack_pointer <= bounds[1]
    ) {
        cached_stack_low = bounds[0];
        cached_stack_high = bounds[1];
    }
}
#endif

LighterStackBoundFunction lighter_enable_stack_bound_cache(
    LighterStackBoundFunction original
) {
#if !defined(LIGHTER_HAS_EXACT_STACK_BOUNDS)
    /*
     * Windows is already a no-op. Other libc fallbacks estimate bounds from
     * the current frame, so caching them would weaken Go's stack-switching
     * protection.
     */
    (void)original;
    return NULL;
#else
    lock_setup();
    if (original == lighter_cached_stack_bounds) {
        unlock_setup();
        return original;
    }
    if (original == NULL) {
        unlock_setup();
        return NULL;
    }
    atomic_store_explicit(
        &original_stack_bound_function, original, memory_order_release
    );
    unlock_setup();
    return lighter_cached_stack_bounds;
#endif
}

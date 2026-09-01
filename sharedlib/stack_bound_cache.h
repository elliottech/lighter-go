#ifndef LIGHTER_STACK_BOUND_CACHE_H
#define LIGHTER_STACK_BOUND_CACHE_H

#include <stdint.h>

typedef void (*LighterStackBoundFunction)(uintptr_t bounds[2]);

LighterStackBoundFunction lighter_enable_stack_bound_cache(
    LighterStackBoundFunction original
);

#endif

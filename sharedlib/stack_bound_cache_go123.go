//go:build cgo && !go1.26

package main

/*
#include <stdint.h>
typedef void (*LighterStackBoundFunction)(uintptr_t bounds[2]);
LighterStackBoundFunction lighter_enable_stack_bound_cache(
    LighterStackBoundFunction original
);
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	stackBoundCacheEnableOnce   sync.Once
	stackBoundCacheEnableResult C.int
)

// The standard Go runtime asks libc to rediscover a foreign thread's stack
// bounds on every C-to-Go callback. The optional cache keeps exact pthread
// bounds per thread and falls back whenever a caller has switched stacks.
//
// WARNING: _cgo_getstackbound is a private Go runtime hook with no compatibility
// guarantee. This integration is limited by the build tag above, tied to the Go
// toolchain in go.mod, and must be revalidated before every toolchain upgrade.
//
//go:linkname cgoGetStackbound _cgo_getstackbound
var cgoGetStackbound unsafe.Pointer

//export FastEnableStackBoundCache
func FastEnableStackBoundCache() C.int {
	stackBoundCacheEnableOnce.Do(func() {
		original := atomic.LoadPointer(&cgoGetStackbound)
		cached := C.lighter_enable_stack_bound_cache(
			(C.LighterStackBoundFunction)(original),
		)
		if cached != nil {
			atomic.StorePointer(&cgoGetStackbound, unsafe.Pointer(cached))
			stackBoundCacheEnableResult = 1
		}
	})
	return stackBoundCacheEnableResult
}

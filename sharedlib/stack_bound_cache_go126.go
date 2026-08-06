//go:build cgo && go1.26

package main

/*
#include <stdint.h>
*/
import "C"

// FastEnableStackBoundCache is disabled on unvalidated Go toolchains. The
// private runtime hook must be re-audited before raising the version guard.
//
//export FastEnableStackBoundCache
func FastEnableStackBoundCache() C.int {
	return 0
}

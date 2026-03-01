//go:build !byollvm && linux && llvm21

package llvm

// #cgo CPPFLAGS: -I/usr/include/llvm-21 -I/usr/include/llvm-c-21 -D_GNU_SOURCE -D__STDC_CONSTANT_MACROS -D__STDC_FORMAT_MACROS -D__STDC_LIMIT_MACROS
// #cgo CXXFLAGS: -std=c++17
// #cgo LDFLAGS: -L/usr/lib/llvm-21/lib -lLLVM-21
import "C"

type run_build_sh int

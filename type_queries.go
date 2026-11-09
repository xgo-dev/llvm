// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

/*
#include "IntrinsicBindings.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

const (
	HalfTypeKind           TypeKind = C.LLVMHalfTypeKind
	BFloatTypeKind         TypeKind = C.LLVMBFloatTypeKind
	ScalableVectorTypeKind TypeKind = C.LLVMScalableVectorTypeKind
)

func (c Context) HalfType() (t Type)   { t.C = C.LLVMHalfTypeInContext(c.C); return }
func (c Context) BFloatType() (t Type) { t.C = C.LLVMBFloatTypeInContext(c.C); return }

// ArrayType64 and ArrayLength64 preserve LLVM's full 64-bit element count.
// The older ArrayType/ArrayLength APIs use LLVM's 32-bit C entry points.
func ArrayType64(element Type, count uint64) (t Type) {
	t.C = C.LLVMGoArrayType64(element.C, C.uint64_t(count))
	return
}
func (t Type) ArrayLength64() uint64 { return uint64(C.LLVMGoArrayLength64(t.C)) }

// ScalableVectorType constructs a vector with vscale * minimum elements.
// LLVM requires a valid vector element type and nonzero minimum.
func ScalableVectorType(element Type, minimum int) (t Type) {
	t.C = C.LLVMScalableVectorType(element.C, C.unsigned(minimum))
	return
}

// VectorElementCount returns the fixed count or scalable minimum. Non-vectors
// return (0, false). TypeKind distinguishes a non-vector from a fixed vector.
func (t Type) VectorElementCount() (minimum int, scalable bool) {
	if t.IsNil() {
		return
	}
	switch t.TypeKind() {
	case VectorTypeKind:
		return t.VectorSize(), false
	case ScalableVectorTypeKind:
		return t.VectorSize(), true
	}
	return
}

// ScalarSizeInBits returns LLVM's scalar bit width, including vector elements.
// Non-scalar/aggregate types return zero, as in LLVM's getScalarSizeInBits.
func (t Type) ScalarSizeInBits() uint {
	if t.IsNil() {
		return 0
	}
	return uint(C.LLVMGoScalarBits(t.C))
}

// ConstIntInRange reports whether v is a scalar integer literal in [low, high),
// interpreting its exact APInt as signed, including widths above 64 bits.
func (v Value) ConstIntInRange(low, high int64) bool {
	return C.LLVMGoConstIntInRange(v.C, C.uint64_t(low), C.uint64_t(high), 0) != 0
}

// ConstUintInRange is the unsigned counterpart of ConstIntInRange.
func (v Value) ConstUintInRange(low, high uint64) bool {
	return C.LLVMGoConstIntInRange(v.C, C.uint64_t(low), C.uint64_t(high), 1) != 0
}

// ParseTargetData parses a layout without aborting on malformed input. Dispose
// the returned TargetData after use. NewTargetData retains its original API.
func ParseTargetData(layout string) (td TargetData, err error) {
	var message *C.char
	td = parseTargetData(layout, &message)
	if message != nil {
		defer C.LLVMDisposeMessage(message)
		err = errors.New(C.GoString(message))
	}
	return
}

func parseTargetData(layout string, message **C.char) (td TargetData) {
	c := C.CString(layout)
	defer C.free(unsafe.Pointer(c))
	td.C = C.LLVMGoParseTargetData(c, message)
	return
}

// IndexTypeSizeInBits returns the pointer index width for a pointer or vector
// of pointers, preserving the address space. It can differ from pointer size.
// Returns zero for a nil layout or a non-pointer type.
func (td TargetData) IndexTypeSizeInBits(t Type) uint {
	return uint(C.LLVMGoIndexTypeSizeInBits(td.C, t.C))
}

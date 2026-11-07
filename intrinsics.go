// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

/*
#include "IntrinsicBindings.h"
*/
import "C"

import "unsafe"

func llvmTypeRefs(types []Type) (*C.LLVMTypeRef, int) {
	if len(types) == 0 {
		return nil, 0
	}
	return llvmTypeRefPtr(&types[0]), len(types)
}

// IntrinsicCount is the exclusive upper bound of the linked LLVM's intrinsic
// IDs. Zero is not an intrinsic. IDs are local to the linked LLVM version.
func IntrinsicCount() int { return int(C.LLVMGoIntrinsicCount()) }

func validIntrinsicID(id int) bool { return id > 0 && id < IntrinsicCount() }

// IntrinsicIsTarget reports whether id names a target-specific intrinsic.
func IntrinsicIsTarget(id int) bool {
	return validIntrinsicID(id) && C.LLVMGoIsTargetIntrinsic(C.unsigned(id)) != 0
}

// IntrinsicName returns the base name, without an overload suffix, or "" for
// an invalid ID. It also works for overloaded intrinsics.
func IntrinsicName(id int) string {
	if !validIntrinsicID(id) {
		return ""
	}
	var n C.size_t
	p := C.LLVMIntrinsicGetName(C.unsigned(id), &n)
	return C.GoStringN(p, C.int(n))
}

func IntrinsicIsOverloaded(id int) bool {
	return validIntrinsicID(id) && C.LLVMIntrinsicIsOverloaded(C.unsigned(id)) != 0
}

// MatchIntrinsicSignature validates the complete function type, including its
// vararg flag, and infers its overload types. It creates no module, declaration
// or call. This checks LLVM's type metadata, not semantic operand constraints
// or target features; verify the completed IR separately.
func MatchIntrinsicSignature(id int, signature Type) (overloads []Type, ok bool) {
	if !validIntrinsicID(id) {
		return nil, false
	}
	r := C.LLVMGoMatchIntrinsicSignature(C.unsigned(id), signature.C)
	defer C.LLVMGoDisposeIntrinsicSignature(r)
	if r.matched == 0 {
		return nil, false
	}
	if r.count == 0 {
		return nil, true
	}
	overloads = make([]Type, int(r.count))
	for i, typ := range (*[1 << 28]C.LLVMTypeRef)(unsafe.Pointer(r.types))[:int(r.count):int(r.count)] {
		overloads[i].C = typ
	}
	return overloads, true
}

// IntrinsicType obtains a canonical function type from valid overload types.
// The overload list must satisfy LLVM's metadata and belong to c; use
// MatchIntrinsicSignature to infer it from a candidate function type.
func (c Context) IntrinsicType(id int, overloads []Type) (t Type) {
	if c.IsNil() || !validIntrinsicID(id) {
		return
	}
	p, n := llvmTypeRefs(overloads)
	t.C = C.LLVMIntrinsicGetType(c.C, C.unsigned(id), p, C.size_t(n))
	return
}

// IntrinsicName obtains the overload name in m. In particular, unnamed
// identified structs receive a module-local suffix. Overloads must be valid
// and belong to m's context. This does not insert a declaration.
func (m Module) IntrinsicName(id int, overloads []Type) string {
	if m.IsNil() || !validIntrinsicID(id) {
		return ""
	}
	if !IntrinsicIsOverloaded(id) {
		return IntrinsicName(id)
	}
	p, n := llvmTypeRefs(overloads)
	var size C.size_t
	name := C.LLVMIntrinsicCopyOverloadedName2(m.C, C.unsigned(id), p, C.size_t(n), &size)
	defer C.LLVMDisposeMessage(name)
	return C.GoStringN(name, C.int(size))
}

// IntrinsicDeclaration gets or inserts LLVM's intrinsic declaration, including
// its attributes. Overloads must be valid and belong to m's context. As with
// LLVMGetIntrinsicDeclaration, an existing symbol must have the canonical
// function type; callers can check NamedValue before requesting insertion.
func (m Module) IntrinsicDeclaration(id int, overloads []Type) (v Value) {
	if m.IsNil() || !validIntrinsicID(id) {
		return
	}
	p, n := llvmTypeRefs(overloads)
	v.C = C.LLVMGetIntrinsicDeclaration(m.C, C.unsigned(id), p, C.size_t(n))
	return
}

// IntrinsicImmediateArguments returns zero-based parameter indices carrying
// LLVM's immarg attribute. Signature must be the matched canonical function
// type. Additional constraints absent from LLVM's metadata are not inferred.
func IntrinsicImmediateArguments(id int, signature Type) []int {
	if !validIntrinsicID(id) {
		return nil
	}
	n := int(C.LLVMGoIntrinsicImmArgs(C.unsigned(id), signature.C, nil))
	if n == 0 {
		return nil
	}
	raw := make([]C.unsigned, n)
	C.LLVMGoIntrinsicImmArgs(C.unsigned(id), signature.C, &raw[0])
	out := make([]int, n)
	for i, value := range raw {
		out[i] = int(value)
	}
	return out
}

// IITDescriptor is one record in LLVM's intrinsic type metadata. Kind uses
// LLVM's descriptor names, with HalfVecArgument normalized to
// OneNthEltsVecArgument (Value=2) on older LLVM versions.
//
// Value is an integer bit width, vector minimum lane count, pointer address
// space, struct field count or vector divisor, according to Kind. Argument is
// an overload index; ArgumentKind is LLVM's overload domain (0=any, 1=integer,
// 2=float, 3=vector, 4=pointer, 7=match). Reference is the source overload index
// for OneNthEltsVecArgument and VecOfAnyPtrsToElt. Scalable applies to Vector.
// Fields without a meaning for Kind are zero. The result descriptor precedes
// the parameter descriptors; composite types are encoded recursively.
type IITDescriptor struct {
	Kind         string
	Value        uint32
	Argument     uint32
	ArgumentKind uint32
	Reference    uint32
	Scalable     bool
}

// IntrinsicDescriptors reads the linked LLVM's IIT table. No instruction list
// or signature table is maintained by this binding. Returns nil for invalid IDs.
func IntrinsicDescriptors(id int) []IITDescriptor {
	if !validIntrinsicID(id) {
		return nil
	}
	n := int(C.LLVMGoIntrinsicDescriptors(C.unsigned(id), nil))
	if n == 0 {
		return nil
	}
	raw := make([]C.LLVMGoIITDescriptor, n)
	C.LLVMGoIntrinsicDescriptors(C.unsigned(id), &raw[0])
	out := make([]IITDescriptor, n)
	for i, d := range raw {
		out[i] = IITDescriptor{Kind: C.GoString(d.kind), Value: uint32(d.value),
			Argument: uint32(d.argument), ArgumentKind: uint32(d.argument_kind),
			Reference: uint32(d.reference), Scalable: d.scalable != 0}
	}
	return out
}

//===- string.go - Stringer implementation for Type -----------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This file implements the Stringer interface for the Type type.
//
//===----------------------------------------------------------------------===//

package llvm

//#include "llvm-c/Core.h"
import "C"

func (t TypeKind) String() string {
	switch t {
	case VoidTypeKind:
		return "VoidTypeKind"
	case FloatTypeKind:
		return "FloatTypeKind"
	case DoubleTypeKind:
		return "DoubleTypeKind"
	case X86_FP80TypeKind:
		return "X86_FP80TypeKind"
	case FP128TypeKind:
		return "FP128TypeKind"
	case PPC_FP128TypeKind:
		return "PPC_FP128TypeKind"
	case LabelTypeKind:
		return "LabelTypeKind"
	case IntegerTypeKind:
		return "IntegerTypeKind"
	case FunctionTypeKind:
		return "FunctionTypeKind"
	case StructTypeKind:
		return "StructTypeKind"
	case ArrayTypeKind:
		return "ArrayTypeKind"
	case PointerTypeKind:
		return "PointerTypeKind"
	case VectorTypeKind:
		return "VectorTypeKind"
	case MetadataTypeKind:
		return "MetadataTypeKind"
	}
	panic("unreachable")
}

func (t Type) String() string {
	return C.GoString(C.LLVMPrintTypeToString(t.C))
}

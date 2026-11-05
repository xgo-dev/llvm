// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

/*
#include "transforms.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// ConstantAllocationSize queries LLVM's allocation-size analysis using the
// module's target triple and data layout. It returns false for unknown sizes,
// non-call values, or sizes that do not fit uint64. A known size does not imply
// that the allocation is removable or safe to move to the stack.
func (v Value) ConstantAllocationSize() (size uint64, known bool) {
	var bytes C.uint64_t
	known = C.LLVMGoGetConstantAllocationSize(v.C, &bytes) != 0
	return uint64(bytes), known
}

// CloneFunction copies a function into its module, including its attributes,
// blocks, instructions and debug information. Local references are remapped;
// references to other functions and globals remain unchanged. The caller can
// rename the returned function. The original function is not modified.
// It panics if v is nil or is not a function; declarations may be cloned.
func (v Value) CloneFunction() (clone Value) {
	clone.C = C.LLVMGoCloneFunction(v.C)
	if clone.IsNil() {
		panic("llvm: CloneFunction requires a function")
	}
	return
}

// SimplifyInstructionsInBlock folds instructions using locally available facts.
// It returns whether the block changed and may erase instructions.
func (bb BasicBlock) SimplifyInstructionsInBlock() bool {
	return C.LLVMGoSimplifyInstructionsInBlock(bb.C) != 0
}

// ConstantFoldTerminator folds a constant branch or switch and updates successor
// PHIs. When deleteDead is true, it also removes dead condition instructions.
func (bb BasicBlock) ConstantFoldTerminator(deleteDead bool) bool {
	return C.LLVMGoConstantFoldTerminator(bb.C, boolToLLVMBool(deleteDead)) != 0
}

// RemoveUnreachableBlocks deletes blocks unreachable from the function entry
// and updates successor PHIs. It returns whether the function changed.
// It panics if v is nil or is not a function definition.
func (v Value) RemoveUnreachableBlocks() bool {
	changed := C.LLVMGoRemoveUnreachableBlocks(v.C)
	if changed < 0 {
		panic("llvm: RemoveUnreachableBlocks requires a function definition")
	}
	return changed != 0
}

// SetSubprogramLinkageName changes the linker name of a DISubprogram without
// changing its source-language display name. It panics if md is nil or is not
// a DISubprogram.
func (md Metadata) SetSubprogramLinkageName(name string) {
	text := C.CString(name)
	defer C.free(unsafe.Pointer(text))
	if C.LLVMGoSetSubprogramLinkageName(md.C, text, C.size_t(len(name))) == 0 {
		panic("llvm: SetSubprogramLinkageName requires a DISubprogram")
	}
}

// AsMetadata unwraps a MetadataAsValue, or wraps a constant, argument, or
// instruction as metadata using LLVMValueAsMetadata.
func (v Value) AsMetadata() (md Metadata) {
	md.C = C.LLVMValueAsMetadata(v.C)
	return
}

// TailCallKind distinguishes optional tail-call hints from mandatory tail calls.
type TailCallKind uint32

const (
	// These values match llvm::CallInst::TailCallKind, checked in transforms.cpp.
	TailCallKindNone TailCallKind = iota
	TailCallKindTail
	TailCallKindMustTail
	TailCallKindNoTail
)

// TailCallKind returns the tail-call kind of a call instruction.
// It panics if v is nil or is not a call instruction.
func (v Value) TailCallKind() TailCallKind {
	kind := TailCallKind(C.LLVMGoGetTailCallKind(v.C))
	if kind > TailCallKindNoTail {
		panic("llvm: TailCallKind requires a call instruction")
	}
	return kind
}

// SetTailCallKind sets the tail-call kind of a call instruction. The caller must
// satisfy LLVM's ABI and control-flow requirements for mandatory tail calls.
// It panics if v is nil, is not a call instruction, or kind is invalid.
func (v Value) SetTailCallKind(kind TailCallKind) {
	if C.LLVMGoSetTailCallKind(v.C, C.unsigned(kind)) == 0 {
		panic("llvm: SetTailCallKind requires a call instruction and a valid tail-call kind")
	}
}

// InlineAsmInfo contains all properties needed to rebuild an inline assembly
// value while changing its text or constraints.
type InlineAsmInfo struct {
	Type                                   Type
	Assembly, Constraints                  string
	HasSideEffects, IsAlignStack, CanThrow bool
	Dialect                                InlineAsmDialect
}

// InlineAsmInfo returns the properties of an inline assembly value.
// It panics if v is nil or is not an inline assembly value.
func (v Value) InlineAsmInfo() InlineAsmInfo {
	a := C.LLVMGoGetInlineAsmInfo(v.C)
	if a.Type == nil {
		panic("llvm: InlineAsmInfo requires an inline assembly value")
	}
	return InlineAsmInfo{
		Type:           Type{C: a.Type},
		Assembly:       C.GoStringN(a.Assembly, C.int(a.AssemblyLen)),
		Constraints:    C.GoStringN(a.Constraints, C.int(a.ConstraintsLen)),
		HasSideEffects: a.SideEffects != 0,
		IsAlignStack:   a.AlignStack != 0,
		CanThrow:       a.CanThrow != 0,
		Dialect:        InlineAsmDialect(a.Dialect),
	}
}

// NamedValue looks up a symbol of any global kind, including functions, global
// variables, aliases and indirect functions. It returns a nil Value if absent.
func (m Module) NamedValue(name string) (value Value) {
	text := C.CString(name)
	defer C.free(unsafe.Pointer(text))
	value.C = C.LLVMGoGetNamedValue(m.C, text, C.size_t(len(name)))
	return
}

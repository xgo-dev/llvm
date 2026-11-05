// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

// Generic IR transformation utilities missing from llvm-c.
#ifndef LLVM_BINDINGS_GO_TRANSFORMS_H
#define LLVM_BINDINGS_GO_TRANSFORMS_H
#include "llvm-c/Core.h"
#ifdef __cplusplus
extern "C" {
#endif
LLVMValueRef LLVMGoGetNamedValue(LLVMModuleRef M, const char *Name, size_t Len);
LLVMBool LLVMGoGetConstantAllocationSize(LLVMValueRef Call, uint64_t *Size);
// Invalid inputs are reported to the Go layer, which raises a Go panic.
LLVMValueRef LLVMGoCloneFunction(LLVMValueRef Fn); // null on invalid input
LLVMBool LLVMGoSimplifyInstructionsInBlock(LLVMBasicBlockRef BB);
LLVMBool LLVMGoConstantFoldTerminator(LLVMBasicBlockRef BB,
                                      LLVMBool DeleteDead);
// Returns -1 on invalid input, otherwise whether the function changed.
int LLVMGoRemoveUnreachableBlocks(LLVMValueRef Fn);
LLVMBool LLVMGoSetSubprogramLinkageName(LLVMMetadataRef SP, const char *Name,
                                        size_t Len);
LLVMBool LLVMGoSetTailCallKind(LLVMValueRef Call, unsigned Kind);
// Returns ~0u on invalid input.
unsigned LLVMGoGetTailCallKind(LLVMValueRef Call);
struct LLVMGoInlineAsmInfo {
  LLVMTypeRef Type;
  const char *Assembly, *Constraints;
  size_t AssemblyLen, ConstraintsLen;
  LLVMBool SideEffects, AlignStack, CanThrow;
  LLVMInlineAsmDialect Dialect;
};
struct LLVMGoInlineAsmInfo LLVMGoGetInlineAsmInfo(LLVMValueRef Asm);
#ifdef __cplusplus
}
#endif
#endif

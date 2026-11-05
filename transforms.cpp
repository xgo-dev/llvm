// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

// Generic LLVM utilities exposed to Go; policy belongs to the caller.
#include "transforms.h"
#include "llvm/Analysis/MemoryBuiltins.h"
#include "llvm/Analysis/TargetLibraryInfo.h"
#include "llvm/IR/DebugInfoMetadata.h"
#include "llvm/IR/InlineAsm.h"
#include "llvm/IR/Instructions.h"
#include "llvm/IR/Module.h"
#include "llvm/Transforms/Utils/Cloning.h"
#include "llvm/Transforms/Utils/Local.h"
using namespace llvm;

// Keep the Go TailCallKind constants and the C boundary in sync with LLVM.
static_assert(CallInst::TCK_None == 0 && CallInst::TCK_Tail == 1 &&
                  CallInst::TCK_MustTail == 2 && CallInst::TCK_NoTail == 3,
              "update the Go TailCallKind constants for this LLVM version");

LLVMBool LLVMGoGetConstantAllocationSize(LLVMValueRef Call, uint64_t *Size) {
  auto *CB = Call ? dyn_cast<CallBase>(unwrap(Call)) : nullptr;
  if (!CB || !CB->getParent() || !CB->getFunction() || !CB->getModule() ||
      !CB->getType()->isPointerTy() || !Size)
    return false;
  FunctionAnalysisManager FAM;
  TargetLibraryAnalysis TLA;
  auto TLI = TLA.run(*CB->getFunction(), FAM);
  auto Bytes = getAllocSize(CB, &TLI, [](const Value *V) { return V; });
  if (!Bytes || Bytes->getActiveBits() > 64)
    return false;
  *Size = Bytes->getZExtValue();
  return true;
}

LLVMValueRef LLVMGoGetNamedValue(LLVMModuleRef M, const char *Name,
                                 size_t Len) {
  return wrap(unwrap(M)->getNamedValue(StringRef(Name, Len)));
}
LLVMValueRef LLVMGoCloneFunction(LLVMValueRef Fn) {
  auto *F = Fn ? dyn_cast<Function>(unwrap(Fn)) : nullptr;
  if (!F)
    return nullptr;
  ValueToValueMapTy Map;
  return wrap(CloneFunction(F, Map));
}
LLVMBool LLVMGoSimplifyInstructionsInBlock(LLVMBasicBlockRef BB) {
  return SimplifyInstructionsInBlock(unwrap(BB));
}
LLVMBool LLVMGoConstantFoldTerminator(LLVMBasicBlockRef BB,
                                      LLVMBool DeleteDead) {
  return ConstantFoldTerminator(unwrap(BB), DeleteDead);
}
int LLVMGoRemoveUnreachableBlocks(LLVMValueRef Fn) {
  auto *F = Fn ? dyn_cast<Function>(unwrap(Fn)) : nullptr;
  if (!F || F->isDeclaration())
    return -1;
  return removeUnreachableBlocks(*F);
}
LLVMBool LLVMGoSetSubprogramLinkageName(LLVMMetadataRef SP, const char *Name,
                                        size_t Len) {
  auto *Subprogram = SP ? dyn_cast<DISubprogram>(unwrap(SP)) : nullptr;
  if (!Subprogram)
    return false;
  Subprogram->replaceLinkageName(
      MDString::get(Subprogram->getContext(), StringRef(Name, Len)));
  return true;
}
LLVMBool LLVMGoSetTailCallKind(LLVMValueRef Call, unsigned Kind) {
  auto *CI = Call ? dyn_cast<CallInst>(unwrap(Call)) : nullptr;
  if (!CI || Kind > CallInst::TCK_NoTail)
    return false;
  CI->setTailCallKind(static_cast<CallInst::TailCallKind>(Kind));
  return true;
}
unsigned LLVMGoGetTailCallKind(LLVMValueRef Call) {
  auto *CI = Call ? dyn_cast<CallInst>(unwrap(Call)) : nullptr;
  return CI ? CI->getTailCallKind() : ~0u;
}
LLVMGoInlineAsmInfo LLVMGoGetInlineAsmInfo(LLVMValueRef Asm) {
  auto *A = Asm ? dyn_cast<InlineAsm>(unwrap(Asm)) : nullptr;
  if (!A)
    return {};
  // Borrow InlineAsm's storage with either the older std::string& API or the
  // newer StringRef API. Copying a std::string here would return dangling data.
  StringRef Text = A->getAsmString();
  StringRef Constraints = A->getConstraintString();
  return {wrap(A->getFunctionType()),
          Text.data(),
          Constraints.data(),
          Text.size(),
          Constraints.size(),
          A->hasSideEffects(),
          A->isAlignStack(),
          A->canThrow(),
          static_cast<LLVMInlineAsmDialect>(A->getDialect())};
}

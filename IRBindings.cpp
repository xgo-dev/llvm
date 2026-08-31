//===- IRBindings.cpp - Additional bindings for ir ------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This file defines additional C bindings for the ir component.
//
//===----------------------------------------------------------------------===//

#include "IRBindings.h"
#include "llvm/ADT/SmallVector.h"
#include "llvm/IR/Attributes.h"
#include "llvm/IR/DebugLoc.h"
#include "llvm/IR/DebugInfoMetadata.h"
#include "llvm/IR/Function.h"
#include "llvm/IR/IRBuilder.h"
#include "llvm/IR/Intrinsics.h"
#include "llvm/IR/LLVMContext.h"
#include "llvm/IR/Module.h"

using namespace llvm;

LLVMMetadataRef LLVMConstantAsMetadata(LLVMValueRef C) {
  return wrap(ConstantAsMetadata::get(unwrap<Constant>(C)));
}

LLVMMetadataRef LLVMMDString2(LLVMContextRef C, const char *Str, unsigned SLen) {
  return wrap(MDString::get(*unwrap(C), StringRef(Str, SLen)));
}

LLVMMetadataRef LLVMMDNode2(LLVMContextRef C, LLVMMetadataRef *MDs,
                            unsigned Count) {
  return wrap(
      MDNode::get(*unwrap(C), ArrayRef<Metadata *>(unwrap(MDs), Count)));
}

void LLVMAddNamedMetadataOperand2(LLVMModuleRef M, const char *name,
                                  LLVMMetadataRef Val) {
  NamedMDNode *N = unwrap(M)->getOrInsertNamedMetadata(name);
  if (!N)
    return;
  if (!Val)
    return;
  N->addOperand(unwrap<MDNode>(Val));
}

void LLVMSetMetadata2(LLVMValueRef Inst, unsigned KindID, LLVMMetadataRef MD) {
  MDNode *N = MD ? unwrap<MDNode>(MD) : nullptr;
  unwrap<Instruction>(Inst)->setMetadata(KindID, N);
}

void LLVMGoSetCurrentDebugLocation(LLVMBuilderRef Bref, unsigned Line,
                                  unsigned Col, LLVMMetadataRef Scope,
                                  LLVMMetadataRef InlinedAt) {
  if (!Scope)
    unwrap(Bref)->SetCurrentDebugLocation(DebugLoc());
  else
    unwrap(Bref)->SetCurrentDebugLocation(DILocation::get(
        unwrap<MDNode>(Scope)->getContext(), Line, Col, unwrap<MDNode>(Scope),
        InlinedAt ? unwrap<MDNode>(InlinedAt) : nullptr));
}

LLVMDebugLocMetadata LLVMGoGetCurrentDebugLocation(LLVMBuilderRef Bref) {
  const auto& Loc = unwrap(Bref)->getCurrentDebugLocation();
  const auto* InlinedAt = Loc.getInlinedAt();
  const LLVMDebugLocMetadata md{
    Loc.getLine(),
    Loc.getCol(),
    wrap(Loc.getScope()),
    InlinedAt == nullptr ? nullptr : wrap(InlinedAt->getRawInlinedAt()),
  };
  return md;
}

LLVMValueRef LLVMGoGetInlineAsm(LLVMTypeRef Ty, char *AsmString,
                                size_t AsmStringSize, char *Constraints,
                                size_t ConstraintsSize, LLVMBool HasSideEffects,
                                LLVMBool IsAlignStack,
                                LLVMInlineAsmDialect Dialect, LLVMBool CanThrow)
{
  return LLVMGetInlineAsm(Ty, AsmString,
                          AsmStringSize, Constraints,
                          ConstraintsSize, HasSideEffects,
                          IsAlignStack,
                          Dialect, CanThrow);
}

LLVMValueRef LLVMGoBuildIntrinsicCall(LLVMBuilderRef B, LLVMTypeRef RetTy,
                                      unsigned ID, LLVMValueRef *Args,
                                      unsigned Count, const char *Name) {
  Intrinsic::ID IntrinsicID = static_cast<Intrinsic::ID>(ID);
  ArrayRef<Value *> UnwrappedArgs(unwrap(Args), Count);
  SmallVector<Type *, 8> ParamTys;
  SmallVector<Type *, 8> OverloadTys;
  SmallVector<Intrinsic::IITDescriptor, 8> Infos;

  ParamTys.reserve(Count);
  for (Value *Arg : UnwrappedArgs)
    ParamTys.push_back(Arg->getType());

  FunctionType *FTy = FunctionType::get(unwrap(RetTy), ParamTys, false);
  Intrinsic::getIntrinsicInfoTableEntries(IntrinsicID, Infos);
  ArrayRef<Intrinsic::IITDescriptor> InfoRef(Infos);
  if (Intrinsic::matchIntrinsicSignature(FTy, InfoRef, OverloadTys) !=
      Intrinsic::MatchIntrinsicTypes_Match)
    return nullptr;

  return wrap(unwrap(B)->CreateIntrinsic(IntrinsicID, OverloadTys,
                                         UnwrappedArgs, nullptr, Name));
}

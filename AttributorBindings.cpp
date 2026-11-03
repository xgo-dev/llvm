// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
#include "AttributorBindings.h"
#include "llvm/Config/llvm-config.h"

#if LLVM_VERSION_MAJOR >= 22
#include "llvm/IR/Module.h"
#include "llvm/Passes/PassBuilder.h"
#include "llvm/Support/CBindingWrapping.h"
#include "llvm/Target/TargetMachine.h"
#include "llvm/Transforms/IPO/Attributor.h"
#include "llvm/Transforms/Utils/CallGraphUpdater.h"
#endif

LLVMErrorRef LLVMGoRunAttributorHeapToStack(LLVMModuleRef M,
                                            LLVMTargetMachineRef TM) {
#if LLVM_VERSION_MAJOR >= 22
  using namespace llvm;
  if (!M)
    return LLVMCreateStringError("RunAttributorHeapToStack requires a module");
  auto &Mod = *unwrap(M);
  LoopAnalysisManager LAM;
  FunctionAnalysisManager FAM;
  CGSCCAnalysisManager CGAM;
  ModuleAnalysisManager MAM;
  PassBuilder PB(reinterpret_cast<TargetMachine *>(TM));
  PB.registerModuleAnalyses(MAM);
  PB.registerCGSCCAnalyses(CGAM);
  PB.registerFunctionAnalyses(FAM);
  PB.registerLoopAnalyses(LAM);
  PB.crossRegisterProxies(LAM, FAM, CGAM, MAM);

  SetVector<Function *> Functions;
  for (Function &F : Mod)
    if (!F.isDeclaration() && !F.hasOptNone())
      Functions.insert(&F);
  if (Functions.empty())
    return LLVMErrorSuccess;

  AnalysisGetter AG(FAM);
  CallGraphUpdater CGUpdater;
  BumpPtrAllocator Allocator;
  InformationCache InfoCache(Mod, AG, Allocator, nullptr);
  AttributorConfig Config(CGUpdater);
  Config.DeleteFns = false;
  Config.RewriteSignatures = false;
  Config.DefaultInitializeLiveInternals = false;
  Attributor A(Functions, InfoCache, Config);
  // Only seed heap-to-stack and let it request its dependencies. Seeding all
  // default analyses also changes signatures and runs unrelated deductions,
  // including LLVM 22's unsafe invariant-load inference for indirect calls.
  for (Function *F : Functions)
    A.getOrCreateAAFor<AAHeapToStack>(IRPosition::function(*F));
  A.run();
  return LLVMErrorSuccess;
#else
  return LLVMCreateStringError(
      "RunAttributorHeapToStack requires LLVM 22 or newer");
#endif
}

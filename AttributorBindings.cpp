//===- AttributorBindings.cpp - Attributor bindings -----------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

#include "AttributorBindings.h"
#include "llvm/Config/llvm-config.h"

#if LLVM_VERSION_MAJOR >= 22
#include "llvm/IR/Instructions.h"
#include "llvm/IR/Module.h"
#include "llvm/Passes/PassBuilder.h"
#include "llvm/Support/CBindingWrapping.h"
#include "llvm/Target/TargetMachine.h"
#include "llvm/Transforms/IPO/Attributor.h"
#include "llvm/Transforms/Utils/CallGraphUpdater.h"

using namespace llvm;

DEFINE_SIMPLE_CONVERSION_FUNCTIONS(TargetMachine, LLVMTargetMachineRef)

#endif

LLVMErrorRef LLVMGoRunAttributorHeapToStack(LLVMModuleRef M,
                                            LLVMTargetMachineRef TM,
                                            LLVMValueRef *ExcludedCalls,
                                            unsigned NumExcludedCalls) {
#if LLVM_VERSION_MAJOR >= 22
  using namespace llvm;
  if (!M)
    return LLVMCreateStringError("RunAttributorHeapToStack requires a module");
  auto &Mod = *unwrap(M);
  if (NumExcludedCalls && !ExcludedCalls)
    return LLVMCreateStringError("missing excluded call array");
  for (unsigned I = 0; I < NumExcludedCalls; ++I) {
    auto *CB = ExcludedCalls[I] ? dyn_cast<CallBase>(unwrap(ExcludedCalls[I]))
                                : nullptr;
    if (!CB || !CB->getParent() || !CB->getFunction() ||
        CB->getModule() != &Mod)
      return LLVMCreateStringError(
          "excluded values must be calls in this module");
  }
  LoopAnalysisManager LAM;
  FunctionAnalysisManager FAM;
  CGSCCAnalysisManager CGAM;
  ModuleAnalysisManager MAM;
  PassBuilder PB(unwrap(TM));
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

  // Register exclusions before any AA is seeded. Stopping the use walk also
  // prevents it from collecting free calls, so the must-free fallback cannot
  // bypass an exclusion. Other abstract attributes are unaffected.
  for (unsigned I = 0; I < NumExcludedCalls; ++I)
    A.registerVirtualUseCallback(
        *unwrap(ExcludedCalls[I]),
        [](Attributor &, const AbstractAttribute *QueryingAA) {
          return !isa_and_nonnull<AAHeapToStack>(QueryingAA);
        });
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

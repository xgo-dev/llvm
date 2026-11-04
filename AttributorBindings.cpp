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
#include "llvm/ADT/SCCIterator.h"
#include "llvm/ADT/SmallPtrSet.h"
#include "llvm/Analysis/MemoryBuiltins.h"
#include "llvm/IR/CFG.h"
#include "llvm/IR/Instructions.h"
#include "llvm/IR/Module.h"
#include "llvm/Passes/PassBuilder.h"
#include "llvm/Support/CBindingWrapping.h"
#include "llvm/Target/TargetMachine.h"
#include "llvm/Transforms/IPO/Attributor.h"
#include "llvm/Transforms/Utils/CallGraphUpdater.h"

using namespace llvm;

DEFINE_SIMPLE_CONVERSION_FUNCTIONS(TargetMachine, LLVMTargetMachineRef)

// A sufficient condition for reusing storage on the next execution of CB:
// all pointer uses finish in this block, after CB, without carrying the pointer
// through a phi, storing it elsewhere, or observing its address. HeapToStack
// still has to prove that calls do not capture it (or that it is safely freed).
// Keep cross-block uses on the heap until we have a more general scope proof.
static bool canReuseAllocation(CallBase &CB) {
  SmallVector<Value *, 8> Work{&CB};
  SmallPtrSet<Value *, 16> Seen;
  while (!Work.empty()) {
    Value *V = Work.pop_back_val();
    if (!Seen.insert(V).second)
      continue;
    for (Use &U : V->uses()) {
      auto *I = dyn_cast<Instruction>(U.getUser());
      if (!I || I->getParent() != CB.getParent() || !CB.comesBefore(I))
        return false;
      if (isa<GetElementPtrInst>(I) || isa<BitCastInst>(I)) {
        Work.push_back(I);
      } else if (auto *LI = dyn_cast<LoadInst>(I)) {
        if (LI->getPointerOperand() != V)
          return false;
      } else if (auto *SI = dyn_cast<StoreInst>(I)) {
        if (SI->getPointerOperand() != V || SI->getValueOperand() == V)
          return false;
      } else if (auto *Call = dyn_cast<CallBase>(I)) {
        if (!Call->isArgOperand(&U))
          return false;
        if (auto *CI = dyn_cast<CallInst>(Call))
          if (CI->isTailCall())
            return false;
      } else {
        return false;
      }
    }
  }
  return true;
}
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

  SmallPtrSet<AllocaInst *, 32> ExistingAllocas;
  for (Function *F : Functions) {
    for (BasicBlock &BB : *F)
      for (Instruction &I : BB)
        if (auto *AI = dyn_cast<AllocaInst>(&I))
          ExistingAllocas.insert(AI);

    const auto &TLI = FAM.getResult<TargetLibraryAnalysis>(*F);
    // SCCs include irreducible cycles. Only repeated allocations need a
    // storage-reuse proof; other allocations in this function remain eligible.
    for (auto I = scc_begin(F); !I.isAtEnd(); ++I) {
      if (!I.hasCycle())
        continue;
      for (BasicBlock *BB : *I)
        for (Instruction &Inst : *BB) {
          auto *CB = dyn_cast<CallBase>(&Inst);
          if (!CB || !isAllocationFn(CB, &TLI))
            continue;
          if (getAllocSize(CB, &TLI) && canReuseAllocation(*CB))
            continue;
          // Register before seeding any AA. Returning false prevents the use
          // walk from collecting PotentialFreeCalls too, so the must-free
          // fallback cannot bypass this restriction. Other AAs are unaffected.
          A.registerVirtualUseCallback(
              *CB, [](Attributor &, const AbstractAttribute *QueryingAA) {
                return !isa_and_nonnull<AAHeapToStack>(QueryingAA);
              });
        }
    }
  }
  // Only seed heap-to-stack and let it request its dependencies. Seeding all
  // default analyses also changes signatures and runs unrelated deductions,
  // including LLVM 22's unsafe invariant-load inference for indirect calls.
  for (Function *F : Functions)
    A.getOrCreateAAFor<AAHeapToStack>(IRPosition::function(*F));
  A.run();

  // HeapToStack leaves repeated allocations at the call site. The filter
  // above proved reuse safe for those it can convert. Move only its new,
  // constant-size slots; keep initialization at the original call site and
  // leave pre-existing allocas under their owning transform's control.
  for (Function *F : Functions) {
    SmallVector<AllocaInst *, 8> ToHoist;
    for (BasicBlock &BB : *F)
      if (&BB != &F->getEntryBlock())
        for (Instruction &I : BB)
          if (auto *AI = dyn_cast<AllocaInst>(&I))
            if (!ExistingAllocas.contains(AI) &&
                isa<ConstantInt>(AI->getArraySize()))
              ToHoist.push_back(AI);
    for (AllocaInst *AI : ToHoist)
      AI->moveBefore(F->getEntryBlock(),
                     F->getEntryBlock().getFirstInsertionPt());
  }
  return LLVMErrorSuccess;
#else
  return LLVMCreateStringError(
      "RunAttributorHeapToStack requires LLVM 22 or newer");
#endif
}

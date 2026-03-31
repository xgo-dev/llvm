//===- TargetBindings.cpp - Additional bindings for target ---------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

#include "TargetBindings.h"

#include "llvm/Support/CBindingWrapping.h"
#include "llvm/Target/TargetMachine.h"

using namespace llvm;

DEFINE_SIMPLE_CONVERSION_FUNCTIONS(TargetMachine, LLVMTargetMachineRef)

LLVMTargetMachineRef LLVMGoCreateTargetMachineWithOptions(
    LLVMTargetRef T, const char *Triple, const char *CPU,
    const char *Features, LLVMCodeGenOptLevel Level, LLVMRelocMode Reloc,
    LLVMCodeModel CodeModel,
    LLVMBool FunctionSections, LLVMBool DataSections,
    LLVMBool UniqueSectionNames) {
  LLVMTargetMachineRef TM =
      LLVMCreateTargetMachine(T, Triple, CPU, Features, Level, Reloc,
                              CodeModel);
  if (!TM)
    return nullptr;

  TargetMachine *Machine = unwrap(TM);
  Machine->Options.FunctionSections = !!FunctionSections;
  Machine->Options.DataSections = !!DataSections;
  Machine->Options.UniqueSectionNames = !!UniqueSectionNames;
  return TM;
}

LLVMBool LLVMGoTargetMachineFunctionSections(LLVMTargetMachineRef TM) {
  TargetMachine *Machine = unwrap(TM);
  return Machine && Machine->Options.FunctionSections;
}

LLVMBool LLVMGoTargetMachineDataSections(LLVMTargetMachineRef TM) {
  TargetMachine *Machine = unwrap(TM);
  return Machine && Machine->Options.DataSections;
}

LLVMBool LLVMGoTargetMachineUniqueSectionNames(LLVMTargetMachineRef TM) {
  TargetMachine *Machine = unwrap(TM);
  return Machine && Machine->Options.UniqueSectionNames;
}

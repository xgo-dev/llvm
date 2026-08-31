//===- TargetBindings.h - Additional bindings for target --------*- C++ -*-===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

#ifndef LLVM_BINDINGS_GO_LLVM_TARGETBINDINGS_H
#define LLVM_BINDINGS_GO_LLVM_TARGETBINDINGS_H

#include "llvm-c/Target.h"
#include "llvm-c/TargetMachine.h"

#ifdef __cplusplus
extern "C" {
#endif

LLVMTargetMachineRef LLVMGoCreateTargetMachineWithOptions(
    LLVMTargetRef T, const char *Triple, const char *CPU,
    const char *Features, LLVMCodeGenOptLevel Level, LLVMRelocMode Reloc,
    LLVMCodeModel CodeModel,
    LLVMBool FunctionSections, LLVMBool DataSections,
    LLVMBool UniqueSectionNames);
LLVMBool LLVMGoTargetMachineFunctionSections(LLVMTargetMachineRef TM);
LLVMBool LLVMGoTargetMachineDataSections(LLVMTargetMachineRef TM);
LLVMBool LLVMGoTargetMachineUniqueSectionNames(LLVMTargetMachineRef TM);

#ifdef __cplusplus
}
#endif

#endif

//===- AttributorBindings.h - Attributor bindings -----------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

#ifndef LLVM_BINDINGS_GO_ATTRIBUTOR_H
#define LLVM_BINDINGS_GO_ATTRIBUTOR_H
#include "llvm-c/Core.h"
#include "llvm-c/Error.h"
#include "llvm-c/TargetMachine.h"
#ifdef __cplusplus
extern "C" {
#endif
LLVMErrorRef LLVMGoRunAttributorHeapToStack(LLVMModuleRef M,
                                         LLVMTargetMachineRef TM);
#ifdef __cplusplus
}
#endif
#endif

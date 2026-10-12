// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
#ifndef LLVM_BINDINGS_GO_C_API_BINDINGS_H
#define LLVM_BINDINGS_GO_C_API_BINDINGS_H

#include "llvm-c/Core.h"

#ifdef __cplusplus
extern "C" {
#endif

LLVMValueRef LLVMGoGetOrInsertFunction(LLVMModuleRef M, const char *Name,
                                       size_t NameLen, LLVMTypeRef FunctionTy);
LLVMBool LLVMGoParseIRInContext(LLVMContextRef Context, LLVMMemoryBufferRef Buffer,
                               LLVMModuleRef *OutModule, char **OutMessage);

#ifdef __cplusplus
}
#endif
#endif

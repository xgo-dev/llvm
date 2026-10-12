// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
#include "CApiBindings.h"
#include "llvm-c/IRReader.h"
#include "llvm/Config/llvm-config.h"
#include "llvm/IR/Module.h"
#include "llvm/Support/MemoryBuffer.h"

LLVMValueRef LLVMGoGetOrInsertFunction(LLVMModuleRef M, const char *Name,
                                       size_t NameLen, LLVMTypeRef FunctionTy) {
#if LLVM_VERSION_MAJOR >= 22
  return LLVMGetOrInsertFunction(M, Name, NameLen, FunctionTy);
#else
  return llvm::wrap(llvm::unwrap(M)->getOrInsertFunction(
      llvm::StringRef(Name, NameLen), llvm::unwrap<llvm::FunctionType>(FunctionTy))
                        .getCallee());
#endif
}

LLVMBool LLVMGoParseIRInContext(LLVMContextRef Context, LLVMMemoryBufferRef Buffer,
                               LLVMModuleRef *OutModule, char **OutMessage) {
#if LLVM_VERSION_MAJOR >= 22
  return LLVMParseIRInContext2(Context, Buffer, OutModule, OutMessage);
#else
  // The legacy parser consumes its input even on failure. Give it an owned
  // copy so the caller keeps the same ownership contract on older LLVMs.
  LLVMMemoryBufferRef Copy = LLVMCreateMemoryBufferWithMemoryRangeCopy(
      LLVMGetBufferStart(Buffer), LLVMGetBufferSize(Buffer),
      llvm::unwrap(Buffer)->getBufferIdentifier().str().c_str());
  return LLVMParseIRInContext(Context, Copy, OutModule, OutMessage);
#endif
}

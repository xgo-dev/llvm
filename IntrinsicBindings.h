// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
#ifndef LLVM_BINDINGS_GO_INTRINSICBINDINGS_H
#define LLVM_BINDINGS_GO_INTRINSICBINDINGS_H

#include "llvm-c/Core.h"
#include "llvm-c/Target.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
  LLVMTypeRef *types;
  size_t count;
  LLVMBool matched;
} LLVMGoIntrinsicSignature;

LLVMGoIntrinsicSignature LLVMGoMatchIntrinsicSignature(unsigned ID,
                                                       LLVMTypeRef Signature);
void LLVMGoDisposeIntrinsicSignature(LLVMGoIntrinsicSignature Signature);
unsigned LLVMGoIntrinsicCount(void);
LLVMBool LLVMGoIsTargetIntrinsic(unsigned ID);
unsigned LLVMGoIntrinsicImmArgs(unsigned ID, LLVMTypeRef Signature,
                                unsigned *Indices);

typedef struct {
  const char *kind;
  unsigned value, argument, argument_kind, reference;
  LLVMBool scalable;
} LLVMGoIITDescriptor;
size_t LLVMGoIntrinsicDescriptors(unsigned ID, LLVMGoIITDescriptor *Out);

LLVMTypeRef LLVMGoArrayType64(LLVMTypeRef Element, uint64_t Count);
uint64_t LLVMGoArrayLength64(LLVMTypeRef Array);
unsigned LLVMGoScalarBits(LLVMTypeRef Type);
LLVMBool LLVMGoConstIntInRange(LLVMValueRef Value, uint64_t Low, uint64_t High,
                               LLVMBool Unsigned);
LLVMTargetDataRef LLVMGoParseTargetData(const char *Layout, char **Error);
unsigned LLVMGoIndexTypeSizeInBits(LLVMTargetDataRef Layout, LLVMTypeRef Type);

#ifdef __cplusplus
}
#endif
#endif

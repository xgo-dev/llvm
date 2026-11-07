// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
#include "IntrinsicBindings.h"
#include "llvm/ADT/SmallVector.h"
#include "llvm/Config/llvm-config.h"
#include "llvm/IR/Attributes.h"
#include "llvm/IR/Constants.h"
#include "llvm/IR/DataLayout.h"
#include "llvm/IR/DerivedTypes.h"
#include "llvm/IR/Function.h"
#include "llvm/IR/Intrinsics.h"
#include "llvm/IR/Module.h"
#include "llvm/Support/Error.h"
#include <algorithm>

using namespace llvm;

#if LLVM_VERSION_MAJOR < 20
namespace {
#define GET_INTRINSIC_TARGET_DATA
#include "llvm/IR/IntrinsicImpl.inc"
#undef GET_INTRINSIC_TARGET_DATA
} // namespace
#endif

static bool validID(unsigned ID) {
  return ID && ID < Intrinsic::num_intrinsics;
}

LLVMGoIntrinsicSignature LLVMGoMatchIntrinsicSignature(unsigned ID,
                                                       LLVMTypeRef Signature) {
  LLVMGoIntrinsicSignature Result{};
  if (!validID(ID) || !Signature)
    return Result;
  auto *FT = dyn_cast<FunctionType>(unwrap(Signature));
  if (!FT)
    return Result;
  SmallVector<Type *, 8> Types;
#if LLVM_VERSION_MAJOR >= 23
  if (!Intrinsic::isSignatureValid(ID, FT, Types))
    return Result;
#else
  SmallVector<Intrinsic::IITDescriptor, 16> Entries;
  Intrinsic::getIntrinsicInfoTableEntries(ID, Entries);
  ArrayRef<Intrinsic::IITDescriptor> Infos(Entries);
  if (Intrinsic::matchIntrinsicSignature(FT, Infos, Types) !=
          Intrinsic::MatchIntrinsicTypes_Match ||
      Intrinsic::matchIntrinsicVarArg(FT->isVarArg(), Infos))
    return Result;
#endif
  if (Intrinsic::getType(FT->getContext(), ID, Types) != FT)
    return Result;
  Result.count = Types.size();
  Result.matched = true;
  if (!Types.empty()) {
    Result.types = new LLVMTypeRef[Types.size()];
    for (size_t I = 0; I < Types.size(); ++I)
      Result.types[I] = wrap(Types[I]);
  }
  return Result;
}

void LLVMGoDisposeIntrinsicSignature(LLVMGoIntrinsicSignature Signature) {
  delete[] Signature.types;
}

unsigned LLVMGoIntrinsicCount(void) { return Intrinsic::num_intrinsics; }
LLVMBool LLVMGoIsTargetIntrinsic(unsigned ID) {
  if (!validID(ID))
    return false;
#if LLVM_VERSION_MAJOR >= 20
  return Intrinsic::isTargetIntrinsic(ID);
#else
  return ID > TargetInfos[0].Count;
#endif
}

unsigned LLVMGoIntrinsicImmArgs(unsigned ID, LLVMTypeRef Signature,
                                unsigned *Indices) {
  if (!validID(ID) || !Signature)
    return 0;
  auto *FT = dyn_cast<FunctionType>(unwrap(Signature));
  if (!FT)
    return 0;
#if LLVM_VERSION_MAJOR >= 21
  auto Attrs = Intrinsic::getAttributes(FT->getContext(), ID, FT);
#else
  auto Attrs = Intrinsic::getAttributes(FT->getContext(), ID);
#endif
  unsigned Count = 0;
  for (unsigned I = 0; I < FT->getNumParams(); ++I)
    if (Attrs.hasParamAttr(I, Attribute::ImmArg)) {
      if (Indices)
        Indices[Count] = I;
      ++Count;
    }
  return Count;
}

size_t LLVMGoIntrinsicDescriptors(unsigned ID, LLVMGoIITDescriptor *Out) {
  if (!validID(ID))
    return 0;
  SmallVector<Intrinsic::IITDescriptor, 16> Entries;
  Intrinsic::getIntrinsicInfoTableEntries(ID, Entries);
  if (!Out)
    return Entries.size();
  using D = Intrinsic::IITDescriptor;
  for (size_t I = 0; I < Entries.size(); ++I) {
    auto Dsc = Entries[I];
    LLVMGoIITDescriptor R{};
    switch (Dsc.Kind) {
#define KIND(K)                                                                \
  case D::K:                                                                   \
    R.kind = #K;                                                               \
    break
      KIND(Void);
      KIND(VarArg);
      KIND(MMX);
      KIND(Token);
      KIND(Metadata);
      KIND(Half);
      KIND(BFloat);
      KIND(Float);
      KIND(Double);
      KIND(Quad);
      KIND(AMX);
      KIND(PPCQuad);
#if LLVM_VERSION_MAJOR >= 17
      KIND(AArch64Svcount);
#endif
#undef KIND
    case D::Integer:
      R.kind = "Integer";
      R.value = Dsc.Integer_Width;
      break;
    case D::Pointer:
      R.kind = "Pointer";
      R.value = Dsc.Pointer_AddressSpace;
      break;
    case D::Struct:
      R.kind = "Struct";
      R.value = Dsc.Struct_NumElements;
      break;
    case D::Vector:
      R.kind = "Vector";
      R.value = Dsc.Vector_Width.getKnownMinValue();
      R.scalable = Dsc.Vector_Width.isScalable();
      break;
    case D::Argument:
      R.kind = "Argument";
      R.argument = Dsc.getArgumentNumber();
      R.argument_kind = Dsc.getArgumentKind();
      break;
#define ARG(K)                                                                 \
  case D::K:                                                                   \
    R.kind = #K;                                                               \
    R.argument = Dsc.getArgumentNumber();                                      \
    break
      ARG(ExtendArgument);
      ARG(TruncArgument);
      ARG(SameVecWidthArgument);
      ARG(VecElementArgument);
      ARG(Subdivide2Argument);
      ARG(Subdivide4Argument);
      ARG(VecOfBitcastsToInt);
#if LLVM_VERSION_MAJOR < 17
      ARG(PtrToArgument);
      ARG(PtrToElt);
#endif
#undef ARG
#if LLVM_VERSION_MAJOR >= 21
    case D::OneNthEltsVecArgument:
      R.kind = "OneNthEltsVecArgument";
      R.reference = Dsc.getRefArgNumber();
      R.value = Dsc.getVectorDivisor();
      break;
#else
    case D::HalfVecArgument:
      R.kind = "OneNthEltsVecArgument";
      R.reference = Dsc.getArgumentNumber();
      R.value = 2;
      break;
#endif
    case D::VecOfAnyPtrsToElt:
      R.kind = "VecOfAnyPtrsToElt";
      R.argument = Dsc.getOverloadArgNumber();
      R.reference = Dsc.getRefArgNumber();
      break;
#if LLVM_VERSION_MAJOR >= 15 && LLVM_VERSION_MAJOR < 17
    case D::AnyPtrToElt:
      R.kind = "AnyPtrToElt";
      R.argument = Dsc.getOverloadArgNumber();
      R.reference = Dsc.getRefArgNumber();
      break;
#endif
    }
    Out[I] = R;
  }
  return Entries.size();
}

LLVMTypeRef LLVMGoArrayType64(LLVMTypeRef Element, uint64_t Count) {
#if LLVM_VERSION_MAJOR >= 17
  return LLVMArrayType2(Element, Count);
#else
  return wrap(ArrayType::get(unwrap(Element), Count));
#endif
}
uint64_t LLVMGoArrayLength64(LLVMTypeRef Array) {
#if LLVM_VERSION_MAJOR >= 17
  return LLVMGetArrayLength2(Array);
#else
  return cast<ArrayType>(unwrap(Array))->getNumElements();
#endif
}
unsigned LLVMGoScalarBits(LLVMTypeRef Type) {
  return unwrap(Type)->getScalarSizeInBits();
}
LLVMBool LLVMGoConstIntInRange(LLVMValueRef Value, uint64_t Low, uint64_t High,
                               LLVMBool Unsigned) {
  auto *Integer = Value ? dyn_cast<ConstantInt>(unwrap(Value)) : nullptr;
  if (!Integer)
    return false;
  unsigned Width = std::max(64u, Integer->getBitWidth());
  if (Unsigned) {
    auto Actual = Integer->getValue().zextOrTrunc(Width);
    return Actual.uge(APInt(64, Low).zextOrTrunc(Width)) &&
           Actual.ult(APInt(64, High).zextOrTrunc(Width));
  }
  auto Actual = Integer->getValue().sextOrTrunc(Width);
  return Actual.sge(APInt(64, Low).sextOrTrunc(Width)) &&
         Actual.slt(APInt(64, High).sextOrTrunc(Width));
}

LLVMTargetDataRef LLVMGoParseTargetData(const char *Layout, char **Error) {
  auto Parsed = DataLayout::parse(Layout);
  if (!Parsed) {
    *Error = LLVMCreateMessage(toString(Parsed.takeError()).c_str());
    return nullptr;
  }
  return wrap(new DataLayout(std::move(*Parsed)));
}
unsigned LLVMGoIndexTypeSizeInBits(LLVMTargetDataRef Layout, LLVMTypeRef Type) {
  auto *Ty = Type ? unwrap(Type) : nullptr;
  if (!Layout || !Ty || !Ty->isPtrOrPtrVectorTy())
    return 0;
  return unwrap(Layout)->getIndexTypeSizeInBits(Ty);
}

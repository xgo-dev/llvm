
#include "backports.h"
#include "llvm/Analysis/ModuleSummaryAnalysis.h"
#include "llvm/Analysis/ProfileSummaryInfo.h"
#include "llvm/Bitcode/BitcodeWriter.h"
#include "llvm/IR/Instructions.h"
#if LLVM_VERSION_MAJOR >= 16
#include "llvm/IR/PassManager.h"
#include "llvm/Analysis/LoopAnalysisManager.h"
#include "llvm/Analysis/CGSCCPassManager.h"
#include "llvm/Passes/PassBuilder.h"
#include "llvm/Transforms/IPO/ThinLTOBitcodeWriter.h"
#else
#include "llvm/IR/LegacyPassManager.h"
#include "llvm/Transforms/IPO/PassManagerBuilder.h"
#endif
#include "llvm/IR/Module.h"
#include "llvm/Object/Archive.h"
#include "llvm/Object/ArchiveWriter.h"
#include "llvm/Pass.h"
#include "llvm/Support/Error.h"
#if LLVM_VERSION_MAJOR >= 18
#include "llvm/TargetParser/Host.h"
#else
#include "llvm/Support/Host.h"
#endif
#include "llvm/Support/MemoryBuffer.h"
#include "llvm/Support/Path.h"
#if LLVM_VERSION_MAJOR >= 16
#include "llvm/TargetParser/Triple.h"
#else
#include "llvm/ADT/Triple.h"
#endif
#include "llvm/Transforms/IPO.h"
#include "llvm-c/Core.h"
#include "llvm-c/DebugInfo.h"
#include <deque>

struct LLVMGoArchiveWriterOpaque {
  llvm::object::Archive::Kind Kind;
  std::vector<llvm::NewArchiveMember> Members;
  std::deque<std::string> Names;
};

static char *LLVMGoArchiveError(llvm::Error Err) {
  if (!Err)
    return nullptr;
  return LLVMCreateMessage(llvm::toString(std::move(Err)).c_str());
}

LLVMGoArchiveWriterRef LLVMGoCreateArchiveWriter(const char *TargetTriple) {
  auto TripleText = TargetTriple && TargetTriple[0]
                        ? llvm::Triple::normalize(TargetTriple)
                        : llvm::sys::getDefaultTargetTriple();
  llvm::Triple Triple(TripleText);
  auto Kind = llvm::object::Archive::K_GNU;
  if (Triple.isOSDarwin())
    Kind = llvm::object::Archive::K_DARWIN;
  else if (Triple.isOSAIX())
    Kind = llvm::object::Archive::K_AIXBIG;
  else if (Triple.isOSWindows())
    Kind = llvm::object::Archive::K_COFF;
  return new LLVMGoArchiveWriterOpaque{Kind, {}, {}};
}

char *LLVMGoArchiveWriterAddFile(LLVMGoArchiveWriterRef Writer,
                                 const char *Path, const char *Name) {
  if (!Writer)
    return LLVMCreateMessage("archive writer is nil");
  if (!Path || !Path[0])
    return LLVMCreateMessage("archive member path is empty");

  auto MemberOrErr = llvm::NewArchiveMember::getFile(Path, true);
  if (!MemberOrErr)
    return LLVMGoArchiveError(MemberOrErr.takeError());

  Writer->Names.emplace_back(
      Name && Name[0] ? Name
                      : llvm::sys::path::filename(Path).str());
  MemberOrErr->MemberName = Writer->Names.back();
  Writer->Members.push_back(std::move(*MemberOrErr));
  return nullptr;
}

char *LLVMGoArchiveWriterAddMemoryBuffer(LLVMGoArchiveWriterRef Writer,
                                         LLVMMemoryBufferRef Buffer,
                                         const char *Name) {
  if (!Writer)
    return LLVMCreateMessage("archive writer is nil");
  if (!Buffer)
    return LLVMCreateMessage("archive member buffer is nil");
  if (!Name || !Name[0])
    return LLVMCreateMessage("archive member name is empty");

  Writer->Names.emplace_back(Name);
  Writer->Members.emplace_back(llvm::unwrap(Buffer)->getMemBufferRef());
  Writer->Members.back().MemberName = Writer->Names.back();
  return nullptr;
}

char *LLVMGoArchiveWriterWrite(LLVMGoArchiveWriterRef Writer,
                               const char *ArchivePath) {
  if (!Writer)
    return LLVMCreateMessage("archive writer is nil");
  if (!ArchivePath || !ArchivePath[0])
    return LLVMCreateMessage("archive path is empty");
  if (Writer->Members.empty())
    return LLVMCreateMessage("archive has no members");

#if LLVM_VERSION_MAJOR >= 18
  auto WriteSymtab = llvm::SymtabWritingMode::NormalSymtab;
#else
  auto WriteSymtab = true;
#endif
  return LLVMGoArchiveError(llvm::writeArchive(
      ArchivePath, Writer->Members, WriteSymtab, Writer->Kind,
      /*Deterministic=*/true, /*Thin=*/false));
}

void LLVMGoDisposeArchiveWriter(LLVMGoArchiveWriterRef Writer) {
  delete Writer;
}

void LLVMGlobalObjectAddMetadata(LLVMValueRef Global, unsigned KindID, LLVMMetadataRef MD) {
  llvm::MDNode *N = MD ? llvm::unwrap<llvm::MDNode>(MD) : nullptr;
  llvm::GlobalObject *O = llvm::unwrap<llvm::GlobalObject>(Global);
  O->addMetadata(KindID, *N);
}

// See https://reviews.llvm.org/D119431
LLVMMemoryBufferRef LLVMGoWriteThinLTOBitcodeToMemoryBuffer(LLVMModuleRef M) {
  std::string Data;
  llvm::raw_string_ostream OS(Data);
#if LLVM_VERSION_MAJOR >= 16
  llvm::LoopAnalysisManager LAM;
  llvm::FunctionAnalysisManager FAM;
  llvm::CGSCCAnalysisManager CGAM;
  llvm::ModuleAnalysisManager MAM;
  llvm::PassBuilder PB;
  PB.registerModuleAnalyses(MAM);
  PB.registerCGSCCAnalyses(CGAM);
  PB.registerFunctionAnalyses(FAM);
  PB.registerLoopAnalyses(LAM);
  PB.crossRegisterProxies(LAM, FAM, CGAM, MAM);
  llvm::ModulePassManager MPM;
  MPM.addPass(llvm::ThinLTOBitcodeWriterPass(OS, nullptr));
  MPM.run(*llvm::unwrap(M), MAM);
#else
  llvm::legacy::PassManager PM;
  PM.add(createWriteThinLTOBitcodePass(OS));
  PM.run(*llvm::unwrap(M));
#endif
  return llvm::wrap(llvm::MemoryBuffer::getMemBufferCopy(OS.str()).release());
}

LLVMMemoryBufferRef LLVMGoWriteFullLTOBitcodeToMemoryBuffer(
    LLVMModuleRef M, LLVMBool EnableSplitLTOUnit) {
  std::string Data;
  llvm::raw_string_ostream OS(Data);
  llvm::Module *Mod = llvm::unwrap(M);
  Mod->addModuleFlag(llvm::Module::Error, "ThinLTO", uint32_t(0));
  llvm::ProfileSummaryInfo PSI(*Mod);
  llvm::ModuleSummaryIndex Index =
      llvm::buildModuleSummaryIndex(*Mod, nullptr, &PSI);
  if (EnableSplitLTOUnit)
    Index.setEnableSplitLTOUnit();
  llvm::WriteBitcodeToFile(*Mod, OS, false, &Index, false);
  return llvm::wrap(llvm::MemoryBuffer::getMemBufferCopy(OS.str()).release());
}

void LLVMGoDIBuilderInsertDbgValueRecordAtEnd(
    LLVMDIBuilderRef Builder, LLVMValueRef Val, LLVMMetadataRef VarInfo,
    LLVMMetadataRef Expr, LLVMMetadataRef DebugLoc, LLVMBasicBlockRef Block) {
#if LLVM_VERSION_MAJOR >= 19
  // Note: this returns a LLVMDbgRecordRef. Previously, InsertValueAtEnd would
  // return a Value. But since the type changed, and I'd like to keep the API
  // consistent across LLVM versions, I decided to drop the return value.
  LLVMDIBuilderInsertDbgValueRecordAtEnd(Builder, Val, VarInfo, Expr, DebugLoc, Block);
#else
  // Old llvm.dbg.* API.
  LLVMDIBuilderInsertDbgValueAtEnd(Builder, Val, VarInfo, Expr, DebugLoc, Block);
#endif
}

void LLVMGoDIBuilderInsertDbgDeclareRecordAtEnd(
    LLVMDIBuilderRef Builder, LLVMValueRef Val, LLVMMetadataRef VarInfo,
    LLVMMetadataRef Expr, LLVMMetadataRef DebugLoc, LLVMBasicBlockRef Block) {
#if LLVM_VERSION_MAJOR >= 19
  // Note: this returns a LLVMDbgRecordRef. Previously, InsertDeclareAtEnd would
  // return a Declare. But since the type changed, and I'd like to keep the API
  // consistent across LLVM versions, I decided to drop the return value.
  LLVMDIBuilderInsertDeclareRecordAtEnd(Builder, Val, VarInfo, Expr, DebugLoc, Block);
#else
  // Old llvm.dbg.* API.
  LLVMDIBuilderInsertDeclareAtEnd(Builder, Val, VarInfo, Expr, DebugLoc, Block);
#endif
}

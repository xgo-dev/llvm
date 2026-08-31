
#include "llvm-c/DebugInfo.h"
#include "llvm-c/Types.h"

#ifdef __cplusplus
extern "C" {
#endif

void LLVMGlobalObjectAddMetadata(LLVMValueRef objValue, unsigned KindID, LLVMMetadataRef md);

LLVMMemoryBufferRef LLVMGoWriteThinLTOBitcodeToMemoryBuffer(LLVMModuleRef M);

LLVMMemoryBufferRef LLVMGoWriteFullLTOBitcodeToMemoryBuffer(
    LLVMModuleRef M, LLVMBool EnableSplitLTOUnit);

typedef struct LLVMGoArchiveWriterOpaque *LLVMGoArchiveWriterRef;

LLVMGoArchiveWriterRef LLVMGoCreateArchiveWriter(const char *TargetTriple);

char *LLVMGoArchiveWriterAddFile(LLVMGoArchiveWriterRef Writer,
                                 const char *Path, const char *Name);

char *LLVMGoArchiveWriterAddMemoryBuffer(LLVMGoArchiveWriterRef Writer,
                                         LLVMMemoryBufferRef Buffer,
                                         const char *Name);

char *LLVMGoArchiveWriterWrite(LLVMGoArchiveWriterRef Writer,
                               const char *ArchivePath);

void LLVMGoDisposeArchiveWriter(LLVMGoArchiveWriterRef Writer);

void LLVMGoDIBuilderInsertDbgValueRecordAtEnd(
    LLVMDIBuilderRef Builder, LLVMValueRef Val, LLVMMetadataRef VarInfo,
    LLVMMetadataRef Expr, LLVMMetadataRef DebugLoc, LLVMBasicBlockRef Block);

void LLVMGoDIBuilderInsertDbgDeclareRecordAtEnd(
    LLVMDIBuilderRef Builder, LLVMValueRef Val, LLVMMetadataRef VarInfo,
    LLVMMetadataRef Expr, LLVMMetadataRef DebugLoc, LLVMBasicBlockRef Block);

#ifdef __cplusplus
}
#endif /* defined(__cplusplus) */

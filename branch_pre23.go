//go:build !llvm23 && (llvm14 || llvm15 || llvm16 || llvm17 || llvm18 || llvm19 || llvm20 || llvm21 || llvm22)

package llvm

// #include "llvm-c/Core.h"
import "C"

const Br Opcode = C.LLVMBr

func (v Value) IsABranchInst() (rv Value) { rv.C = C.LLVMIsABranchInst(v.C); return }

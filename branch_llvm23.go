//go:build llvm23 || (!llvm14 && !llvm15 && !llvm16 && !llvm17 && !llvm18 && !llvm19 && !llvm20 && !llvm21 && !llvm22)

package llvm

// #include "llvm-c/Core.h"
import "C"

// LLVM 23 gives conditional and unconditional branches distinct opcodes.
const (
	UncondBr Opcode = C.LLVMUncondBr
	CondBr   Opcode = C.LLVMCondBr
)

func (v Value) IsAUncondBrInst() (rv Value) { rv.C = C.LLVMIsAUncondBrInst(v.C); return }
func (v Value) IsACondBrInst() (rv Value)   { rv.C = C.LLVMIsACondBrInst(v.C); return }

// IsABranchInst recognizes either branch kind.
func (v Value) IsABranchInst() Value {
	if branch := v.IsAUncondBrInst(); !branch.IsNil() {
		return branch
	}
	return v.IsACondBrInst()
}

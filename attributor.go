// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

/*
#include "AttributorBindings.h"
*/
import "C"
import "errors"

// RunAttributorHeapToStack runs LLVM's heap-to-stack analysis and its required
// Attributor dependencies. It does not seed unrelated attribute deductions,
// rewrite function signatures, or delete functions. A subsequent optimization
// pipeline can promote/scalarize the resulting stack storage.
//
// The caller supplies allocation contracts, data layout, and target triple on
// the module. A zero TargetMachine selects target-independent analysis.
// Functions marked optnone are not seeded. Functions with CFG cycles are
// excluded to prevent loop-local allocas from exhausting the native stack.
// LLVM 22 or newer is required.
func (mod Module) RunAttributorHeapToStack(tm TargetMachine) error {
	if err := C.LLVMGoRunAttributorHeapToStack(mod.C, tm.C); err != nil {
		message := C.LLVMGetErrorMessage(err)
		defer C.LLVMDisposeErrorMessage(message)
		return errors.New(C.GoString(message))
	}
	return nil
}

// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
package llvm

/*
#include "CApiBindings.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// GetOrInsertFunction returns the callee for name, creating an external function
// declaration with type ft if the name is absent. Existing symbols, types, and
// attributes are preserved. The result can be an alias or, with typed pointers,
// a constant-expression cast rather than a Function; callers should use ft when
// constructing calls and check IsAFunction before accessing function attributes.
func (m Module) GetOrInsertFunction(name string, ft Type) (v Value) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	v.C = C.LLVMGoGetOrInsertFunction(m.C, cname, C.size_t(len(name)), ft.C)
	return
}

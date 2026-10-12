//===- irreader.go - Bindings for irreader --------------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This file defines bindings for the irreader component.
//
//===----------------------------------------------------------------------===//

package llvm

/*
#include "llvm-c/Core.h"
#include "llvm-c/IRReader.h"
#include "CApiBindings.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
)

// ParseIR parses LLVM assembly or bitcode into a new module in this context.
// It consumes buf on both success and failure. The caller must not access or
// dispose buf after calling ParseIR. Use ParseIRBuffer to retain ownership.
func (c *Context) ParseIR(buf MemoryBuffer) (Module, error) {
	var m Module
	var errmsg *C.char
	if C.LLVMParseIRInContext(c.C, buf.C, &m.C, &errmsg) != 0 {
		err := errors.New(C.GoString(errmsg))
		C.LLVMDisposeMessage(errmsg)
		return Module{}, err
	}
	return m, nil
}

// ParseIRBuffer parses LLVM assembly or bitcode into a new module in this
// context without consuming buf, on either success or failure. The caller may
// reuse buf and is responsible for calling buf.Dispose when finished.
func (c *Context) ParseIRBuffer(buf MemoryBuffer) (Module, error) {
	var m Module
	var errmsg *C.char
	if C.LLVMGoParseIRInContext(c.C, buf.C, &m.C, &errmsg) != 0 {
		err := errors.New(C.GoString(errmsg))
		C.LLVMDisposeMessage(errmsg)
		return Module{}, err
	}
	return m, nil
}

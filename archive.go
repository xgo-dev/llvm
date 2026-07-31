//===- archive.go - Bindings for archive writing --------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This file defines bindings for llvm::writeArchive.
//
//===----------------------------------------------------------------------===//

package llvm

/*
#include "llvm-c/Core.h"
#include "backports.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// ArchiveMember is an object or bitcode file to add to an archive.
//
// A memory-buffer member borrows its buffer. The caller must keep the buffer
// alive until WriteArchive returns.
type ArchiveMember struct {
	name   string
	path   string
	buffer MemoryBuffer
}

// NewArchiveMemberFromFile creates a path-backed archive member. The archive
// member name is the base name of path.
func NewArchiveMemberFromFile(path string) ArchiveMember {
	return ArchiveMember{path: path}
}

// NewArchiveMemberFromMemoryBuffer creates a memory-backed archive member.
// name is the name stored in the archive.
func NewArchiveMemberFromMemoryBuffer(name string, buffer MemoryBuffer) ArchiveMember {
	return ArchiveMember{name: name, buffer: buffer}
}

// WriteArchive creates a deterministic, non-thin archive with a symbol table.
// targetTriple determines the platform archive format.
func WriteArchive(path, targetTriple string, members []ArchiveMember) error {
	if path == "" {
		return errors.New("archive path is empty")
	}
	if len(members) == 0 {
		return errors.New("archive has no members")
	}

	cTriple := C.CString(targetTriple)
	defer C.free(unsafe.Pointer(cTriple))
	writer := C.LLVMGoCreateArchiveWriter(cTriple)
	if writer == nil {
		return errors.New("failed to create archive writer")
	}
	defer C.LLVMGoDisposeArchiveWriter(writer)

	for i, member := range members {
		var cErr *C.char
		switch {
		case member.path != "":
			cPath := C.CString(member.path)
			var cName *C.char
			if member.name != "" {
				cName = C.CString(member.name)
			}
			cErr = C.LLVMGoArchiveWriterAddFile(writer, cPath, cName)
			C.free(unsafe.Pointer(cPath))
			C.free(unsafe.Pointer(cName))
		case !member.buffer.IsNil():
			if member.name == "" {
				return fmt.Errorf("archive member %d has an empty name", i)
			}
			cName := C.CString(member.name)
			cErr = C.LLVMGoArchiveWriterAddMemoryBuffer(writer, member.buffer.C, cName)
			C.free(unsafe.Pointer(cName))
		default:
			return fmt.Errorf("archive member %d has no file or memory buffer", i)
		}
		if cErr != nil {
			err := errors.New(C.GoString(cErr))
			C.LLVMDisposeMessage(cErr)
			label := member.name
			if member.path != "" {
				label = member.path
			}
			return fmt.Errorf("add archive member %d (%q): %w", i, label, err)
		}
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if cErr := C.LLVMGoArchiveWriterWrite(writer, cPath); cErr != nil {
		err := errors.New(C.GoString(cErr))
		C.LLVMDisposeMessage(cErr)
		return err
	}
	return nil
}

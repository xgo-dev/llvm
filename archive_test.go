//===- archive_test.go - Tests for archive writing ------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

package llvm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteArchiveMemoryAndFileMembers(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()

	mod := ctx.NewModule("archive")
	defer mod.Dispose()
	mod.SetTarget("x86_64-unknown-linux-gnu")
	AddFunction(mod, "answer", FunctionType(ctx.Int32Type(), nil, false))

	memoryBuf := WriteBitcodeToMemoryBuffer(mod)
	defer memoryBuf.Dispose()

	fileBuf := WriteBitcodeToMemoryBuffer(mod)
	defer fileBuf.Dispose()
	filePath := filepath.Join(t.TempDir(), "from-file.bc")
	if err := os.WriteFile(filePath, fileBuf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(t.TempDir(), "libarchive.a")
	err := WriteArchive(archivePath, mod.Target(), []ArchiveMember{
		NewArchiveMemberFromMemoryBuffer("from-memory.bc", memoryBuf),
		NewArchiveMemberFromFile(filePath),
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("!<arch>\n")) {
		end := len(data)
		if end > 8 {
			end = 8
		}
		t.Fatalf("archive magic = %q", data[:end])
	}
	for _, name := range []string{"from-memory.bc", "from-file.bc"} {
		if !bytes.Contains(data, []byte(name)) {
			t.Errorf("archive does not contain member name %q", name)
		}
	}
}

func TestWriteArchiveRejectsInvalidMembers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.a")
	missing := filepath.Join(t.TempDir(), "missing.o")
	tests := []struct {
		name    string
		path    string
		members []ArchiveMember
		want    string
	}{
		{name: "empty path", members: []ArchiveMember{{}}, want: "archive path is empty"},
		{name: "no members", path: path, want: "archive has no members"},
		{name: "empty member", path: path, members: []ArchiveMember{{}}, want: "has no file or memory buffer"},
		{name: "missing file", path: path, members: []ArchiveMember{NewArchiveMemberFromFile(missing)}, want: fmt.Sprintf("%q", missing)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteArchive(tt.path, "x86_64-unknown-linux-gnu", tt.members)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("WriteArchive error = %v, want %q", err, tt.want)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
package llvm

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseIRBufferOwnership(t *testing.T) {
	for _, format := range []string{"assembly", "bitcode", "invalid"} {
		t.Run(format, func(t *testing.T) {
			ctx := NewContext()
			defer ctx.Dispose()
			var buf MemoryBuffer
			var inputPath string
			if format == "bitcode" {
				m := ctx.NewModule("source")
				AddFunction(m, "target", FunctionType(ctx.VoidType(), nil, false))
				buf = WriteBitcodeToMemoryBuffer(m)
				m.Dispose()
			} else {
				src := "declare void @target()\n"
				if format == "invalid" {
					src = "define void @broken( {\n"
				}
				path := filepath.Join(t.TempDir(), "input.ll")
				inputPath = path
				if err := os.WriteFile(path, []byte(src), 0600); err != nil {
					t.Fatal(err)
				}
				var err error
				buf, err = NewMemoryBufferFromFile(path)
				if err != nil {
					t.Fatal(err)
				}
			}
			defer buf.Dispose()
			want := buf.Bytes()
			// Parse twice, with an explicit read after each module is disposed.
			// This covers ownership on success and on the diagnostic path.
			for i := 0; i < 2; i++ {
				m, err := ctx.ParseIRBuffer(buf)
				if format == "invalid" {
					if err == nil || err.Error() == "" || !m.IsNil() {
						t.Fatalf("invalid IR: error=%v", err)
					}
					if !strings.Contains(err.Error(), inputPath) {
						t.Fatalf("diagnostic lost buffer identifier: %v", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if m.NamedFunction("target").IsNil() {
						t.Fatal("parsed module is missing target")
					}
					if err := VerifyModule(m, ReturnStatusAction); err != nil {
						t.Fatal(err)
					}
					if format == "assembly" && !strings.Contains(m.String(), `source_filename = "`+inputPath+`"`) {
						t.Fatalf("module lost buffer identifier: %s", m)
					}
					m.Dispose()
				}
				if !bytes.Equal(buf.Bytes(), want) {
					t.Fatal("parser changed the caller's buffer")
				}
			}
		})
	}
}

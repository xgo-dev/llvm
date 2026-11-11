package llvm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFullLTOBitcodeToMemoryBufferAddsFullLTOFlag(t *testing.T) {
	ctx := NewContext()
	mod := ctx.NewModule("full_lto")
	defer mod.Dispose()

	fnTy := FunctionType(ctx.Int32Type(), nil, false)
	AddFunction(mod, "main", fnTy)

	buf := WriteFullLTOBitcodeToMemoryBuffer(mod, true)
	defer buf.Dispose()
	if buf.IsNil() {
		t.Fatal("WriteFullLTOBitcodeToMemoryBuffer returned nil buffer")
	}

	ir := mod.String()
	if !strings.Contains(ir, `i32 1, !"ThinLTO", i32 0`) {
		t.Fatalf("missing full LTO ThinLTO module flag:\n%s", ir)
	}
	if strings.Contains(ir, `"EnableSplitLTOUnit"`) {
		t.Fatalf("split-unit state should be controlled by the summary index, not a module flag:\n%s", ir)
	}
}

// LTO summaries must also cover definitions added after the optimizer ran.
func TestLTOBitcodeLateDefinitions(t *testing.T) {
	for _, mode := range []string{"thin", "full"} {
		t.Run(mode, func(t *testing.T) {
			ctx := NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("late-definitions")
			defer mod.Dispose()
			b := ctx.NewBuilder()
			defer b.Dispose()
			add := func(name string) {
				fn := AddFunction(mod, name, FunctionType(ctx.Int32Type(), nil, false))
				b.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
				b.CreateRet(ConstInt(ctx.Int32Type(), 7, false))
			}
			add("early")
			opts := NewPassBuilderOptions()
			defer opts.Dispose()
			if err := mod.RunPasses("default<O2>", TargetMachine{}, opts); err != nil {
				t.Fatal(err)
			}
			add("late")
			var buf MemoryBuffer
			if mode == "thin" {
				buf = WriteThinLTOBitcodeToMemoryBuffer(mod)
			} else {
				buf = WriteFullLTOBitcodeToMemoryBuffer(mod, false)
			}
			defer buf.Dispose()
			path := filepath.Join(t.TempDir(), "lto.bc")
			if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			parsed, err := ctx.ParseBitcodeFile(path)
			if err != nil {
				t.Fatal(err)
			}
			defer parsed.Dispose()
			if parsed.NamedFunction("late").IsDeclaration() {
				t.Fatal("lost late definition")
			}
			if err := VerifyModule(parsed, ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

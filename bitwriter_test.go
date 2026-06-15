package llvm

import (
	"strings"
	"testing"
)

func TestWriteFullLTOBitcodeToMemoryBufferAddsFullLTOFlag(t *testing.T) {
	ctx := NewContext()
	mod := ctx.NewModule("full_lto")
	defer mod.Dispose()

	fnTy := FunctionType(ctx.Int32Type(), nil, false)
	AddFunction(mod, "main", fnTy)

	buf := WriteFullLTOBitcodeToMemoryBuffer(mod)
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

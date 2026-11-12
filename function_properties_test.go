// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFunctionProperties(t *testing.T) {
	const source = `
$source = comdat any
declare i32 @personality(...)
define linkonce_odr hidden fastcc i32 @source(ptr nonnull %p) local_unnamed_addr nounwind section "example" comdat align 64 gc "shadow-stack" prefix i32 123 prologue i32 456 personality ptr @personality {
  ret i32 0
}
define internal void @destination(ptr noalias %out, i32 %value) noinline {
  ret void
}
define void @empty() {
  ret void
}
`
	ctx := NewContext()
	defer ctx.Dispose()
	ir := source
	personality := "personality ptr @personality"
	if strings.HasPrefix(Version, "14.") {
		// LLVM 14 still defaults to typed pointers; retain the same fixture
		// and property checks rather than skipping the supported version.
		ir = strings.ReplaceAll(ir, "ptr", "i8*")
		personality = "personality i32 (...)* @personality"
		ir = strings.ReplaceAll(ir, "personality i8* @personality", personality)
	}
	path := filepath.Join(t.TempDir(), "properties.ll")
	if err := os.WriteFile(path, []byte(ir), 0600); err != nil {
		t.Fatal(err)
	}
	buf, err := NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	from, to := mod.NamedFunction("source"), mod.NamedFunction("destination")
	if !to.CopyFunctionProperties(from) {
		t.Fatal("valid functions rejected")
	}
	for _, property := range []string{"linkonce_odr hidden fastcc", "local_unnamed_addr", `section "example"`, "comdat($source)", "align 64", `gc "shadow-stack"`, "prefix i32 123", "prologue i32 456", personality} {
		if !strings.Contains(to.String(), property) {
			t.Errorf("lost %s:\n%s", property, to.String())
		}
	}
	if to.GlobalValueType().ParamTypesCount() != 2 || to.GetEnumAttributeAtIndex(1, AttributeKindID("noalias")).IsNil() || !to.GetEnumAttributeAtIndex(1, AttributeKindID("nonnull")).IsNil() || to.GetEnumFunctionAttribute(AttributeKindID("noinline")).IsNil() || !to.GetEnumFunctionAttribute(AttributeKindID("nounwind")).IsNil() {
		t.Fatalf("destination signature or attribute list changed:\n%s", to.String())
	}
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	empty := mod.NamedFunction("empty")
	empty.SetLinkage(InternalLinkage)
	if !to.CopyFunctionProperties(empty) || to.Linkage() != InternalLinkage {
		t.Fatal("copy to internal linkage failed")
	}
	for _, property := range []string{"linkonce_odr", "hidden", "fastcc", "unnamed_addr", "section", "comdat", "align 64", "gc ", "prefix", "prologue", "personality"} {
		if strings.Contains(to.String(), property) {
			t.Errorf("did not clear %s:\n%s", property, to.String())
		}
	}
	if to.GetEnumFunctionAttribute(AttributeKindID("noinline")).IsNil() {
		t.Fatal("clearing properties removed destination attributes")
	}
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}

func TestCopyFunctionPropertiesRejectsInvalidInputs(t *testing.T) {
	ctx, otherCtx := NewContext(), NewContext()
	defer ctx.Dispose()
	defer otherCtx.Dispose()
	mod, otherMod := ctx.NewModule("local"), ctx.NewModule("other")
	defer mod.Dispose()
	defer otherMod.Dispose()
	foreignMod := otherCtx.NewModule("foreign")
	defer foreignMod.Dispose()
	fn := AddFunction(mod, "local", FunctionType(ctx.VoidType(), nil, false))
	other := AddFunction(otherMod, "other", fn.GlobalValueType())
	foreign := AddFunction(foreignMod, "foreign", FunctionType(otherCtx.VoidType(), nil, false))
	global := AddGlobal(mod, ctx.Int8Type(), "data")
	for _, tc := range []struct {
		name     string
		to, from Value
	}{
		{"nil destination", Value{}, fn}, {"nil source", fn, Value{}},
		{"global destination", global, fn}, {"global source", fn, global},
		{"other module", fn, other}, {"other context", fn, foreign},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := mod.String()
			if tc.to.CopyFunctionProperties(tc.from) {
				t.Error("invalid inputs accepted")
			}
			if mod.String() != before {
				t.Error("rejected copy modified the destination module")
			}
		})
	}
}

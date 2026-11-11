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
define void @destination(ptr noalias %out, i32 %value) noinline {
  ret void
}
define void @empty() {
  ret void
}
`
	ctx := NewContext()
	defer ctx.Dispose()
	path := filepath.Join(t.TempDir(), "properties.ll")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
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
	to.CopyFunctionProperties(from)
	for _, property := range []string{"linkonce_odr hidden fastcc", "local_unnamed_addr", `section "example"`, "comdat($source)", "align 64", `gc "shadow-stack"`, "prefix i32 123", "prologue i32 456", "personality ptr @personality"} {
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
	to.CopyFunctionProperties(mod.NamedFunction("empty"))
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

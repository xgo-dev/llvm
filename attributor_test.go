package llvm

import (
	"strconv"
	"strings"
	"testing"
)

func TestAttributorHeapToStack(t *testing.T) {
	major, _ := strconv.Atoi(strings.SplitN(Version, ".", 2)[0])
	if major < 22 {
		ctx := NewContext()
		defer ctx.Dispose()
		mod := ctx.NewModule("unsupported")
		defer mod.Dispose()
		if err := mod.RunAttributorHeapToStack(TargetMachine{}); err == nil {
			t.Fatal("expected an unsupported-version error")
		}
		return
	}
	if err := (Module{}).RunAttributorHeapToStack(TargetMachine{}); err == nil {
		t.Fatal("nil module must return an error")
	}
	InitializeNativeTarget()
	target, err := GetTargetFromTriple(DefaultTargetTriple())
	if err != nil {
		t.Fatal(err)
	}
	tm := target.CreateTargetMachine(DefaultTargetTriple(), "", "", CodeGenLevelDefault, RelocDefault, CodeModelDefault)
	defer tm.Dispose()
	mod := transformationModule(t, `
declare nonnull ptr @alloc(i64) allocsize(0)
declare nonnull ptr @zero(i64) allocsize(0)
declare void @capture(ptr)

define i64 @local(i64 %x) {
  %p = call ptr @alloc(i64 16) allockind("alloc,uninitialized")
  store i64 %x, ptr %p, align 8
  %v = load i64, ptr %p, align 8
  ret i64 %v
}
define i64 @zeroed() {
  %p = call ptr @zero(i64 16) allockind("alloc,zeroed")
  %v = load i64, ptr %p, align 8
  ret i64 %v
}
define ptr @escape() {
  %p = call ptr @alloc(i64 16) allockind("alloc,uninitialized")
  ret ptr %p
}
define void @captured() {
  %p = call ptr @alloc(i64 16) allockind("alloc,uninitialized")
  call void @capture(ptr %p)
  ret void
}
define ptr @dynamic(i64 %size) {
  %p = call ptr @alloc(i64 %size)
  ret ptr %p
}
define i64 @unoptimized(i64 %x) noinline optnone {
  %p = call ptr @alloc(i64 16) allockind("alloc,uninitialized")
  store i64 %x, ptr %p, align 8
  %v = load i64, ptr %p, align 8
  ret i64 %v
}
; Full Attributor in LLVM 22 crashes in invariant-load inference on this
; indirect-call result. Heap-to-stack does not require that analysis.
define i64 @indirect(ptr %fn) {
  %p = call ptr %fn()
  %v = load i64, ptr %p, align 8
  ret i64 %v
}
`)
	mod.SetTarget(DefaultTargetTriple())
	td := tm.CreateTargetData()
	defer td.Dispose()
	mod.SetDataLayout(td.String())
	if err := mod.RunAttributorHeapToStack(tm); err != nil {
		t.Fatal(err)
	}
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"local", "zeroed"} {
		body := mod.NamedFunction(name).String()
		if !strings.Contains(body, "alloca ") || strings.Contains(body, "call ptr @") {
			t.Fatalf("%s was not moved to stack:\n%s", name, body)
		}
	}
	if body := mod.NamedFunction("zeroed").String(); !strings.Contains(body, "@llvm.memset") {
		t.Fatalf("zero initialization lost:\n%s", body)
	}
	for _, name := range []string{"escape", "captured", "dynamic", "unoptimized"} {
		if body := mod.NamedFunction(name).String(); !strings.Contains(body, "@alloc(") {
			t.Fatalf("%s lost a required allocation:\n%s", name, body)
		}
	}
	opts := NewPassBuilderOptions()
	defer opts.Dispose()
	if err := mod.RunPasses("default<O2>", tm, opts); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"local", "zeroed"} {
		body := mod.NamedFunction(name).String()
		if strings.Contains(body, "alloca ") || strings.Contains(body, "call ") {
			t.Fatalf("O2 retained local storage:\n%s", body)
		}
	}
	if body := mod.NamedFunction("zeroed").String(); !strings.Contains(body, "ret i64 0") {
		t.Fatalf("O2 changed the zero value:\n%s", body)
	}
}

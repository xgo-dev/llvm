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
; Atomic accesses keep the promoted slot alive through O2. A loop-local
; alloca would consume additional stack on every iteration.
declare ptr @malloc(i64)
define i64 @loop(i64 %n) {
entry:
  br label %body
body:
  %i = phi i64 [0, %entry], [%next, %body]
  %p = call ptr @alloc(i64 16) allockind("alloc,uninitialized")
  store atomic i64 %i, ptr %p seq_cst, align 8
  %v = load atomic i64, ptr %p seq_cst, align 8
  %next = add i64 %i, 1
  %more = icmp ult i64 %next, %n
  br i1 %more, label %body, label %exit
exit:
  ret i64 %v
}
; Multiple entries make this cycle irreducible, so LoopInfo alone is
; insufficient. Exercise LLVM's standard allocator recognition as well.
define i64 @irreducible(i1 %first, i1 %again) {
entry:
  br i1 %first, label %left, label %right
left:
  %p = call ptr @malloc(i64 16)
  store atomic i64 42, ptr %p seq_cst, align 8
  %v = load atomic i64, ptr %p seq_cst, align 8
  br i1 %again, label %right, label %exit
right:
  br label %left
exit:
  ret i64 %v
}
define i64 @loop_caller(i64 %n) {
  %v = call i64 @loop(i64 %n)
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
	checkCycles := func() {
		for _, name := range []string{"loop", "irreducible"} {
			body := mod.NamedFunction(name).String()
			if strings.Contains(body, "alloca ") || (!strings.Contains(body, "@alloc(") && !strings.Contains(body, "@malloc(")) {
				t.Fatalf("%s must retain heap storage, not allocate stack per iteration:\n%s", name, body)
			}
		}
	}
	checkCycles()
	opts := NewPassBuilderOptions()
	defer opts.Dispose()
	if err := mod.RunPasses("default<O2>", tm, opts); err != nil {
		t.Fatal(err)
	}
	checkCycles()
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

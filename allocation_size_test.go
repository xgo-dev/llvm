package llvm

import (
	"strconv"
	"strings"
	"testing"
)

func TestConstantAllocationSize(t *testing.T) {
	source := `
declare ptr @alloc(i64) allocsize(0)
declare ptr @pair(i64, i64) allocsize(0, 1)
declare ptr @unknown(i64)
define void @test(i64 %n) {
  %fixed = call ptr @alloc(i64 16)
  %dynamic = call ptr @alloc(i64 %n)
  %product = call ptr @pair(i64 3, i64 7)
  %overflow = call ptr @pair(i64 -1, i64 2)
  %zero = call ptr @alloc(i64 0)
  %unclassified = call ptr @unknown(i64 16)
  ret void
}
`
	major, _ := strconv.Atoi(strings.SplitN(Version, ".", 2)[0])
	if major < 15 {
		source = strings.ReplaceAll(source, "ptr", "i8*")
	}
	mod := transformationModule(t, source)
	mod.SetTarget("x86_64-unknown-linux-gnu")
	mod.SetDataLayout("e-p:64:64")
	want := []struct {
		size  uint64
		known bool
	}{{16, true}, {0, false}, {21, true}, {0, false}, {0, true}, {0, false}, {0, false}}
	i := mod.NamedFunction("test").EntryBasicBlock().FirstInstruction()
	for _, w := range want {
		size, known := i.ConstantAllocationSize()
		if size != w.size || known != w.known {
			t.Fatalf("%s: got (%d,%v), want (%d,%v)", i.String(), size, known, w.size, w.known)
		}
		i = NextInstruction(i)
	}
	if _, known := (Value{}).ConstantAllocationSize(); known {
		t.Fatal("nil value has known allocation size")
	}
	if _, known := mod.NamedFunction("test").ConstantAllocationSize(); known {
		t.Fatal("function is not an allocation call")
	}
}

func TestCallArgumentUses(t *testing.T) {
	source := `
declare void @callee(ptr)
define void @test(ptr %p) {
  call void @callee(ptr %p) ["test"(ptr %p)]
  ret void
}
`
	major, _ := strconv.Atoi(strings.SplitN(Version, ".", 2)[0])
	if major < 15 {
		source = strings.ReplaceAll(source, "ptr", "i8*")
	}
	mod := transformationModule(t, source)
	call := mod.NamedFunction("test").EntryBasicBlock().FirstInstruction()
	if call.ArgOperandsCount() != 1 {
		t.Fatal("operand bundle counted as argument")
	}
	var arguments, bundles int
	for u := mod.NamedFunction("test").Param(0).FirstUse(); !u.IsNil(); u = u.NextUse() {
		if u == call.OperandUse(0) {
			arguments++
		} else {
			bundles++
		}
	}
	if arguments != 1 || bundles != 1 {
		t.Fatalf("argument/bundle uses: %d/%d", arguments, bundles)
	}
}

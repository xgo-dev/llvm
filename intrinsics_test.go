// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

import (
	"reflect"
	"testing"
)

func TestIntrinsicSignatureAndDeclarations(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("intrinsics")
	defer mod.Dispose()
	i32 := ctx.Int32Type()
	id := LookupIntrinsicID("llvm.ctpop")
	ft := FunctionType(i32, []Type{i32}, false)
	overloads, ok := MatchIntrinsicSignature(id, ft)
	if !ok || !reflect.DeepEqual(overloads, []Type{i32}) || ctx.IntrinsicType(id, overloads) != ft {
		t.Fatalf("signature: %v %v", overloads, ok)
	}
	if !mod.FirstFunction().IsNil() {
		t.Fatal("matching created a declaration")
	}
	if got := mod.IntrinsicName(id, overloads); got != "llvm.ctpop.i32" {
		t.Fatalf("name: %s", got)
	}
	decl := mod.IntrinsicDeclaration(id, overloads)
	if decl != mod.IntrinsicDeclaration(id, overloads) || decl.GlobalValueType() != ft || mod.NamedValue(decl.Name()) != decl {
		t.Fatal("declaration was not canonical or reused")
	}
	global := AddGlobal(mod, i32, "global")
	if mod.NamedValue("global") != global || !mod.NamedValue("missing").IsNil() {
		t.Fatal("global lookup")
	}
	caller := AddFunction(mod, "caller", FunctionType(i32, []Type{i32}, false))
	b := ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(ctx.AddBasicBlock(caller, "entry"))
	b.CreateRet(b.CreateCall(ft, decl, []Value{caller.Param(0)}, ""))
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []Type{Type{}, i32, FunctionType(i32, []Type{ctx.Int64Type()}, false), FunctionType(i32, []Type{i32}, true)} {
		if _, ok := MatchIntrinsicSignature(id, bad); ok {
			t.Fatalf("accepted %s", bad)
		}
	}
	ctlz := LookupIntrinsicID("llvm.ctlz")
	ctlzType := FunctionType(i32, []Type{i32, ctx.Int1Type()}, false)
	if got := IntrinsicImmediateArguments(ctlz, ctlzType); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("immarg: %v", got)
	}
	for _, invalid := range []int{-1, 0, IntrinsicCount(), int(^uint(0) >> 1)} {
		if IntrinsicName(invalid) != "" || IntrinsicIsOverloaded(invalid) || IntrinsicIsTarget(invalid) ||
			IntrinsicDescriptors(invalid) != nil || IntrinsicImmediateArguments(invalid, ft) != nil ||
			!ctx.IntrinsicType(invalid, nil).IsNil() || !mod.IntrinsicDeclaration(invalid, nil).IsNil() {
			t.Fatalf("accepted invalid ID %d", invalid)
		}
		if _, ok := MatchIntrinsicSignature(invalid, ft); ok {
			t.Fatalf("matched invalid ID %d", invalid)
		}
	}
}

func TestIntrinsicAggregateAndPointerOverloads(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("overloads")
	defer mod.Dispose()
	ptr0, ptr1 := PointerType(ctx.Int8Type(), 0), PointerType(ctx.Int8Type(), 1)
	ft := FunctionType(ctx.VoidType(), []Type{ptr1, ptr0, ctx.Int64Type(), ctx.Int1Type()}, false)
	id := LookupIntrinsicID("llvm.memcpy")
	overloads, ok := MatchIntrinsicSignature(id, ft)
	if !ok || !reflect.DeepEqual(overloads, []Type{ptr1, ptr0, ctx.Int64Type()}) {
		t.Fatalf("pointer overloads: %v %v", overloads, ok)
	}
	if ctx.IntrinsicType(id, overloads) != ft || mod.IntrinsicDeclaration(id, overloads).GlobalValueType() != ft {
		t.Fatal("memory signature changed")
	}
	// Unnamed identified structures receive a module-local overload suffix.
	a, b := ctx.StructCreateNamed(""), ctx.StructCreateNamed("")
	a.StructSetBody([]Type{ctx.Int32Type()}, false)
	b.StructSetBody([]Type{ctx.Int64Type()}, false)
	copyID := LookupIntrinsicID("llvm.ssa.copy")
	other := ctx.NewModule("other")
	defer other.Dispose()
	if mod.IntrinsicName(copyID, []Type{a}) == mod.IntrinsicName(copyID, []Type{b}) {
		t.Fatal("distinct unnamed overloads have the same name")
	}
	for _, typ := range []Type{a, b, ArrayType(ctx.Int8Type(), 3)} {
		sig := FunctionType(typ, []Type{typ}, false)
		overloads, ok := MatchIntrinsicSignature(copyID, sig)
		if !ok || !reflect.DeepEqual(overloads, []Type{typ}) {
			t.Fatal("aggregate matching")
		}
		for _, m := range []Module{mod, other} {
			decl := m.IntrinsicDeclaration(copyID, overloads)
			if decl.GlobalValueType() != sig || decl.Name() != m.IntrinsicName(copyID, overloads) {
				t.Fatal("module-local overload declaration")
			}
		}
	}
	for _, m := range []Module{mod, other} {
		if err := VerifyModule(m, ReturnStatusAction); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIntrinsicVariadicSignature(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("varargs")
	defer mod.Dispose()
	id := LookupIntrinsicID("llvm.experimental.stackmap")
	params := []Type{ctx.Int64Type(), ctx.Int32Type()}
	ft := FunctionType(ctx.VoidType(), params, true)
	types, ok := MatchIntrinsicSignature(id, ft)
	if !ok || len(types) != 0 {
		t.Fatal("stackmap signature did not match")
	}
	if _, ok := MatchIntrinsicSignature(id, FunctionType(ctx.VoidType(), params, false)); ok {
		t.Fatal("accepted missing ellipsis")
	}
	fn := AddFunction(mod, "caller", FunctionType(ctx.VoidType(), nil, false))
	b := ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
	b.CreateCall(ft, mod.IntrinsicDeclaration(id, types), []Value{
		ConstInt(params[0], 1, false), ConstInt(params[1], 0, false), ConstInt(ctx.Int8Type(), 7, false),
	}, "")
	b.CreateRetVoid()
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}

func TestIntrinsicMetadata(t *testing.T) {
	generic, target := 0, 0
	for id := 1; id < IntrinsicCount(); id++ {
		name := IntrinsicName(id)
		if name == "" || LookupIntrinsicID(name) != id {
			t.Fatalf("ID %d name %q", id, name)
		}
		if IntrinsicIsTarget(id) {
			target++
		} else {
			generic++
		}
		for _, d := range IntrinsicDescriptors(id) {
			if d.Kind == "" {
				t.Fatalf("unhandled descriptor for %s", name)
			}
		}
	}
	if generic == 0 || target == 0 || IntrinsicIsTarget(LookupIntrinsicID("llvm.ctpop")) || !IntrinsicIsTarget(LookupIntrinsicID("llvm.x86.sse2.pause")) {
		t.Fatalf("generic=%d target=%d", generic, target)
	}
	got := IntrinsicDescriptors(LookupIntrinsicID("llvm.ctpop"))
	want := []IITDescriptor{{Kind: "Argument", ArgumentKind: 1}, {Kind: "Argument", ArgumentKind: 7}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ctpop descriptors: %+v", got)
	}
}

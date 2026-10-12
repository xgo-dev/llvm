// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
package llvm

import "testing"

func TestGetOrInsertFunction(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	m := ctx.NewModule("callee")
	defer m.Dispose()
	ft := FunctionType(ctx.VoidType(), nil, false)
	fn := m.GetOrInsertFunction("target", ft)
	if fn.IsAFunction().IsNil() || fn.GlobalValueType() != ft || fn.Linkage() != ExternalLinkage {
		t.Fatalf("unexpected declaration: %s", fn)
	}
	attr := ctx.CreateEnumAttribute(AttributeKindID("nounwind"), 0)
	fn.AddFunctionAttr(attr)
	if got := m.GetOrInsertFunction("target", ft); got != fn || got.GetEnumFunctionAttribute(AttributeKindID("nounwind")) != attr {
		t.Fatal("existing function or attributes were not preserved")
	}
	// Reusing a name with a different function type must preserve the original
	// declaration and return a callable value, without creating target.1.
	otherType := FunctionType(ctx.VoidType(), []Type{ctx.Int32Type()}, false)
	callee := m.GetOrInsertFunction("target", otherType)
	if m.NamedFunction("target") != fn || fn.GlobalValueType() != ft || !m.NamedFunction("target.1").IsNil() {
		t.Fatal("type mismatch changed or duplicated the existing declaration")
	}
	caller := AddFunction(m, "caller", ft)
	b := ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(ctx.AddBasicBlock(caller, "entry"))
	b.CreateCall(otherType, callee, []Value{ConstInt(ctx.Int32Type(), 7, false)}, "")
	b.CreateRetVoid()
	if err := VerifyModule(m, ReturnStatusAction); err != nil {
		t.Fatalf("callee is not callable with requested type: %v\n%s", err, m)
	}
}

func TestGetOrInsertFunctionAlias(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	m := ctx.NewModule("alias")
	defer m.Dispose()
	ft := FunctionType(ctx.VoidType(), nil, false)
	fn := AddFunction(m, "implementation", ft)
	b := ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
	b.CreateRetVoid()
	alias := AddAlias(m, ft, 0, fn, "entrypoint")
	if got := m.GetOrInsertFunction("entrypoint", ft); got != alias || got.IsAGlobalAlias().IsNil() {
		t.Fatalf("expected existing alias, got %s", got)
	}
	if !m.NamedFunction("entrypoint.1").IsNil() {
		t.Fatal("created a duplicate declaration for an existing alias")
	}
	if err := VerifyModule(m, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}

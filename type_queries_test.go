// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

import "testing"

func TestParseTargetDataWithoutErrorOutput(t *testing.T) {
	if td := parseTargetData("e-p:abc", nil); td.C != nil {
		td.Dispose()
		t.Fatal("invalid layout accepted without an error pointer")
	}
	td := parseTargetData("e-p:64:64", nil)
	if td.C == nil {
		t.Fatal("valid layout failed without an error pointer")
	}
	td.Dispose()
}

func TestExtendedTypeQueries(t *testing.T) {
	c := NewContext()
	defer c.Dispose()
	i8 := c.Int8Type()
	for _, n := range []uint64{0, 3, 1<<32 + 7} {
		a := ArrayType64(i8, n)
		if a.ArrayLength64() != n || a.ElementType() != i8 {
			t.Fatalf("array count %d", n)
		}
	}
	for _, v := range []struct {
		typ      Type
		scalable bool
	}{{VectorType(i8, 4), false}, {ScalableVectorType(i8, 4), true}} {
		n, scale := v.typ.VectorElementCount()
		if n != 4 || scale != v.scalable || v.typ.ScalarSizeInBits() != 8 {
			t.Fatalf("vector %s: %d %v", v.typ, n, scale)
		}
	}
	for _, typ := range []Type{Type{}, i8, c.StructType(nil, false)} {
		if n, scale := typ.VectorElementCount(); n != 0 || scale {
			t.Fatal("non-vector has lanes")
		}
	}
	if c.HalfType().ScalarSizeInBits() != 16 || c.BFloatType().ScalarSizeInBits() != 16 || c.FP128Type().ScalarSizeInBits() != 128 || c.IntType(129).ScalarSizeInBits() != 129 {
		t.Fatal("scalar bit widths")
	}
	for _, k := range []TypeKind{HalfTypeKind, BFloatTypeKind, ScalableVectorTypeKind, TokenTypeKind} {
		if k.String() == "" {
			t.Fatal("missing type kind name")
		}
	}
	layout, err := ParseTargetData("e-p:64:64:64:32-p1:32:32:32:16")
	if err != nil {
		t.Fatal(err)
	}
	defer layout.Dispose()
	for as, want := range []uint{32, 16} {
		ptr := PointerType(i8, as)
		if layout.IndexTypeSizeInBits(ptr) != want || layout.IndexTypeSizeInBits(VectorType(ptr, 4)) != want {
			t.Fatalf("address space %d", as)
		}
	}
	if layout.IndexTypeSizeInBits(i8) != 0 || layout.IndexTypeSizeInBits(Type{}) != 0 || (TargetData{}).IndexTypeSizeInBits(i8) != 0 {
		t.Fatal("non-pointer accepted")
	}
	if td, err := ParseTargetData("e-p:abc"); err == nil || td.C != nil {
		t.Fatal("invalid layout accepted")
	}
}

func TestExactIntegerLiteralRanges(t *testing.T) {
	c := NewContext()
	defer c.Dispose()
	if !ConstInt(c.Int1Type(), 1, false).ConstIntInRange(-1, 0) ||
		!ConstInt(c.Int1Type(), 1, false).ConstUintInRange(1, 2) {
		t.Fatal("i1 sign/zero extension")
	}
	for _, tc := range []struct {
		text             string
		signed, unsigned bool
	}{
		{"7", true, true}, {"-1", false, false}, {"18446744073709551623", false, false},
		{"-18446744073709551609", false, false},
	} {
		v := ConstIntFromString(c.IntType(128), tc.text, 10)
		if v.ConstIntInRange(0, 8) != tc.signed || v.ConstUintInRange(0, 8) != tc.unsigned {
			t.Fatalf("truncated i128 %s", tc.text)
		}
	}
	for _, v := range []Value{Value{}, Undef(c.Int8Type()), ConstFloat(c.DoubleType(), 1)} {
		if v.ConstIntInRange(0, 8) || v.ConstUintInRange(0, 8) {
			t.Fatal("non-integer literal accepted")
		}
	}
}

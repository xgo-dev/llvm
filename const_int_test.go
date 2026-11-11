package llvm

import "testing"

// Narrow constants must be canonical even when callers pass a signed Go int
// converted to uint64. Invalid high bits are invisible in printed IR but affect
// constant uniquing and ScalarEvolution in LLVM 23.
func TestConstIntTruncation(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	for _, width := range []int{1, 8, 16, 32, 63, 64} {
		ty := ctx.IntType(width)
		mask := ^uint64(0) >> (64 - width)
		for _, n := range []uint64{0, 1, 255, 0x100000001, ^uint64(0), ^uint64(0) - 1} {
			for _, signed := range []bool{false, true} {
				got := ConstInt(ty, n, signed)
				want := ConstInt(ty, n&mask, false)
				if got != want || got.ZExtValue() != n&mask {
					t.Errorf("i%d n=%#x signed=%v: bits=%#x, want canonical %#x", width, n, signed, got.ZExtValue(), n&mask)
				}
			}
		}
	}
	for _, signed := range []bool{false, true} {
		got := ConstInt(ctx.IntType(128), ^uint64(0), signed)
		value := "18446744073709551615"
		if signed {
			value = "-1"
		}
		want := ConstIntFromString(ctx.IntType(128), value, 10)
		if got != want {
			t.Errorf("i128 signed=%v: got %s, want %s", signed, got.String(), want.String())
		}
	}
}

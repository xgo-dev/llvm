//go:build llvm14 || llvm15 || llvm16 || llvm17 || llvm18 || llvm19 || llvm20 || llvm21

package llvm

// GetSwitchCaseValue obtains the case value for a successor of a switch
// instruction. i corresponds to the successor index; the first successor (0)
// is the default destination, so i must be greater than zero.
//
// Before LLVM 22, switch case values were stored as regular instruction
// operands (alternating with their destination blocks, after the leading
// condition/default-destination pair), so this is implemented via Operand
// access instead of the LLVM 22+-only LLVMGetSwitchCaseValue.
func (v Value) GetSwitchCaseValue(i int) Value {
	return v.Operand(2 * i)
}

//===- target_test.go - Tests for target bindings -------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//

package llvm

import "testing"

func TestCreateTargetMachineWithOptionsSectionOptions(t *testing.T) {
	InitializeNativeTarget()

	triple := DefaultTargetTriple()
	target, err := GetTargetFromTriple(triple)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		opts TargetMachineOptions
	}{
		{
			name: "zero value",
			opts: TargetMachineOptions{},
		},
		{
			name: "mixed flags",
			opts: TargetMachineOptions{
				FunctionSections:   true,
				DataSections:       false,
				UniqueSectionNames: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tm := target.CreateTargetMachineWithOptions(
				triple, "", "", CodeGenLevelDefault, RelocDefault, CodeModelDefault, tc.opts,
			)
			if tm.C == nil {
				t.Fatal("CreateTargetMachineWithOptions returned a nil target machine")
			}
			defer tm.Dispose()

			if got := tm.sectionOptions(); got != tc.opts {
				t.Fatalf("section options mismatch: got %+v, want %+v", got, tc.opts)
			}
		})
	}
}

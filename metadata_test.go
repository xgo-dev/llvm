package llvm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNamedMetadataRoundTrip(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()

	mod := ctx.NewModule("metadata_roundtrip")
	defer mod.Dispose()

	mdstr := func(s string) Metadata {
		return ctx.MDString(s)
	}
	mdint := func(n uint64) Metadata {
		return ConstInt(ctx.Int32Type(), n, false).ConstantAsMetadata()
	}
	addRow := func(name string, fields ...Metadata) {
		mod.AddNamedMetadataOperand(name, ctx.MDNode(fields))
	}

	addRow("llgo.useiface",
		mdstr("github.com/goplus/llgo/demo.main"),
		mdstr("*_llgo_github.com/goplus/llgo/demo.File"),
	)
	addRow("llgo.useifacemethod",
		mdstr("github.com/goplus/llgo/demo.consume"),
		mdstr("_llgo_github.com/goplus/llgo/demo.ReadCloser"),
		mdstr("Read"),
		mdstr("_llgo_func$readsig"),
	)
	addRow("llgo.interfaceinfo",
		mdstr("_llgo_github.com/goplus/llgo/demo.ReadCloser"),
		mdstr("Read"),
		mdstr("_llgo_func$readsig"),
	)
	addRow("llgo.interfaceinfo",
		mdstr("_llgo_github.com/goplus/llgo/demo.ReadCloser"),
		mdstr("Close"),
		mdstr("_llgo_func$closesig"),
	)
	addRow("llgo.methodinfo",
		mdstr("*_llgo_github.com/goplus/llgo/demo.File"),
		mdint(0),
		mdstr("Read"),
		mdstr("_llgo_func$readsig"),
		mdstr("github.com/goplus/llgo/demo.(*File).Read"),
		mdstr("github.com/goplus/llgo/demo.File.Read"),
	)
	addRow("llgo.methodinfo",
		mdstr("*_llgo_github.com/goplus/llgo/demo.File"),
		mdint(1),
		mdstr("Close"),
		mdstr("_llgo_func$closesig"),
		mdstr("github.com/goplus/llgo/demo.(*File).Close"),
		mdstr("github.com/goplus/llgo/demo.File.Close"),
	)
	addRow("llgo.usenamedmethod",
		mdstr("github.com/goplus/llgo/demo.lookup"),
		mdstr("ServeHTTP"),
	)
	addRow("llgo.reflectmethod",
		mdstr("github.com/goplus/llgo/demo.lookup"),
	)

	ir := mod.String()
	for _, want := range []string{
		`!llgo.useiface = !{!0}`,
		`!0 = !{!"github.com/goplus/llgo/demo.main", !"*_llgo_github.com/goplus/llgo/demo.File"}`,
		`!llgo.useifacemethod = !{!1}`,
		`!1 = !{!"github.com/goplus/llgo/demo.consume", !"_llgo_github.com/goplus/llgo/demo.ReadCloser", !"Read", !"_llgo_func$readsig"}`,
		`!llgo.interfaceinfo = !{!2, !3}`,
		`!2 = !{!"_llgo_github.com/goplus/llgo/demo.ReadCloser", !"Read", !"_llgo_func$readsig"}`,
		`!3 = !{!"_llgo_github.com/goplus/llgo/demo.ReadCloser", !"Close", !"_llgo_func$closesig"}`,
		`!llgo.methodinfo = !{!4, !5}`,
		`!4 = !{!"*_llgo_github.com/goplus/llgo/demo.File", i32 0, !"Read", !"_llgo_func$readsig", !"github.com/goplus/llgo/demo.(*File).Read", !"github.com/goplus/llgo/demo.File.Read"}`,
		`!5 = !{!"*_llgo_github.com/goplus/llgo/demo.File", i32 1, !"Close", !"_llgo_func$closesig", !"github.com/goplus/llgo/demo.(*File).Close", !"github.com/goplus/llgo/demo.File.Close"}`,
		`!llgo.usenamedmethod = !{!6}`,
		`!6 = !{!"github.com/goplus/llgo/demo.lookup", !"ServeHTTP"}`,
		`!llgo.reflectmethod = !{!7}`,
		`!7 = !{!"github.com/goplus/llgo/demo.lookup"}`,
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("module IR missing expected fragment:\n%s\n\nfull IR:\n%s", want, ir)
		}
	}

	path := filepath.Join(t.TempDir(), "metadata.ll")
	if err := os.WriteFile(path, []byte(ir), 0o600); err != nil {
		t.Fatal(err)
	}

	parseCtx := NewContext()
	defer parseCtx.Dispose()

	buf, err := NewMemoryBufferFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := (&parseCtx).ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer parsed.Dispose()

	if got := parsed.NamedMetadataNumOperands("llgo.useiface"); got != 1 {
		t.Fatalf("NamedMetadataNumOperands(llgo.useiface) = %d, want 1", got)
	}
	if got := parsed.NamedMetadataNumOperands("llgo.interfaceinfo"); got != 2 {
		t.Fatalf("NamedMetadataNumOperands(llgo.interfaceinfo) = %d, want 2", got)
	}
	if got := parsed.NamedMetadataNumOperands("llgo.methodinfo"); got != 2 {
		t.Fatalf("NamedMetadataNumOperands(llgo.methodinfo) = %d, want 2", got)
	}
	if got := parsed.NamedMetadataNumOperands("llgo.missing"); got != 0 {
		t.Fatalf("NamedMetadataNumOperands(llgo.missing) = %d, want 0", got)
	}
	if got := parsed.NamedMetadataOperands("llgo.missing"); len(got) != 0 {
		t.Fatalf("NamedMetadataOperands(llgo.missing) returned %d rows, want 0", len(got))
	}

	useIfaceRows := parsed.NamedMetadataOperands("llgo.useiface")
	requireMDStrings(t, useIfaceRows, 0,
		"github.com/goplus/llgo/demo.main",
		"*_llgo_github.com/goplus/llgo/demo.File",
	)

	useIfaceMethodRows := parsed.NamedMetadataOperands("llgo.useifacemethod")
	requireMDStrings(t, useIfaceMethodRows, 0,
		"github.com/goplus/llgo/demo.consume",
		"_llgo_github.com/goplus/llgo/demo.ReadCloser",
		"Read",
		"_llgo_func$readsig",
	)

	interfaceInfoRows := parsed.NamedMetadataOperands("llgo.interfaceinfo")
	requireMDStrings(t, interfaceInfoRows, 0,
		"_llgo_github.com/goplus/llgo/demo.ReadCloser",
		"Read",
		"_llgo_func$readsig",
	)
	requireMDStrings(t, interfaceInfoRows, 1,
		"_llgo_github.com/goplus/llgo/demo.ReadCloser",
		"Close",
		"_llgo_func$closesig",
	)

	methodInfoRows := parsed.NamedMetadataOperands("llgo.methodinfo")
	requireMethodInfoRow(t, methodInfoRows, 0,
		"*_llgo_github.com/goplus/llgo/demo.File",
		0,
		"Read",
		"_llgo_func$readsig",
		"github.com/goplus/llgo/demo.(*File).Read",
		"github.com/goplus/llgo/demo.File.Read",
	)
	requireMethodInfoRow(t, methodInfoRows, 1,
		"*_llgo_github.com/goplus/llgo/demo.File",
		1,
		"Close",
		"_llgo_func$closesig",
		"github.com/goplus/llgo/demo.(*File).Close",
		"github.com/goplus/llgo/demo.File.Close",
	)

	useNamedMethodRows := parsed.NamedMetadataOperands("llgo.usenamedmethod")
	requireMDStrings(t, useNamedMethodRows, 0,
		"github.com/goplus/llgo/demo.lookup",
		"ServeHTTP",
	)

	reflectMethodRows := parsed.NamedMetadataOperands("llgo.reflectmethod")
	requireMDStrings(t, reflectMethodRows, 0,
		"github.com/goplus/llgo/demo.lookup",
	)
}

func requireMDStrings(t *testing.T, rows []Value, rowIndex int, want ...string) {
	t.Helper()
	if rowIndex >= len(rows) {
		t.Fatalf("row index %d out of range for %d rows", rowIndex, len(rows))
	}
	fields := rows[rowIndex].MDNodeOperands()
	if len(fields) != len(want) {
		t.Fatalf("row %d has %d fields, want %d", rowIndex, len(fields), len(want))
	}
	for i, field := range fields {
		if !field.IsAMDString() {
			t.Fatalf("row %d field %d is not an MDString", rowIndex, i)
		}
		if got := field.MDString(); got != want[i] {
			t.Fatalf("row %d field %d = %q, want %q", rowIndex, i, got, want[i])
		}
	}
}

func requireMethodInfoRow(t *testing.T, rows []Value, rowIndex int, wantType string, wantIndex uint64, wantName string, wantMType string, wantIFn string, wantTFn string) {
	t.Helper()
	if rowIndex >= len(rows) {
		t.Fatalf("row index %d out of range for %d rows", rowIndex, len(rows))
	}
	fields := rows[rowIndex].MDNodeOperands()
	if len(fields) != 6 {
		t.Fatalf("methodinfo row %d has %d fields, want 6", rowIndex, len(fields))
	}
	if !fields[0].IsAMDString() {
		t.Fatalf("methodinfo row %d field 0 is not an MDString", rowIndex)
	}
	if got := fields[0].MDString(); got != wantType {
		t.Fatalf("methodinfo row %d field 0 = %q, want %q", rowIndex, got, wantType)
	}
	if fields[1].IsAMDString() {
		t.Fatalf("methodinfo row %d field 1 unexpectedly reports MDString", rowIndex)
	}
	if got := fields[1].ZExtValue(); got != wantIndex {
		t.Fatalf("methodinfo row %d field 1 = %d, want %d", rowIndex, got, wantIndex)
	}
	for i, want := range []string{wantName, wantMType, wantIFn, wantTFn} {
		fieldIndex := i + 2
		if !fields[fieldIndex].IsAMDString() {
			t.Fatalf("methodinfo row %d field %d is not an MDString", rowIndex, fieldIndex)
		}
		if got := fields[fieldIndex].MDString(); got != want {
			t.Fatalf("methodinfo row %d field %d = %q, want %q", rowIndex, fieldIndex, got, want)
		}
	}
}

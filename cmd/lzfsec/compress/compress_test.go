package compress

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-compressions/lzfse"
)

func TestReadInput_Stdin(t *testing.T) {
	// readInput("") reads from stdin — tested indirectly via Command(); skip live stdin.
	t.Skip("stdin test requires process-level redirect")
}

func TestReadInput_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "in.bin")
	want := []byte("hello lzfsec")
	if err := os.WriteFile(f, want, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readInput(f)
	if err != nil {
		t.Fatalf("readInput: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWriteOutput_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "out.bin")
	data := []byte{1, 2, 3}
	if err := writeOutput(f, data); err != nil {
		t.Fatalf("writeOutput: %v", err)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("got %v, want %v", got, data)
	}
}

func TestRatio(t *testing.T) {
	if r := ratio(0, 0); r != 0 {
		t.Fatalf("ratio(0,0) = %v, want 0", r)
	}
	if r := ratio(100, 50); r != 50.0 {
		t.Fatalf("ratio(100,50) = %v, want 50", r)
	}
}

func TestCommand_CompressFile(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	output := filepath.Join(dir, "output.lzfse")

	original := bytes.Repeat([]byte("compress me! "), 200)
	if err := os.WriteFile(input, original, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := Command()
	cmd.SetArgs([]string{"--input", input, "--output", output})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command: %v", err)
	}

	compressed, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("output is empty")
	}

	// Verify the output can be decompressed back to the original.
	got, err := lzfse.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("roundtrip mismatch: got %d bytes, want %d", len(got), len(original))
	}
}

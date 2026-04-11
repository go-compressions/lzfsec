package decompress

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-compressions/lzfse"
)

func TestReadInput_Stdin(t *testing.T) {
	t.Skip("stdin test requires process-level redirect")
}

func TestReadInput_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "in.lzfse")
	want := []byte("hello lzfsec decompressor")
	compressed, err := lzfse.Compress(want)
	if err != nil {
		t.Fatalf("lzfse.Compress: %v", err)
	}
	if err := os.WriteFile(f, compressed, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readInput(f)
	if err != nil {
		t.Fatalf("readInput: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("got empty data")
	}
}

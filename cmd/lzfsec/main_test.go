package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestCompressDecompressRoundtrip(t *testing.T) {
    // Use the binary via `go run` to ensure module uses new path.
    src := []byte("hello lzfse world")

    // Compress
    cmd := exec.Command("go", "run", "./main.go", "compress")
    cmd.Stdin = bytes.NewReader(src)
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("compress failed: %v: %s", err, string(out))
    }

    // Decompress
    cmd = exec.Command("go", "run", "./main.go", "decompress")
    cmd.Stdin = bytes.NewReader(out)
    out2, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("decompress failed: %v: %s", err, string(out2))
    }
    if string(out2) != string(src) {
        t.Fatalf("roundtrip mismatch: got %q want %q", string(out2), string(src))
    }
}

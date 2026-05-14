package decompress

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-compressions/lzfse"
)

func TestCommand_File(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.lzfse")
	output := filepath.Join(dir, "out.bin")

	original := bytes.Repeat([]byte("decompress me! "), 200)
	compressed, err := lzfse.Compress(original)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if err := os.WriteFile(input, compressed, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := Command()
	var stderr bytes.Buffer
	cmd.SetArgs([]string{"--input", input, "--output", output})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command: %v", err)
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("roundtrip mismatch: got %d bytes, want %d", len(got), len(original))
	}
	if !strings.Contains(stderr.String(), "decompressed") {
		t.Fatalf("stderr summary missing: %q", stderr.String())
	}
}

func TestCommand_InputNotFound(t *testing.T) {
	cmd := Command()
	cmd.SetArgs([]string{"--input", "/does/not/exist/lzfsec.input"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing input file")
	}
}

func TestCommand_BadCompressedData(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "garbage.lzfse")
	// Random bytes that don't parse as an LZFSE stream.
	if err := os.WriteFile(input, []byte{0xDE, 0xAD, 0xBE, 0xEF}, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := Command()
	cmd.SetArgs([]string{"--input", input})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected decompress error for garbage input")
	}
}

func TestCommand_OutputNotWritable(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.lzfse")
	compressed, _ := lzfse.Compress([]byte("hi"))
	if err := os.WriteFile(input, compressed, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "no-such-subdir", "out.bin")
	cmd := Command()
	cmd.SetArgs([]string{"--input", input, "--output", output})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unwriteable output path")
	}
}

// TestCommand_StdinStdoutRoundtrip: pipe compressed data to stdin,
// confirm decompressed bytes land on stdout and no summary line
// pollutes the binary output.
func TestCommand_StdinStdoutRoundtrip(t *testing.T) {
	payload := []byte("hello from stdin to stdout")
	compressed, err := lzfse.Compress(payload)
	if err != nil {
		t.Fatal(err)
	}
	got := withStdinAndStdoutCapture(t, compressed, func() {
		cmd := Command()
		var stderr bytes.Buffer
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&stderr)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("command: %v", err)
		}
		if strings.Contains(stderr.String(), "decompressed") {
			t.Errorf("stderr summary printed for stdout output: %q", stderr.String())
		}
	})
	if !bytes.Equal(got, payload) {
		t.Fatalf("stdin→stdout roundtrip: got %q, want %q", got, payload)
	}
}

// withStdinAndStdoutCapture pipes `stdin` to os.Stdin, replaces
// os.Stdout with a pipe, runs fn, and returns whatever was written
// to stdout.
func withStdinAndStdoutCapture(t *testing.T, stdin []byte, fn func()) []byte {
	t.Helper()
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prevIn, prevOut := os.Stdin, os.Stdout
	os.Stdin = stdinR
	os.Stdout = stdoutW
	defer func() {
		os.Stdin = prevIn
		os.Stdout = prevOut
	}()
	go func() {
		stdinW.Write(stdin)
		stdinW.Close()
	}()
	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		buf.ReadFrom(stdoutR)
		done <- buf.Bytes()
	}()
	fn()
	stdoutW.Close()
	return <-done
}

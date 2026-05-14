package compress

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-compressions/lzfse"
)

func TestRatio(t *testing.T) {
	if r := Ratio(0, 0); r != 0 {
		t.Fatalf("Ratio(0,0) = %v, want 0", r)
	}
	if r := Ratio(100, 50); r != 50.0 {
		t.Fatalf("Ratio(100,50) = %v, want 50", r)
	}
	if r := Ratio(200, 50); r != 25.0 {
		t.Fatalf("Ratio(200,50) = %v, want 25", r)
	}
}

func TestCommand_File(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	output := filepath.Join(dir, "output.lzfse")

	original := bytes.Repeat([]byte("compress me! "), 200)
	if err := os.WriteFile(input, original, 0o644); err != nil {
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

	compressed, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("output is empty")
	}
	got, err := lzfse.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("roundtrip mismatch: got %d bytes, want %d", len(got), len(original))
	}
	if !strings.Contains(stderr.String(), "compressed") {
		t.Fatalf("stderr summary missing: %q", stderr.String())
	}
}

// TestCommand_StdoutNoSummary: when output is stdout (--output omitted),
// the human-readable summary line MUST NOT be printed (we don't want it
// mixed with binary compressed data on stdout).
func TestCommand_StdoutNoSummary(t *testing.T) {
	withStdin(t, []byte("hello, stdin"), func() {
		withStdoutCapture(t, func() []byte {
			cmd := Command()
			var stderr bytes.Buffer
			cmd.SetArgs([]string{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&stderr)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("command: %v", err)
			}
			if strings.Contains(stderr.String(), "compressed") {
				t.Errorf("stderr summary printed for stdout output: %q", stderr.String())
			}
			return nil
		})
	})
}

// TestCommand_InputNotFound surfaces an os.ReadFile error.
func TestCommand_InputNotFound(t *testing.T) {
	cmd := Command()
	cmd.SetArgs([]string{"--input", "/does/not/exist/lzfsec.input"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing input file")
	}
}

// TestCommand_OutputNotWritable surfaces an os.WriteFile error
// (output path under an unwriteable directory).
func TestCommand_OutputNotWritable(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(input, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Output path under a non-existent directory → write fails.
	output := filepath.Join(dir, "no-such-subdir", "out.lzfse")
	cmd := Command()
	cmd.SetArgs([]string{"--input", input, "--output", output})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unwriteable output path")
	}
}

// withStdin pipes `data` to os.Stdin for the duration of fn.
func withStdin(t *testing.T, data []byte, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prev := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = prev
		r.Close()
	}()
	go func() {
		w.Write(data)
		w.Close()
	}()
	fn()
}

// withStdoutCapture replaces os.Stdout with a pipe, runs fn, and
// returns the bytes written to stdout.
func withStdoutCapture(t *testing.T, fn func() []byte) []byte {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prev := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = prev }()

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		buf.ReadFrom(r)
		done <- buf.Bytes()
	}()
	_ = fn()
	w.Close()
	return <-done
}

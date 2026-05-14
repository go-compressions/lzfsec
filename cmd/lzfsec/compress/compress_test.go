package compress

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestFormatDuration(t *testing.T) {
	// Sub-microsecond values keep their nanosecond precision.
	if got := FormatDuration(500 * time.Nanosecond); !strings.Contains(got, "ns") {
		t.Errorf("FormatDuration(500ns): %q does not contain ns", got)
	}
	// Microsecond+ values are rounded to whole microseconds.
	if got := FormatDuration(1234567 * time.Nanosecond); got != "1.235ms" {
		t.Errorf("FormatDuration(1.234567ms): got %q, want 1.235ms", got)
	}
	// Zero stays zero.
	if got := FormatDuration(0); got != "0s" {
		t.Errorf("FormatDuration(0): got %q, want 0s", got)
	}
}

func TestCommand_File_Silent(t *testing.T) {
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
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty without --verbose, got: %q", stderr.String())
	}

	compressed, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	got, err := lzfse.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("roundtrip mismatch: got %d bytes, want %d", len(got), len(original))
	}
}

// TestCommand_File_Verbose: with --verbose the summary line lands on
// stderr and includes the byte counts, ratio, and an elapsed time.
func TestCommand_File_Verbose(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	output := filepath.Join(dir, "output.lzfse")

	original := bytes.Repeat([]byte("compress me! "), 200)
	if err := os.WriteFile(input, original, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := Command()
	// Persistent --verbose lives on the root in production but
	// here we wire it directly onto the sub-command so the test
	// can exercise Command() without standing up the full root.
	cmd.Flags().BoolP("verbose", "v", false, "verbose")
	var stderr bytes.Buffer
	cmd.SetArgs([]string{"--input", input, "--output", output, "--verbose"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command: %v", err)
	}
	line := stderr.String()
	for _, want := range []string{"compressed", "bytes", "%", " in "} {
		if !strings.Contains(line, want) {
			t.Errorf("stderr %q missing %q", line, want)
		}
	}
}

// TestCommand_StdoutNoSummary: when output is stdout (--output omitted)
// and --verbose is NOT set, the summary line stays absent.
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
			if stderr.Len() != 0 {
				t.Errorf("stderr should be empty: %q", stderr.String())
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

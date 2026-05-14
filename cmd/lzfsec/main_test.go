package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRootCmd_NoArgs prints the usage to stdout via cobra's help
// command. We mostly want to confirm Execute returns nil.
func TestRootCmd_NoArgs(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

// TestRootCmd_Roundtrip drives compress then decompress through
// the cobra commands directly (no `go run` subprocess), feeding
// data via a temp file.
func TestRootCmd_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.bin")
	compressed := filepath.Join(dir, "src.lzfse")
	round := filepath.Join(dir, "roundtripped.bin")
	original := bytes.Repeat([]byte("roundtrip via cobra "), 100)
	if err := os.WriteFile(src, original, 0o644); err != nil {
		t.Fatal(err)
	}

	c := RootCmd()
	c.SetArgs([]string{"compress", "--input", src, "--output", compressed})
	c.SetOut(&bytes.Buffer{})
	c.SetErr(&bytes.Buffer{})
	if err := c.Execute(); err != nil {
		t.Fatalf("compress: %v", err)
	}

	d := RootCmd()
	d.SetArgs([]string{"decompress", "--input", compressed, "--output", round})
	d.SetOut(&bytes.Buffer{})
	d.SetErr(&bytes.Buffer{})
	if err := d.Execute(); err != nil {
		t.Fatalf("decompress: %v", err)
	}

	got, err := os.ReadFile(round)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("roundtrip mismatch: got %d bytes, want %d", len(got), len(original))
	}
}

// TestRootCmd_PropagatesError feeds compress an input path that
// doesn't exist; the resulting error must surface from Execute.
func TestRootCmd_PropagatesError(t *testing.T) {
	c := RootCmd()
	c.SetArgs([]string{"compress", "--input", "/does/not/exist/lzfsec.input"})
	c.SetOut(&bytes.Buffer{})
	c.SetErr(&bytes.Buffer{})
	err := c.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no such file") &&
		!errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error %q does not mention missing file", err)
	}
}

// TestMain_OsExitOnError replaces osExit and feeds RootCmd a path
// guaranteed to fail. main() must call osExit(1).
func TestMain_OsExitOnError(t *testing.T) {
	prevExit := osExit
	prevArgs := os.Args
	defer func() {
		osExit = prevExit
		os.Args = prevArgs
	}()

	called := false
	osExit = func(code int) {
		called = true
		if code != 1 {
			t.Errorf("osExit code: got %d, want 1", code)
		}
	}
	os.Args = []string{"lzfsec", "compress", "--input", "/does/not/exist"}
	main()
	if !called {
		t.Fatal("osExit not called")
	}
}

// TestMain_NoArgsExitsCleanly runs main() with no sub-command;
// cobra prints help and main() returns without calling osExit.
func TestMain_NoArgsExitsCleanly(t *testing.T) {
	prevExit := osExit
	prevArgs := os.Args
	defer func() {
		osExit = prevExit
		os.Args = prevArgs
	}()

	osExit = func(code int) {
		t.Fatalf("osExit called with %d on no-args invocation", code)
	}
	os.Args = []string{"lzfsec"}
	// Capture stdout to avoid noise.
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = stdout
		_ = r // discard
	}()
	main()
}

package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metalagman/semverctl/internal/cli"
)

func runMainWithArgs(t *testing.T, args []string) (string, int, bool) {
	t.Helper()

	oldArgs := os.Args
	oldStdout := os.Stdout
	oldExecute := execute
	oldExit := exit
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		execute = oldExecute
		exit = oldExit
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}

	os.Stdout = w
	os.Args = append([]string{"semverctl"}, args...)

	exitCalled := false
	exitCode := 0
	exit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	execute = cliExecuteForMain

	main()

	if err := w.Close(); err != nil {
		t.Fatalf("close write pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close read pipe: %v", err)
	}

	return string(out), exitCode, exitCalled
}

func cliExecuteForMain() error {
	return cli.Execute()
}

func TestMainVersionCommand(t *testing.T) {
	out, _, exitCalled := runMainWithArgs(t, []string{"version"})
	if exitCalled {
		t.Fatal("main() should not call exit on successful version command")
	}
	if !strings.Contains(out, "semverctl version ") {
		t.Fatalf("main() output %q does not contain version banner", out)
	}
}

func TestMainHelpCommand(t *testing.T) {
	out, _, exitCalled := runMainWithArgs(t, []string{"--help"})
	if exitCalled {
		t.Fatal("main() should not call exit on help")
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("main() output %q does not contain usage information", out)
	}
}

func TestMainBumpFileJSON(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(file, []byte("{\n  \"version\": \"1.0.0\",\n  \"name\": \"demo\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	out, _, exitCalled := runMainWithArgs(t, []string{"bump", "file", file, "--json"})
	if exitCalled {
		t.Fatal("main() should not call exit on successful bump")
	}
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("expected successful JSON output, got %q", out)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	if !strings.Contains(string(got), `"1.0.1"`) {
		t.Fatalf("expected bumped version in file, got %q", string(got))
	}
}

func TestMainSetFileDryRunDoesNotMutate(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "package.json")
	original := "{\n  \"version\": \"1.0.0\"\n}\n"
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	out, _, exitCalled := runMainWithArgs(t, []string{"set", "file", "2.0.0", file, "--dry-run"})
	if exitCalled {
		t.Fatal("main() should not call exit on successful dry-run")
	}
	if !strings.Contains(out, "+  \"version\": \"2.0.0\"") && !strings.Contains(out, "2.0.0") {
		t.Fatalf("expected dry-run output to mention new version, got %q", out)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	if string(got) != original {
		t.Fatalf("dry-run should not mutate file, got %q", string(got))
	}
}

func TestMainSetFileJSON(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(file, []byte("{\"version\":\"1.0.0\"}\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	out, _, exitCalled := runMainWithArgs(t, []string{"set", "file", "2.0.0", file, "--dry-run=false", "--json"})
	if exitCalled {
		t.Fatal("main() should not call exit on successful set")
	}
	if !strings.Contains(out, `"command":"set file"`) || !strings.Contains(out, `"ok":true`) {
		t.Fatalf("expected successful JSON output, got %q", out)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	if !strings.Contains(string(got), `"2.0.0"`) {
		t.Fatalf("expected set version in file, got %q", string(got))
	}
}

func TestMainExecuteErrorExitsWithCodeOne(t *testing.T) {
	oldExecute := execute
	oldExit := exit
	defer func() {
		execute = oldExecute
		exit = oldExit
	}()

	execute = func() error { return errors.New("boom") }
	exitCalled := false
	exitCode := 0
	exit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	main()

	if !exitCalled {
		t.Fatal("main() should call exit when execute returns an error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
}

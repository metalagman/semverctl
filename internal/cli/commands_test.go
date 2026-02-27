package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func setFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set flag %q=%q: %v", name, value, err)
	}
}

func newBumpTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("glob", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("numeric", false, "")
	return cmd
}

func newSetTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("glob", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

func TestRunBump_DefaultTargetPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.WriteFile("package.json", []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	cmd := newBumpTestCmd()
	setFlag(t, cmd, "dry-run", "true")

	if err := runBump(cmd, nil); err != nil {
		t.Fatalf("runBump() with default target returned error: %v", err)
	}
}

func TestRunBump_DefaultTargetMissingPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	cmd := newBumpTestCmd()
	setFlag(t, cmd, "dry-run", "true")

	err := runBump(cmd, nil)
	if err == nil {
		t.Fatal("runBump() should error when default package.json is missing")
	}
	if !strings.Contains(err.Error(), `stat "package.json"`) {
		t.Fatalf("runBump() error = %v, want error containing stat package.json", err)
	}
}

func TestRunBump_RejectsPositionalArgs(t *testing.T) {
	cmd := newBumpTestCmd()
	if err := runBump(cmd, []string{"package.json"}); err == nil {
		t.Fatal("runBump() should error on positional args")
	}
}

func TestRunBump_Glob(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.WriteFile("a.json", []byte(`{"version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write a.json: %v", err)
	}
	if err := os.WriteFile("b.yaml", []byte("version: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write b.yaml: %v", err)
	}
	if err := os.WriteFile("ignored.txt", []byte("noop"), 0o644); err != nil {
		t.Fatalf("write ignored.txt: %v", err)
	}

	cmd := newBumpTestCmd()
	setFlag(t, cmd, "glob", "**/*")
	setFlag(t, cmd, "dry-run", "true")

	if err := runBump(cmd, nil); err != nil {
		t.Fatalf("runBump() with --glob returned error: %v", err)
	}
}

func TestRunBump_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "multiple bump types",
			flags:   map[string]string{"file": testFile, "major": "true", "minor": "true"},
			wantErr: true,
		},
		{
			name:    "numeric with bump type",
			flags:   map[string]string{"file": testFile, "numeric": "true", "patch": "true"},
			wantErr: true,
		},
		{
			name:    "invalid path",
			flags:   map[string]string{"file": testFile, "path": ".."},
			wantErr: true,
		},
		{
			name:    "file and glob conflict",
			flags:   map[string]string{"file": testFile, "glob": "**/*"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newBumpTestCmd()
			for k, v := range tt.flags {
				setFlag(t, cmd, k, v)
			}

			err := runBump(cmd, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("runBump() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunSet_DefaultTargetPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.WriteFile("package.json", []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	cmd := newSetTestCmd()
	setFlag(t, cmd, "dry-run", "true")

	if err := runSet(cmd, []string{"1.2.3"}); err != nil {
		t.Fatalf("runSet() with default target returned error: %v", err)
	}
}

func TestRunSet_DefaultTargetMissingPackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	cmd := newSetTestCmd()
	setFlag(t, cmd, "dry-run", "true")

	err := runSet(cmd, []string{"1.2.3"})
	if err == nil {
		t.Fatal("runSet() should error when default package.json is missing")
	}
	if !strings.Contains(err.Error(), `stat "package.json"`) {
		t.Fatalf("runSet() error = %v, want error containing stat package.json", err)
	}
}

func TestRunSet_RejectsExtraPositionalArgs(t *testing.T) {
	cmd := newSetTestCmd()
	if err := runSet(cmd, []string{"1.2.3", "package.json"}); err == nil {
		t.Fatal("runSet() should error on extra positional args")
	}
}

func TestRunSet_Glob(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.WriteFile("a.json", []byte(`{"version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write a.json: %v", err)
	}
	if err := os.WriteFile("b.yaml", []byte("version: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write b.yaml: %v", err)
	}

	cmd := newSetTestCmd()
	setFlag(t, cmd, "glob", "**/*")
	setFlag(t, cmd, "dry-run", "true")

	if err := runSet(cmd, []string{"2.0.0"}); err != nil {
		t.Fatalf("runSet() with --glob returned error: %v", err)
	}
}

func TestRunSet_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "missing version",
			args:    []string{},
			flags:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid path",
			args:    []string{"1.2.3"},
			flags:   map[string]string{"file": testFile, "path": ".."},
			wantErr: true,
		},
		{
			name:    "file and glob conflict",
			args:    []string{"1.2.3"},
			flags:   map[string]string{"file": testFile, "glob": "**/*"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSetTestCmd()
			for k, v := range tt.flags {
				setFlag(t, cmd, k, v)
			}

			err := runSet(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("runSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadBumpFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Int("path", 0, "")
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("glob", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("numeric", false, "")

	if _, err := loadBumpFlags(cmd); err == nil {
		t.Fatal("loadBumpFlags() should error when path flag has wrong type")
	}
}

func TestLoadSetFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("glob", "", "")
	cmd.Flags().String("dry-run", "false", "")

	if _, err := loadSetFlags(cmd); err == nil {
		t.Fatal("loadSetFlags() should error when dry-run flag has wrong type")
	}
}

func TestBumpCommandFlags(t *testing.T) {
	bumpCmd := &cobra.Command{Use: "bump", Short: "Bump version"}

	bumpCmd.Flags().String("path", "version", "")
	bumpCmd.Flags().String("file", "", "")
	bumpCmd.Flags().String("glob", "", "")
	bumpCmd.Flags().Bool("dry-run", false, "")
	bumpCmd.Flags().Bool("major", false, "")
	bumpCmd.Flags().Bool("minor", false, "")
	bumpCmd.Flags().Bool("patch", false, "")
	bumpCmd.Flags().Bool("numeric", false, "")

	requiredFlags := []string{"path", "file", "glob", "dry-run", "major", "minor", "patch", "numeric"}
	for _, flag := range requiredFlags {
		if bumpCmd.Flags().Lookup(flag) == nil {
			t.Errorf("bump command missing flag: %s", flag)
		}
	}
}

func TestSetCommandFlags(t *testing.T) {
	setCmd := &cobra.Command{Use: "set", Short: "Set version"}

	setCmd.Flags().String("path", "version", "")
	setCmd.Flags().String("file", "", "")
	setCmd.Flags().String("glob", "", "")
	setCmd.Flags().Bool("dry-run", false, "")

	requiredFlags := []string{"path", "file", "glob", "dry-run"}
	for _, flag := range requiredFlags {
		if setCmd.Flags().Lookup(flag) == nil {
			t.Errorf("set command missing flag: %s", flag)
		}
	}
}

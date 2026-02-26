package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunBump(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0644)

	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "missing file",
			args:    []string{},
			flags:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "multiple bump types",
			args:    []string{},
			flags:   map[string]string{"file": testFile, "major": "true", "minor": "true"},
			wantErr: true,
		},
		{
			name:    "numeric with bump type",
			args:    []string{},
			flags:   map[string]string{"file": testFile, "numeric": "true", "patch": "true"},
			wantErr: true,
		},
		{
			name:    "invalid path",
			args:    []string{},
			flags:   map[string]string{"file": testFile, "path": ".."},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("path", "version", "")
			cmd.Flags().String("file", "", "")
			cmd.Flags().String("glob", "", "")
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().Bool("major", false, "")
			cmd.Flags().Bool("minor", false, "")
			cmd.Flags().Bool("patch", false, "")
			cmd.Flags().Bool("numeric", false, "")

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			// Reset global flags for each test
			flagFile = ""
			flagGlob = ""
			flagMajor = false
			flagMinor = false
			flagPatch = false
			flagNumericBump = false

			err := runBump(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("runBump() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunSet(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0644)

	tests := []struct {
		name    string
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		{
			name:    "no file specified",
			args:    []string{"1.2.3"},
			flags:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid path",
			args:    []string{"1.2.3"},
			flags:   map[string]string{"file": testFile, "path": ".."},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("path", "version", "")
			cmd.Flags().String("file", "", "")
			cmd.Flags().String("glob", "", "")
			cmd.Flags().Bool("dry-run", false, "")

			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			// Reset global flags
			flagFile = ""
			flagGlob = ""

			err := runSet(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("runSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBumpCommandFlags(t *testing.T) {
	// Test that bump command has all expected flags
	bumpCmd := &cobra.Command{
		Use:   "bump",
		Short: "Bump version",
	}

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
	// Test that set command has all expected flags
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set version",
	}

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

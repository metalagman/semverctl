package app

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid bump patch",
			config: &Config{
				Operation: OpBump,
				BumpType:  "patch",
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: false,
		},
		{
			name: "valid bump major",
			config: &Config{
				Operation: OpBump,
				BumpType:  "major",
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: false,
		},
		{
			name: "valid set",
			config: &Config{
				Operation: OpSet,
				Version:   "1.2.3",
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: false,
		},
		{
			name: "valid numeric bump",
			config: &Config{
				Operation:   OpBump,
				Path:        []string{"version"},
				File:        "test.json",
				NumericBump: true,
			},
			wantErr: false,
		},
		{
			name: "bump without type",
			config: &Config{
				Operation: OpBump,
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: true,
		},
		{
			name: "numeric bump with bump type",
			config: &Config{
				Operation:   OpBump,
				BumpType:    "patch",
				Path:        []string{"version"},
				File:        "test.json",
				NumericBump: true,
			},
			wantErr: true,
		},
		{
			name: "set without version",
			config: &Config{
				Operation: OpSet,
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: true,
		},
		{
			name: "set with invalid version",
			config: &Config{
				Operation: OpSet,
				Version:   "invalid",
				Path:      []string{"version"},
				File:      "test.json",
			},
			wantErr: true,
		},
		{
			name: "set with numeric bump",
			config: &Config{
				Operation:   OpSet,
				Version:     "1.2.3",
				Path:        []string{"version"},
				File:        "test.json",
				NumericBump: true,
			},
			wantErr: true,
		},
		{
			name: "both file and glob",
			config: &Config{
				Operation: OpBump,
				BumpType:  "patch",
				Path:      []string{"version"},
				File:      "test.json",
				Glob:      "*.json",
			},
			wantErr: true,
		},
		{
			name: "neither file nor glob",
			config: &Config{
				Operation: OpBump,
				BumpType:  "patch",
				Path:      []string{"version"},
			},
			wantErr: true,
		},
		{
			name: "default path",
			config: &Config{
				Operation: OpBump,
				BumpType:  "patch",
				File:      "test.json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewRunner(tt.config)
			err := runner.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunner_Run(t *testing.T) {
	// Create temp test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	writeFile(t, testFile, []byte(`{"version": "1.0.0"}`))

	t.Run("bump patch dry run", func(t *testing.T) {
		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			File:      testFile,
			DryRun:    true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	})

	t.Run("set dry run", func(t *testing.T) {
		config := &Config{
			Operation: OpSet,
			Version:   "2.0.0",
			Path:      []string{"version"},
			File:      testFile,
			DryRun:    true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	})

	t.Run("numeric bump dry run", func(t *testing.T) {
		numericFile := filepath.Join(tmpDir, "numeric.json")
		writeFile(t, numericFile, []byte(`{"count": 42}`))

		config := &Config{
			Operation:   OpBump,
			Path:        []string{"count"},
			File:        numericFile,
			DryRun:      true,
			NumericBump: true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			File:      filepath.Join(tmpDir, "missing.json"),
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err == nil {
			t.Error("Run() should error on missing file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		writeFile(t, invalidFile, []byte(`{invalid`))

		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			File:      invalidFile,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err == nil {
			t.Error("Run() should error on invalid JSON")
		}
	})

	t.Run("path not found", func(t *testing.T) {
		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"missing"},
			File:      testFile,
			DryRun:    true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err == nil {
			t.Error("Run() should error on missing path")
		}
	})

	t.Run("invalid semver", func(t *testing.T) {
		invalidVerFile := filepath.Join(tmpDir, "invalidver.json")
		writeFile(t, invalidVerFile, []byte(`{"version": "not-a-version"}`))

		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			File:      invalidVerFile,
			DryRun:    true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err == nil {
			t.Error("Run() should error on invalid semver")
		}
	})
}

func TestRunner_RunWithResults(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	writeFile(t, testFile, []byte(`{"version": "1.0.0"}`))

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	config := &Config{
		Operation: OpBump,
		BumpType:  "patch",
		Path:      []string{"version"},
		File:      testFile,
		DryRun:    true,
	}
	runner := NewRunner(config)
	results, runErr := runner.RunWithResults()
	if runErr != nil {
		t.Fatalf("RunWithResults() error = %v", runErr)
	}
	if len(results) != 1 {
		t.Fatalf("RunWithResults() returned %d results, want 1", len(results))
	}
	if !results[0].Changed {
		t.Fatalf("RunWithResults() expected changed result, got %+v", results[0])
	}

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
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("RunWithResults() should not print output, got %q", string(out))
	}
}

func TestRunner_Run_Glob(t *testing.T) {
	tmpDir := t.TempDir()
	writeFile(t, filepath.Join(tmpDir, "test1.json"), []byte(`{"version": "1.0.0"}`))
	writeFile(t, filepath.Join(tmpDir, "test2.json"), []byte(`{"version": "1.0.0"}`))

	t.Run("glob pattern", func(t *testing.T) {
		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			Glob:      "*.json",
			Roots:     []string{tmpDir},
			DryRun:    true,
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err != nil {
			t.Errorf("Run() error = %v", err)
		}
	})

	t.Run("glob no matches", func(t *testing.T) {
		config := &Config{
			Operation: OpBump,
			BumpType:  "patch",
			Path:      []string{"version"},
			Glob:      "*.nomatch",
			Roots:     []string{tmpDir},
		}
		runner := NewRunner(config)
		err := runner.Run()
		if err == nil {
			t.Error("Run() should error on no matches")
		}
	})
}

func TestJoinPath(t *testing.T) {
	result := JoinPath([]string{"app", "version"})
	if result != "app.version" {
		t.Errorf("JoinPath() = %v, want app.version", result)
	}
}

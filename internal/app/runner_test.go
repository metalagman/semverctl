package app

import (
	"os"
	"path/filepath"
	"testing"
)

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
	os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644)

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
		os.WriteFile(numericFile, []byte(`{"count": 42}`), 0o644)

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
		os.WriteFile(invalidFile, []byte(`{invalid`), 0o644)

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
		os.WriteFile(invalidVerFile, []byte(`{"version": "not-a-version"}`), 0o644)

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

func TestRunner_Run_Glob(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test1.json"), []byte(`{"version": "1.0.0"}`), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "test2.json"), []byte(`{"version": "1.0.0"}`), 0o644)

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

package targets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewResolver(t *testing.T) {
	r := NewResolver(nil)
	if len(r.roots) != 1 || r.roots[0] != "." {
		t.Error("NewResolver(nil) should default to [\".\"]")
	}

	r = NewResolver([]string{"a", "b"})
	if len(r.roots) != 2 {
		t.Error("NewResolver should use provided roots")
	}
}

func TestResolver_Resolve(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	txtFile := filepath.Join(tmpDir, "test.txt")

	os.WriteFile(jsonFile, []byte(`{"version": "1.0.0"}`), 0644)
	os.WriteFile(txtFile, []byte("version=1.0.0"), 0644)

	r := NewResolver([]string{tmpDir})

	t.Run("resolve json file", func(t *testing.T) {
		files, err := r.Resolve(jsonFile)
		if err != nil {
			t.Errorf("Resolve() error = %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Resolve() returned %d files, want 1", len(files))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := r.Resolve(filepath.Join(tmpDir, "missing.json"))
		if err == nil {
			t.Error("Resolve() should error on missing file")
		}
	})

	t.Run("unsupported extension", func(t *testing.T) {
		_, err := r.Resolve(txtFile)
		if err == nil {
			t.Error("Resolve() should error on unsupported extension")
		}
	})

	t.Run("directory", func(t *testing.T) {
		_, err := r.Resolve(tmpDir)
		if err == nil {
			t.Error("Resolve() should error on directory")
		}
	})
}

func TestResolver_ResolveGlob(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte(`{"version": "1.0.0"}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.yaml"), []byte("version: 1.0.0\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("version=1.0.0"), 0644)

	r := NewResolver([]string{tmpDir})

	t.Run("glob json files", func(t *testing.T) {
		files, err := r.ResolveGlob("*.json")
		if err != nil {
			t.Errorf("ResolveGlob() error = %v", err)
		}
		if len(files) != 1 {
			t.Errorf("ResolveGlob() returned %d files, want 1", len(files))
		}
	})

	t.Run("glob all supported", func(t *testing.T) {
		files, err := r.ResolveGlob("*")
		if err != nil {
			t.Errorf("ResolveGlob() error = %v", err)
		}
		if len(files) != 2 { // json and yaml, not txt
			t.Errorf("ResolveGlob() returned %d files, want 2", len(files))
		}
	})

	t.Run("glob no matches", func(t *testing.T) {
		_, err := r.ResolveGlob("*.nomatch")
		if err == nil {
			t.Error("ResolveGlob() should error on no matches")
		}
	})

	t.Run("invalid glob pattern", func(t *testing.T) {
		_, err := r.ResolveGlob("[invalid")
		if err == nil {
			t.Error("ResolveGlob() should error on invalid pattern")
		}
	})
}

func TestIsSupportedExtension(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"test.json", true},
		{"test.yaml", true},
		{"test.yml", true},
		{"test.JSON", true}, // case insensitive
		{"test.txt", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isSupportedExtension(tt.path); got != tt.want {
				t.Errorf("isSupportedExtension(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidateSingleTarget(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		glob    string
		wantErr bool
	}{
		{"only file", "test.json", "", false},
		{"only glob", "", "*.json", false},
		{"both", "test.json", "*.json", true},
		{"neither", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSingleTarget(tt.file, tt.glob)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSingleTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

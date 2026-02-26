package formats

import (
	"math"
	"testing"
)

func TestGetCodec(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"json", "test.json", false},
		{"yaml", "test.yaml", false},
		{"yml", "test.yml", false},
		{"unsupported", "test.txt", true},
		{"no extension", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetCodec(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCodec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONCodec(t *testing.T) {
	codec := &JSONCodec{}

	t.Run("Decode", func(t *testing.T) {
		data := []byte(`{"version": "1.0.0"}`)
		doc, err := codec.Decode(data)
		if err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if doc == nil {
			t.Error("Decode() returned nil")
		}
	})

	t.Run("Decode invalid", func(t *testing.T) {
		data := []byte(`{invalid`)
		_, err := codec.Decode(data)
		if err == nil {
			t.Error("Decode() should error on invalid JSON")
		}
	})

	t.Run("Encode", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		encoded, err := codec.Encode(data)
		if err != nil {
			t.Errorf("Encode() error = %v", err)
		}
		if len(encoded) == 0 {
			t.Error("Encode() returned empty")
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		ver, err := codec.GetVersion(data, []string{"version"})
		if err != nil {
			t.Errorf("GetVersion() error = %v", err)
		}
		if ver != "1.0.0" {
			t.Errorf("GetVersion() = %v, want 1.0.0", ver)
		}
	})

	t.Run("GetVersion not string", func(t *testing.T) {
		data := map[string]any{"version": 123}
		_, err := codec.GetVersion(data, []string{"version"})
		if err == nil {
			t.Error("GetVersion() should error on non-string")
		}
	})

	t.Run("SetVersion", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		err := codec.SetVersion(data, []string{"version"}, "2.0.0")
		if err != nil {
			t.Errorf("SetVersion() error = %v", err)
		}
		if data["version"] != "2.0.0" {
			t.Errorf("SetVersion() didn't update, got %v", data["version"])
		}
	})

	t.Run("GetNumericScalar", func(t *testing.T) {
		data := map[string]any{"count": float64(42)}
		val, err := codec.GetNumericScalar(data, []string{"count"})
		if err != nil {
			t.Errorf("GetNumericScalar() error = %v", err)
		}
		if val != 42 {
			t.Errorf("GetNumericScalar() = %v, want 42", val)
		}
	})

	t.Run("GetNumericScalar not numeric", func(t *testing.T) {
		data := map[string]any{"count": "not a number"}
		_, err := codec.GetNumericScalar(data, []string{"count"})
		if err == nil {
			t.Error("GetNumericScalar() should error on non-numeric")
		}
	})

	t.Run("GetNumericScalar non-integer float", func(t *testing.T) {
		data := map[string]any{"count": 42.5}
		_, err := codec.GetNumericScalar(data, []string{"count"})
		if err == nil {
			t.Error("GetNumericScalar() should error on non-integer float")
		}
	})

	t.Run("GetNumericScalar non-finite", func(t *testing.T) {
		data := map[string]any{"count": math.Inf(1)}
		_, err := codec.GetNumericScalar(data, []string{"count"})
		if err == nil {
			t.Error("GetNumericScalar() should error on infinity")
		}
	})

	t.Run("SetNumericScalar", func(t *testing.T) {
		data := map[string]any{"count": float64(42)}
		err := codec.SetNumericScalar(data, []string{"count"}, 43)
		if err != nil {
			t.Errorf("SetNumericScalar() error = %v", err)
		}
	})
}

func TestYAMLCodec(t *testing.T) {
	codec := &YAMLCodec{}

	t.Run("Decode", func(t *testing.T) {
		data := []byte("version: 1.0.0\n")
		doc, err := codec.Decode(data)
		if err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if doc == nil {
			t.Error("Decode() returned nil")
		}
	})

	t.Run("Decode invalid", func(t *testing.T) {
		data := []byte("\t\t: invalid") // Invalid YAML
		_, err := codec.Decode(data)
		if err == nil {
			t.Error("Decode() should error on invalid YAML")
		}
	})

	t.Run("Encode", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		encoded, err := codec.Encode(data)
		if err != nil {
			t.Errorf("Encode() error = %v", err)
		}
		if len(encoded) == 0 {
			t.Error("Encode() returned empty")
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		ver, err := codec.GetVersion(data, []string{"version"})
		if err != nil {
			t.Errorf("GetVersion() error = %v", err)
		}
		if ver != "1.0.0" {
			t.Errorf("GetVersion() = %v, want 1.0.0", ver)
		}
	})

	t.Run("GetVersion not string", func(t *testing.T) {
		data := map[string]any{"version": 123}
		_, err := codec.GetVersion(data, []string{"version"})
		if err == nil {
			t.Error("GetVersion() should error on non-string")
		}
	})

	t.Run("SetVersion", func(t *testing.T) {
		data := map[string]any{"version": "1.0.0"}
		err := codec.SetVersion(data, []string{"version"}, "2.0.0")
		if err != nil {
			t.Errorf("SetVersion() error = %v", err)
		}
		if data["version"] != "2.0.0" {
			t.Errorf("SetVersion() didn't update, got %v", data["version"])
		}
	})

	t.Run("GetNumericScalar", func(t *testing.T) {
		data := map[string]any{"count": 42}
		val, err := codec.GetNumericScalar(data, []string{"count"})
		if err != nil {
			t.Errorf("GetNumericScalar() error = %v", err)
		}
		if val != 42 {
			t.Errorf("GetNumericScalar() = %v, want 42", val)
		}
	})

	t.Run("GetNumericScalar non-integer float", func(t *testing.T) {
		data := map[string]any{"count": 42.5}
		_, err := codec.GetNumericScalar(data, []string{"count"})
		if err == nil {
			t.Error("GetNumericScalar() should error on non-integer float")
		}
	})

	t.Run("GetNumericScalar non-finite", func(t *testing.T) {
		data := map[string]any{"count": math.NaN()}
		_, err := codec.GetNumericScalar(data, []string{"count"})
		if err == nil {
			t.Error("GetNumericScalar() should error on NaN")
		}
	})

	t.Run("SetNumericScalar", func(t *testing.T) {
		data := map[string]any{"count": 42}
		err := codec.SetNumericScalar(data, []string{"count"}, 43)
		if err != nil {
			t.Errorf("SetNumericScalar() error = %v", err)
		}
	})
}

func TestNormalizeYAML(t *testing.T) {
	tests := []struct {
		name string
		val  any
	}{
		{"string map", map[string]any{"key": "value"}},
		{"interface map", map[any]any{"key": "value"}},
		{"slice", []any{"a", "b"}},
		{"string", "value"},
		{"int", 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeYAML(tt.val)
			if result == nil && tt.val != nil {
				t.Error("normalizeYAML() returned nil for non-nil input")
			}
		})
	}
}

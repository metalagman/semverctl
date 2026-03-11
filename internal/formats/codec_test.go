package formats

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
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

func TestJSONCodec_VersionByteExact(t *testing.T) {
	codec := &JSONCodec{}
	input := []byte("{\n  \"name\": \"demo\",\n  \"scripts\": {\"start\": \"node index.js\"},\n  \"version\": \"1.2.3\",\n  \"dependencies\": {\n    \"a\": \"1.0.0\"\n  }\n}\n")

	ver, err := codec.GetVersion(input, []string{"version"})
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}
	if ver != "1.2.3" {
		t.Fatalf("GetVersion() = %q, want 1.2.3", ver)
	}

	out, err := codec.SetVersion(input, []string{"version"}, "1.2.4")
	if err != nil {
		t.Fatalf("SetVersion() error = %v", err)
	}

	want := []byte(strings.Replace(string(input), `"1.2.3"`, `"1.2.4"`, 1))
	if !bytes.Equal(out, want) {
		t.Fatalf("SetVersion() changed unexpected bytes\nwant:\n%s\n got:\n%s", string(want), string(out))
	}
}

func TestJSONCodec_NumericByteExact(t *testing.T) {
	codec := &JSONCodec{}
	input := []byte("{\n  \"meta\": {\n    \"count\"   :   42\n  }\n}\n")

	val, err := codec.GetNumericScalar(input, []string{"meta", "count"})
	if err != nil {
		t.Fatalf("GetNumericScalar() error = %v", err)
	}
	if val != 42 {
		t.Fatalf("GetNumericScalar() = %d, want 42", val)
	}

	out, err := codec.SetNumericScalar(input, []string{"meta", "count"}, 43)
	if err != nil {
		t.Fatalf("SetNumericScalar() error = %v", err)
	}

	want := []byte(strings.Replace(string(input), "42", "43", 1))
	if !bytes.Equal(out, want) {
		t.Fatalf("SetNumericScalar() changed unexpected bytes\nwant:\n%s\n got:\n%s", string(want), string(out))
	}
}

func TestJSONCodec_Errors(t *testing.T) {
	codec := &JSONCodec{}

	t.Run("invalid JSON", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte(`{invalid`), []string{"version"}); err == nil {
			t.Fatal("GetVersion() should error on invalid JSON")
		}
	})

	t.Run("path not found", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte(`{"version":"1.0.0"}`), []string{"missing"}); err == nil {
			t.Fatal("GetVersion() should error on missing path")
		}
	})

	t.Run("version not string", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte(`{"version":123}`), []string{"version"}); err == nil {
			t.Fatal("GetVersion() should error on non-string")
		}
	})

	t.Run("numeric not numeric", func(t *testing.T) {
		if _, err := codec.GetNumericScalar([]byte(`{"count":"forty-two"}`), []string{"count"}); err == nil {
			t.Fatal("GetNumericScalar() should error on non-numeric")
		}
	})

	t.Run("numeric non integer", func(t *testing.T) {
		if _, err := codec.GetNumericScalar([]byte(`{"count":42.5}`), []string{"count"}); err == nil {
			t.Fatal("GetNumericScalar() should error on non-integer")
		}
	})
}

func TestYAMLCodec_VersionByteExact(t *testing.T) {
	codec := &YAMLCodec{}
	input := []byte("# chart metadata\napiVersion: v2\nname: demo\n# semver for chart\nversion: 1.2.3\nappVersion: \"1.2.3\"\n")

	ver, err := codec.GetVersion(input, []string{"version"})
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}
	if ver != "1.2.3" {
		t.Fatalf("GetVersion() = %q, want 1.2.3", ver)
	}

	out, err := codec.SetVersion(input, []string{"version"}, "1.2.4")
	if err != nil {
		t.Fatalf("SetVersion() error = %v", err)
	}

	want := []byte(strings.Replace(string(input), "version: 1.2.3", "version: 1.2.4", 1))
	if !bytes.Equal(out, want) {
		t.Fatalf("SetVersion() changed unexpected bytes\nwant:\n%s\n got:\n%s", string(want), string(out))
	}
}

func TestYAMLCodec_PreservesQuoteStyle(t *testing.T) {
	codec := &YAMLCodec{}

	t.Run("double quoted", func(t *testing.T) {
		input := []byte("version: \"1.2.3\"\n")
		out, err := codec.SetVersion(input, []string{"version"}, "2.0.0")
		if err != nil {
			t.Fatalf("SetVersion() error = %v", err)
		}
		want := []byte("version: \"2.0.0\"\n")
		if !bytes.Equal(out, want) {
			t.Fatalf("unexpected output\nwant:\n%s\n got:\n%s", string(want), string(out))
		}
	})

	t.Run("single quoted", func(t *testing.T) {
		input := []byte("version: '1.2.3'\n")
		out, err := codec.SetVersion(input, []string{"version"}, "2.0.0")
		if err != nil {
			t.Fatalf("SetVersion() error = %v", err)
		}
		want := []byte("version: '2.0.0'\n")
		if !bytes.Equal(out, want) {
			t.Fatalf("unexpected output\nwant:\n%s\n got:\n%s", string(want), string(out))
		}
	})
}

func TestYAMLCodec_NumericByteExact(t *testing.T) {
	codec := &YAMLCodec{}
	input := []byte("count: 42   # keep comment\n")

	val, err := codec.GetNumericScalar(input, []string{"count"})
	if err != nil {
		t.Fatalf("GetNumericScalar() error = %v", err)
	}
	if val != 42 {
		t.Fatalf("GetNumericScalar() = %d, want 42", val)
	}

	out, err := codec.SetNumericScalar(input, []string{"count"}, 43)
	if err != nil {
		t.Fatalf("SetNumericScalar() error = %v", err)
	}

	want := []byte("count: 43   # keep comment\n")
	if !bytes.Equal(out, want) {
		t.Fatalf("unexpected output\nwant:\n%s\n got:\n%s", string(want), string(out))
	}
}

func TestYAMLCodec_UnsupportedStyle(t *testing.T) {
	codec := &YAMLCodec{}
	input := []byte("version: |\n  1.2.3\n")

	if _, err := codec.SetVersion(input, []string{"version"}, "1.2.4"); err == nil {
		t.Fatal("SetVersion() should error for multiline scalar style")
	}
}

func TestYAMLCodec_Errors(t *testing.T) {
	codec := &YAMLCodec{}

	t.Run("path not found", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte("version: 1.0.0\n"), []string{"missing"}); err == nil {
			t.Fatal("GetVersion() should error on missing path")
		}
	})

	t.Run("version not string", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte("version: 123\n"), []string{"version"}); err == nil {
			t.Fatal("GetVersion() should error on non-string")
		}
	})

	t.Run("numeric non integer", func(t *testing.T) {
		if _, err := codec.GetNumericScalar([]byte("count: 42.5\n"), []string{"count"}); err == nil {
			t.Fatal("GetNumericScalar() should error on non-integer")
		}
	})

	t.Run("cannot traverse scalar", func(t *testing.T) {
		if _, err := codec.GetVersion([]byte("version: 1.0.0\n"), []string{"version", "patch"}); err == nil {
			t.Fatal("GetVersion() should error when traversing scalar")
		}
	})
}

func TestNumericScalarToInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    int64
		wantErr bool
	}{
		{name: "int", input: int(42), want: 42},
		{name: "uint32", input: uint32(42), want: 42},
		{name: "float64 integer", input: float64(42), want: 42},
		{name: "json number int", input: json.Number("42"), want: 42},
		{name: "float64 non integer", input: 42.5, wantErr: true},
		{name: "float64 inf", input: math.Inf(1), wantErr: true},
		{name: "json number non integer", input: json.Number("42.5"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := numericScalarToInt64(tt.input, []string{"value"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("numericScalarToInt64() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("numericScalarToInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

package formats

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestJSONParserHelpers(t *testing.T) {
	t.Run("parseJSONValueEnd handles all value kinds", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{name: "object", input: `{"a":1}`},
			{name: "array", input: `[1,{"a":2},true,false,null]`},
			{name: "string", input: `"a\\tb"`},
			{name: "true", input: `true`},
			{name: "false", input: `false`},
			{name: "null", input: `null`},
			{name: "number", input: `-12.34e+5`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				end, err := parseJSONValueEnd([]byte(tt.input), 0)
				if err != nil {
					t.Fatalf("parseJSONValueEnd() error = %v", err)
				}
				if end != len(tt.input) {
					t.Fatalf("parseJSONValueEnd() end = %d, want %d", end, len(tt.input))
				}
			})
		}
	})

	t.Run("parseJSONValueEnd invalid", func(t *testing.T) {
		if _, err := parseJSONValueEnd([]byte("?"), 0); err == nil {
			t.Fatal("parseJSONValueEnd() should error for invalid value")
		}
	})

	t.Run("parseJSONArrayEnd errors", func(t *testing.T) {
		if _, err := parseJSONArrayEnd([]byte("[1 2]"), 0); err == nil {
			t.Fatal("parseJSONArrayEnd() should error when comma is missing")
		}
		if _, err := parseJSONArrayEnd([]byte("[1,"), 0); err == nil {
			t.Fatal("parseJSONArrayEnd() should error on unexpected end")
		}
	})

	t.Run("parseJSONObjectEnd errors", func(t *testing.T) {
		if _, err := parseJSONObjectEnd([]byte(`{"a" 1}`), 0); err == nil {
			t.Fatal("parseJSONObjectEnd() should error when ':' is missing")
		}
		if _, err := parseJSONObjectEnd([]byte(`{"a":1 "b":2}`), 0); err == nil {
			t.Fatal("parseJSONObjectEnd() should error when comma is missing")
		}
	})

	t.Run("parseJSONString errors", func(t *testing.T) {
		if _, _, err := parseJSONString([]byte("abc"), 0); err == nil {
			t.Fatal("parseJSONString() should error when not starting with quote")
		}
		if _, _, err := parseJSONString([]byte(`"bad\q"`), 0); err == nil {
			t.Fatal("parseJSONString() should error for invalid escape")
		}
		if _, _, err := parseJSONString([]byte("\"bad\n"), 0); err == nil {
			t.Fatal("parseJSONString() should error for newline before closing quote")
		}
		if _, _, err := parseJSONString([]byte("\"bad"), 0); err == nil {
			t.Fatal("parseJSONString() should error for unterminated string")
		}
	})

	t.Run("parseJSONLiteralEnd", func(t *testing.T) {
		end, err := parseJSONLiteralEnd([]byte("true"), 0, "true")
		if err != nil {
			t.Fatalf("parseJSONLiteralEnd() error = %v", err)
		}
		if end != 4 {
			t.Fatalf("parseJSONLiteralEnd() end = %d, want 4", end)
		}
		if _, err := parseJSONLiteralEnd([]byte("tru"), 0, "true"); err == nil {
			t.Fatal("parseJSONLiteralEnd() should error on invalid literal")
		}
	})

	t.Run("parseJSONNumberEnd", func(t *testing.T) {
		end, err := parseJSONNumberEnd([]byte("123"), 0)
		if err != nil {
			t.Fatalf("parseJSONNumberEnd() error = %v", err)
		}
		if end != 3 {
			t.Fatalf("parseJSONNumberEnd() end = %d, want 3", end)
		}
		if _, err := parseJSONNumberEnd([]byte("-"), 0); err == nil {
			t.Fatal("parseJSONNumberEnd() should error on '-' only")
		}
		if _, err := parseJSONNumberEnd([]byte("1e+"), 0); err == nil {
			t.Fatal("parseJSONNumberEnd() should error on invalid exponent")
		}
		if _, err := parseJSONNumberEnd([]byte("1."), 0); err == nil {
			t.Fatal("parseJSONNumberEnd() should error on missing fraction digits")
		}
	})

	t.Run("findJSONPathInValue type traversal error", func(t *testing.T) {
		_, _, err := findJSONValueSpan([]byte(`{"a":true}`), []string{"a", "b"})
		if err == nil {
			t.Fatal("findJSONValueSpan() should error when traversing non-object")
		}
		if !strings.Contains(err.Error(), "cannot traverse bool") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("findJSONValueSpan root and empty", func(t *testing.T) {
		start, end, err := findJSONValueSpan([]byte(" \n\t[1,2]"), nil)
		if err != nil {
			t.Fatalf("findJSONValueSpan() error = %v", err)
		}
		if got := string([]byte(" \n\t[1,2]")[start:end]); got != "[1,2]" {
			t.Fatalf("span mismatch = %q", got)
		}

		if _, _, err := findJSONValueSpan([]byte("   \n\t"), []string{"a"}); err == nil {
			t.Fatal("findJSONValueSpan() should error for empty JSON input")
		}
	})

	t.Run("jsonTypeName", func(t *testing.T) {
		tests := map[byte]string{
			'{': "map[string]interface {}",
			'[': "[]interface {}",
			'"': "string",
			't': "bool",
			'n': "<nil>",
			'-': "float64",
			'9': "float64",
			'?': "interface {}",
		}
		for b, want := range tests {
			if got := jsonTypeName(b); got != want {
				t.Fatalf("jsonTypeName(%q) = %q, want %q", b, got, want)
			}
		}
	})
}

func TestYAMLHelpers(t *testing.T) {
	t.Run("yamlTypeName", func(t *testing.T) {
		tests := []struct {
			name string
			node *yaml.Node
			want string
		}{
			{name: "nil", node: nil, want: "interface {}"},
			{name: "map", node: &yaml.Node{Kind: yaml.MappingNode}, want: "map[string]interface {}"},
			{name: "sequence", node: &yaml.Node{Kind: yaml.SequenceNode}, want: "[]interface {}"},
			{name: "string", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str"}, want: "string"},
			{name: "int", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int"}, want: "int"},
			{name: "float", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float"}, want: "float64"},
			{name: "bool", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool"}, want: "bool"},
			{name: "null", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"}, want: "<nil>"},
			{name: "unknown", node: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!binary"}, want: "interface {}"},
			{name: "other kind", node: &yaml.Node{Kind: yaml.DocumentNode}, want: "interface {}"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := yamlTypeName(tt.node); got != tt.want {
					t.Fatalf("yamlTypeName() = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("yamlScalarSpan errors", func(t *testing.T) {
		path := []string{"version"}
		if _, _, err := yamlScalarSpan([]byte("version: 1.0.0\n"), nil, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for nil location")
		}

		badLoc := &yamlPathLocation{node: &yaml.Node{Kind: yaml.MappingNode, Line: 1, Column: 1}}
		if _, _, err := yamlScalarSpan([]byte("version: 1.0.0\n"), badLoc, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for non-scalar node")
		}

		flowParent := &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}
		flowLoc := &yamlPathLocation{node: &yaml.Node{Kind: yaml.ScalarNode, Line: 1, Column: 1}, parent: flowParent}
		if _, _, err := yamlScalarSpan([]byte("version: 1.0.0\n"), flowLoc, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for flow-style parent")
		}

		literalLoc := &yamlPathLocation{node: &yaml.Node{Kind: yaml.ScalarNode, Style: yaml.LiteralStyle, Line: 1, Column: 1}}
		if _, _, err := yamlScalarSpan([]byte("version: |\n  1.0.0\n"), literalLoc, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for literal style")
		}

		multilineContentLoc := &yamlPathLocation{node: &yaml.Node{Kind: yaml.ScalarNode, Value: "a\nb", Line: 1, Column: 1}}
		if _, _, err := yamlScalarSpan([]byte("version: 1.0.0\n"), multilineContentLoc, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for multiline value content")
		}

		noPosLoc := &yamlPathLocation{node: &yaml.Node{Kind: yaml.ScalarNode, Line: 0, Column: 0}}
		if _, _, err := yamlScalarSpan([]byte("version: 1.0.0\n"), noPosLoc, path); err == nil {
			t.Fatal("yamlScalarSpan() should error for missing source position")
		}
	})

	t.Run("lineColToOffset", func(t *testing.T) {
		data := []byte("a\nβ\n")
		offset, err := lineColToOffset(data, 2, 1)
		if err != nil {
			t.Fatalf("lineColToOffset() error = %v", err)
		}
		if string(data[offset:offset+2]) != "β" {
			t.Fatalf("unexpected offset %d for utf8 rune", offset)
		}

		if _, err := lineColToOffset(data, 0, 1); err == nil {
			t.Fatal("lineColToOffset() should error on invalid target position")
		}
		if _, err := lineColToOffset([]byte{0xff}, 1, 2); err == nil {
			t.Fatal("lineColToOffset() should error on invalid UTF-8")
		}
		if _, err := lineColToOffset([]byte("a\n"), 3, 1); err == nil {
			t.Fatal("lineColToOffset() should error when target is out of range")
		}
	})

	t.Run("quoted scalar end scanners", func(t *testing.T) {
		end, err := findYAMLDoubleQuotedEnd([]byte(`"a\"b"`), 0)
		if err != nil {
			t.Fatalf("findYAMLDoubleQuotedEnd() error = %v", err)
		}
		if end != len(`"a\"b"`) {
			t.Fatalf("findYAMLDoubleQuotedEnd() end = %d", end)
		}
		if _, err := findYAMLDoubleQuotedEnd([]byte("abc"), 0); err == nil {
			t.Fatal("findYAMLDoubleQuotedEnd() should error when not quoted")
		}
		if _, err := findYAMLDoubleQuotedEnd([]byte{'"', '\\'}, 0); err == nil {
			t.Fatal("findYAMLDoubleQuotedEnd() should error on unterminated escape")
		}
		if _, err := findYAMLDoubleQuotedEnd([]byte("\"a\n"), 0); err == nil {
			t.Fatal("findYAMLDoubleQuotedEnd() should error on newline")
		}
		if _, err := findYAMLDoubleQuotedEnd([]byte("\"a"), 0); err == nil {
			t.Fatal("findYAMLDoubleQuotedEnd() should error on unterminated quote")
		}

		end, err = findYAMLSingleQuotedEnd([]byte("'a''b'"), 0)
		if err != nil {
			t.Fatalf("findYAMLSingleQuotedEnd() error = %v", err)
		}
		if end != len("'a''b'") {
			t.Fatalf("findYAMLSingleQuotedEnd() end = %d", end)
		}
		if _, err := findYAMLSingleQuotedEnd([]byte("abc"), 0); err == nil {
			t.Fatal("findYAMLSingleQuotedEnd() should error when not quoted")
		}
		if _, err := findYAMLSingleQuotedEnd([]byte("'a\n"), 0); err == nil {
			t.Fatal("findYAMLSingleQuotedEnd() should error on newline")
		}
		if _, err := findYAMLSingleQuotedEnd([]byte("'a"), 0); err == nil {
			t.Fatal("findYAMLSingleQuotedEnd() should error on unterminated quote")
		}
	})

	t.Run("yamlNumericLiteral styles", func(t *testing.T) {
		got, err := yamlNumericLiteral(42, yaml.DoubleQuotedStyle, []string{"count"})
		if err != nil || string(got) != `"42"` {
			t.Fatalf("yamlNumericLiteral double-quoted got=%q err=%v", string(got), err)
		}

		got, err = yamlNumericLiteral(42, yaml.SingleQuotedStyle, []string{"count"})
		if err != nil || string(got) != `'42'` {
			t.Fatalf("yamlNumericLiteral single-quoted got=%q err=%v", string(got), err)
		}

		got, err = yamlNumericLiteral(42, 0, []string{"count"})
		if err != nil || string(got) != `42` {
			t.Fatalf("yamlNumericLiteral plain got=%q err=%v", string(got), err)
		}

		if _, err := yamlNumericLiteral(42, yaml.LiteralStyle, []string{"count"}); err == nil {
			t.Fatal("yamlNumericLiteral should error for literal style")
		}
	})
}

func TestNumericScalarToInt64_AllBranches(t *testing.T) {
	tests := []struct {
		name    string
		val     any
		want    int64
		wantErr bool
	}{
		{name: "int8", val: int8(7), want: 7},
		{name: "int16", val: int16(8), want: 8},
		{name: "int32", val: int32(9), want: 9},
		{name: "uint", val: uint(10), want: 10},
		{name: "uint8", val: uint8(11), want: 11},
		{name: "uint16", val: uint16(12), want: 12},
		{name: "uint32", val: uint32(13), want: 13},
		{name: "uint64", val: uint64(14), want: 14},
		{name: "uint64 overflow", val: uint64(math.MaxInt64) + 1, wantErr: true},
		{name: "float32", val: float32(15), want: 15},
		{name: "float32 inf", val: float32(math.Inf(1)), wantErr: true},
		{name: "float32 nan", val: float32(math.NaN()), wantErr: true},
		{name: "json number invalid", val: json.Number("abc"), wantErr: true},
		{name: "json number fractional", val: json.Number("1.5"), wantErr: true},
		{name: "unsupported", val: struct{}{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := numericScalarToInt64(tt.val, []string{"x"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("numericScalarToInt64() error=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("numericScalarToInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

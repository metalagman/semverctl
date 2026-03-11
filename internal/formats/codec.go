package formats

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"github.com/metalagman/semverctl/internal/pathx"
)

// Codec handles encoding/decoding of specific file formats.
type Codec interface {
	GetVersion(data []byte, path []string) (string, error)
	SetVersion(data []byte, path []string, version string) ([]byte, error)
	GetNumericScalar(data []byte, path []string) (int64, error)
	SetNumericScalar(data []byte, path []string, value int64) ([]byte, error)
}

// GetCodec returns the appropriate codec for a file based on extension.
func GetCodec(filename string) (Codec, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return &JSONCodec{}, nil
	case ".yaml", ".yml":
		return &YAMLCodec{}, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// JSONCodec handles JSON files.
type JSONCodec struct{}

// GetVersion retrieves a version string at the given path.
func (c *JSONCodec) GetVersion(data []byte, path []string) (string, error) {
	start, end, err := findJSONValueSpan(data, path)
	if err != nil {
		return "", err
	}

	val, err := decodeJSONValue(data[start:end])
	if err != nil {
		return "", err
	}

	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("version at path %s is not a string (got %T)", pathx.Join(path), val)
	}
}

// SetVersion updates a version string at the given path.
func (c *JSONCodec) SetVersion(data []byte, path []string, version string) ([]byte, error) {
	start, end, err := findJSONValueSpan(data, path)
	if err != nil {
		return nil, err
	}

	replacement, err := json.Marshal(version)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JSON string: %w", err)
	}

	return replaceSpan(data, start, end, replacement), nil
}

// GetNumericScalar retrieves a numeric value at the given path.
func (c *JSONCodec) GetNumericScalar(data []byte, path []string) (int64, error) {
	start, end, err := findJSONValueSpan(data, path)
	if err != nil {
		return 0, err
	}

	val, err := decodeJSONValue(data[start:end])
	if err != nil {
		return 0, err
	}

	return numericScalarToInt64(val, path)
}

// SetNumericScalar updates a numeric value at the given path.
func (c *JSONCodec) SetNumericScalar(data []byte, path []string, value int64) ([]byte, error) {
	start, end, err := findJSONValueSpan(data, path)
	if err != nil {
		return nil, err
	}

	return replaceSpan(data, start, end, []byte(strconv.FormatInt(value, 10))), nil
}

// YAMLCodec handles YAML files.
type YAMLCodec struct{}

// GetVersion retrieves a version string at the given path.
func (c *YAMLCodec) GetVersion(data []byte, path []string) (string, error) {
	loc, err := findYAMLPath(data, path)
	if err != nil {
		return "", err
	}

	var val any
	if err := loc.node.Decode(&val); err != nil {
		return "", fmt.Errorf("failed to decode value at path %s: %w", pathx.Join(path), err)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("version at path %s is not a string (got %T)", pathx.Join(path), val)
	}
}

// SetVersion updates a version string at the given path.
func (c *YAMLCodec) SetVersion(data []byte, path []string, version string) ([]byte, error) {
	loc, err := findYAMLPath(data, path)
	if err != nil {
		return nil, err
	}

	start, end, err := yamlScalarSpan(data, loc, path)
	if err != nil {
		return nil, err
	}

	replacement, err := yamlStringLiteral(version, loc.node.Style, path)
	if err != nil {
		return nil, err
	}

	return replaceSpan(data, start, end, replacement), nil
}

// GetNumericScalar retrieves a numeric value at the given path.
func (c *YAMLCodec) GetNumericScalar(data []byte, path []string) (int64, error) {
	loc, err := findYAMLPath(data, path)
	if err != nil {
		return 0, err
	}

	var val any
	if err := loc.node.Decode(&val); err != nil {
		return 0, fmt.Errorf("failed to decode value at path %s: %w", pathx.Join(path), err)
	}

	return numericScalarToInt64(val, path)
}

// SetNumericScalar updates a numeric value at the given path.
func (c *YAMLCodec) SetNumericScalar(data []byte, path []string, value int64) ([]byte, error) {
	loc, err := findYAMLPath(data, path)
	if err != nil {
		return nil, err
	}

	start, end, err := yamlScalarSpan(data, loc, path)
	if err != nil {
		return nil, err
	}

	replacement, err := yamlNumericLiteral(value, loc.node.Style, path)
	if err != nil {
		return nil, err
	}

	return replaceSpan(data, start, end, replacement), nil
}

func decodeJSONValue(raw []byte) (any, error) {
	var val any
	if err := json.Unmarshal(raw, &val); err != nil {
		return nil, fmt.Errorf("failed to decode JSON value: %w", err)
	}
	return val, nil
}

func findJSONValueSpan(data []byte, path []string) (int, int, error) {
	start := skipJSONWhitespace(data, 0)
	if start >= len(data) {
		return 0, 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}

	if len(path) == 0 {
		end, err := parseJSONValueEnd(data, start)
		if err != nil {
			return 0, 0, err
		}
		return start, end, nil
	}

	return findJSONPathInValue(data, start, path, 0)
}

func findJSONPathInValue(data []byte, valueStart int, path []string, depth int) (int, int, error) {
	valueStart = skipJSONWhitespace(data, valueStart)
	if valueStart >= len(data) {
		return 0, 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}

	if depth == len(path) {
		end, err := parseJSONValueEnd(data, valueStart)
		if err != nil {
			return 0, 0, err
		}
		return valueStart, end, nil
	}

	if data[valueStart] != '{' {
		return 0, 0, fmt.Errorf("cannot traverse %s at path %s", jsonTypeName(data[valueStart]), pathx.Join(path[:depth]))
	}

	idx := skipJSONWhitespace(data, valueStart+1)
	if idx >= len(data) {
		return 0, 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}
	if data[idx] == '}' {
		return 0, 0, fmt.Errorf("path not found: %s", pathx.Join(path[:depth+1]))
	}

	targetKey := path[depth]
	for {
		if idx >= len(data) || data[idx] != '"' {
			return 0, 0, fmt.Errorf("failed to decode JSON: expected object key")
		}

		keyEnd, key, err := parseJSONString(data, idx)
		if err != nil {
			return 0, 0, err
		}

		idx = skipJSONWhitespace(data, keyEnd)
		if idx >= len(data) || data[idx] != ':' {
			return 0, 0, fmt.Errorf("failed to decode JSON: expected ':' after object key")
		}
		idx++

		childStart := skipJSONWhitespace(data, idx)
		if childStart >= len(data) {
			return 0, 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
		}

		if key == targetKey {
			if depth == len(path)-1 {
				childEnd, err := parseJSONValueEnd(data, childStart)
				if err != nil {
					return 0, 0, err
				}
				return childStart, childEnd, nil
			}
			return findJSONPathInValue(data, childStart, path, depth+1)
		}

		skippedEnd, err := parseJSONValueEnd(data, childStart)
		if err != nil {
			return 0, 0, err
		}
		idx = skipJSONWhitespace(data, skippedEnd)
		if idx >= len(data) {
			return 0, 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
		}

		if data[idx] == ',' {
			idx = skipJSONWhitespace(data, idx+1)
			continue
		}
		if data[idx] == '}' {
			return 0, 0, fmt.Errorf("path not found: %s", pathx.Join(path[:depth+1]))
		}
		return 0, 0, fmt.Errorf("failed to decode JSON: expected ',' or '}'")
	}
}

func parseJSONValueEnd(data []byte, idx int) (int, error) {
	idx = skipJSONWhitespace(data, idx)
	if idx >= len(data) {
		return 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}

	switch data[idx] {
	case '{':
		return parseJSONObjectEnd(data, idx)
	case '[':
		return parseJSONArrayEnd(data, idx)
	case '"':
		end, _, err := parseJSONString(data, idx)
		return end, err
	case 't':
		return parseJSONLiteralEnd(data, idx, "true")
	case 'f':
		return parseJSONLiteralEnd(data, idx, "false")
	case 'n':
		return parseJSONLiteralEnd(data, idx, "null")
	default:
		if data[idx] == '-' || isDigit(data[idx]) {
			return parseJSONNumberEnd(data, idx)
		}
		return 0, fmt.Errorf("failed to decode JSON: invalid value")
	}
}

func parseJSONObjectEnd(data []byte, idx int) (int, error) {
	idx = skipJSONWhitespace(data, idx+1)
	if idx >= len(data) {
		return 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}
	if data[idx] == '}' {
		return idx + 1, nil
	}

	for {
		if idx >= len(data) || data[idx] != '"' {
			return 0, fmt.Errorf("failed to decode JSON: expected object key")
		}
		keyEnd, _, err := parseJSONString(data, idx)
		if err != nil {
			return 0, err
		}

		idx = skipJSONWhitespace(data, keyEnd)
		if idx >= len(data) || data[idx] != ':' {
			return 0, fmt.Errorf("failed to decode JSON: expected ':' after object key")
		}
		idx++

		valEnd, err := parseJSONValueEnd(data, idx)
		if err != nil {
			return 0, err
		}
		idx = skipJSONWhitespace(data, valEnd)
		if idx >= len(data) {
			return 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
		}

		if data[idx] == ',' {
			idx = skipJSONWhitespace(data, idx+1)
			continue
		}
		if data[idx] == '}' {
			return idx + 1, nil
		}
		return 0, fmt.Errorf("failed to decode JSON: expected ',' or '}'")
	}
}

func parseJSONArrayEnd(data []byte, idx int) (int, error) {
	idx = skipJSONWhitespace(data, idx+1)
	if idx >= len(data) {
		return 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
	}
	if data[idx] == ']' {
		return idx + 1, nil
	}

	for {
		valEnd, err := parseJSONValueEnd(data, idx)
		if err != nil {
			return 0, err
		}
		idx = skipJSONWhitespace(data, valEnd)
		if idx >= len(data) {
			return 0, fmt.Errorf("failed to decode JSON: unexpected end of input")
		}

		if data[idx] == ',' {
			idx = skipJSONWhitespace(data, idx+1)
			continue
		}
		if data[idx] == ']' {
			return idx + 1, nil
		}
		return 0, fmt.Errorf("failed to decode JSON: expected ',' or ']'")
	}
}

func parseJSONString(data []byte, idx int) (int, string, error) {
	if idx >= len(data) || data[idx] != '"' {
		return 0, "", fmt.Errorf("failed to decode JSON: expected string")
	}
	i := idx + 1
	for i < len(data) {
		switch data[i] {
		case '\\':
			i++
			if i >= len(data) {
				return 0, "", fmt.Errorf("failed to decode JSON: unterminated escape sequence")
			}
		case '"':
			raw := data[idx : i+1]
			unquoted, err := strconv.Unquote(string(raw))
			if err != nil {
				return 0, "", fmt.Errorf("failed to decode JSON string: %w", err)
			}
			return i + 1, unquoted, nil
		case '\n', '\r':
			return 0, "", fmt.Errorf("failed to decode JSON: unterminated string")
		}
		i++
	}
	return 0, "", fmt.Errorf("failed to decode JSON: unterminated string")
}

func parseJSONLiteralEnd(data []byte, idx int, literal string) (int, error) {
	if len(data[idx:]) < len(literal) || string(data[idx:idx+len(literal)]) != literal {
		return 0, fmt.Errorf("failed to decode JSON: invalid literal")
	}
	return idx + len(literal), nil
}

func parseJSONNumberEnd(data []byte, idx int) (int, error) {
	i := idx
	if data[i] == '-' {
		i++
		if i >= len(data) {
			return 0, fmt.Errorf("failed to decode JSON: invalid number")
		}
	}

	if data[i] == '0' {
		i++
	} else if isDigitOneToNine(data[i]) {
		for i < len(data) && isDigit(data[i]) {
			i++
		}
	} else {
		return 0, fmt.Errorf("failed to decode JSON: invalid number")
	}

	if i < len(data) && data[i] == '.' {
		i++
		if i >= len(data) || !isDigit(data[i]) {
			return 0, fmt.Errorf("failed to decode JSON: invalid number")
		}
		for i < len(data) && isDigit(data[i]) {
			i++
		}
	}

	if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
		i++
		if i < len(data) && (data[i] == '+' || data[i] == '-') {
			i++
		}
		if i >= len(data) || !isDigit(data[i]) {
			return 0, fmt.Errorf("failed to decode JSON: invalid number")
		}
		for i < len(data) && isDigit(data[i]) {
			i++
		}
	}

	return i, nil
}

func skipJSONWhitespace(data []byte, idx int) int {
	for idx < len(data) {
		switch data[idx] {
		case ' ', '\t', '\n', '\r':
			idx++
		default:
			return idx
		}
	}
	return idx
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isDigitOneToNine(b byte) bool {
	return b >= '1' && b <= '9'
}

func jsonTypeName(firstByte byte) string {
	switch {
	case firstByte == '{':
		return "map[string]interface {}"
	case firstByte == '[':
		return "[]interface {}"
	case firstByte == '"':
		return "string"
	case firstByte == 't' || firstByte == 'f':
		return "bool"
	case firstByte == 'n':
		return "<nil>"
	case firstByte == '-' || isDigit(firstByte):
		return "float64"
	default:
		return "interface {}"
	}
}

type yamlPathLocation struct {
	node   *yaml.Node
	parent *yaml.Node
}

func findYAMLPath(data []byte, path []string) (*yamlPathLocation, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	current := &root
	if current.Kind == yaml.DocumentNode {
		if len(current.Content) == 0 {
			return nil, fmt.Errorf("failed to decode YAML: empty document")
		}
		current = current.Content[0]
	}

	if len(path) == 0 {
		return &yamlPathLocation{node: current, parent: nil}, nil
	}

	for i, key := range path {
		if current.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("cannot traverse %s at path %s", yamlTypeName(current), pathx.Join(path[:i]))
		}

		found := false
		for j := 0; j+1 < len(current.Content); j += 2 {
			keyNode := current.Content[j]
			valNode := current.Content[j+1]
			if keyNode.Kind == yaml.ScalarNode && keyNode.Value == key {
				if i == len(path)-1 {
					return &yamlPathLocation{
						node:   valNode,
						parent: current,
					}, nil
				}
				current = valNode
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("path not found: %s", pathx.Join(path[:i+1]))
		}
	}

	return nil, fmt.Errorf("path not found: %s", pathx.Join(path))
}

func yamlTypeName(node *yaml.Node) string {
	if node == nil {
		return "interface {}"
	}
	switch node.Kind {
	case yaml.MappingNode:
		return "map[string]interface {}"
	case yaml.SequenceNode:
		return "[]interface {}"
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str", "":
			return "string"
		case "!!int":
			return "int"
		case "!!float":
			return "float64"
		case "!!bool":
			return "bool"
		case "!!null":
			return "<nil>"
		default:
			return "interface {}"
		}
	default:
		return "interface {}"
	}
}

func yamlScalarSpan(data []byte, loc *yamlPathLocation, path []string) (int, int, error) {
	if loc == nil || loc.node == nil {
		return 0, 0, fmt.Errorf("path not found: %s", pathx.Join(path))
	}
	if loc.node.Kind != yaml.ScalarNode {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: value is not a scalar", pathx.Join(path))
	}
	if loc.parent != nil && (loc.parent.Style&yaml.FlowStyle) != 0 {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: flow-style parent mapping", pathx.Join(path))
	}
	if (loc.node.Style&yaml.LiteralStyle) != 0 || (loc.node.Style&yaml.FoldedStyle) != 0 {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: multiline scalar style", pathx.Join(path))
	}
	if strings.Contains(loc.node.Value, "\n") {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: multiline scalar content", pathx.Join(path))
	}
	if loc.node.Line <= 0 || loc.node.Column <= 0 {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: missing source position", pathx.Join(path))
	}

	start, err := lineColToOffset(data, loc.node.Line, loc.node.Column)
	if err != nil {
		return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: %w", pathx.Join(path), err)
	}

	switch {
	case (loc.node.Style & yaml.DoubleQuotedStyle) != 0:
		end, err := findYAMLDoubleQuotedEnd(data, start)
		if err != nil {
			return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: %w", pathx.Join(path), err)
		}
		return start, end, nil
	case (loc.node.Style & yaml.SingleQuotedStyle) != 0:
		end, err := findYAMLSingleQuotedEnd(data, start)
		if err != nil {
			return 0, 0, fmt.Errorf("byte-exact YAML update unsupported at path %s: %w", pathx.Join(path), err)
		}
		return start, end, nil
	default:
		end := findYAMLPlainScalarEnd(data, start)
		return start, end, nil
	}
}

func lineColToOffset(data []byte, targetLine int, targetCol int) (int, error) {
	if targetLine < 1 || targetCol < 1 {
		return 0, fmt.Errorf("invalid target position %d:%d", targetLine, targetCol)
	}

	line, col := 1, 1
	for idx := 0; idx < len(data); {
		if line == targetLine && col == targetCol {
			return idx, nil
		}

		r, size := utf8.DecodeRune(data[idx:])
		if r == utf8.RuneError && size == 1 {
			return 0, fmt.Errorf("invalid UTF-8 near byte offset %d", idx)
		}

		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		idx += size
	}

	if line == targetLine && col == targetCol {
		return len(data), nil
	}
	return 0, fmt.Errorf("position %d:%d out of range", targetLine, targetCol)
}

func findYAMLDoubleQuotedEnd(data []byte, start int) (int, error) {
	if start >= len(data) || data[start] != '"' {
		return 0, fmt.Errorf("double-quoted scalar not found at offset %d", start)
	}
	for idx := start + 1; idx < len(data); idx++ {
		switch data[idx] {
		case '\\':
			idx++
			if idx >= len(data) {
				return 0, fmt.Errorf("unterminated escape sequence")
			}
		case '"':
			return idx + 1, nil
		case '\n':
			return 0, fmt.Errorf("multiline double-quoted scalar")
		}
	}
	return 0, fmt.Errorf("unterminated double-quoted scalar")
}

func findYAMLSingleQuotedEnd(data []byte, start int) (int, error) {
	if start >= len(data) || data[start] != '\'' {
		return 0, fmt.Errorf("single-quoted scalar not found at offset %d", start)
	}
	for idx := start + 1; idx < len(data); idx++ {
		switch data[idx] {
		case '\'':
			if idx+1 < len(data) && data[idx+1] == '\'' {
				idx++
				continue
			}
			return idx + 1, nil
		case '\n':
			return 0, fmt.Errorf("multiline single-quoted scalar")
		}
	}
	return 0, fmt.Errorf("unterminated single-quoted scalar")
}

func findYAMLPlainScalarEnd(data []byte, start int) int {
	if start >= len(data) {
		return start
	}

	lineEnd := len(data)
	for idx := start; idx < len(data); idx++ {
		if data[idx] == '\n' {
			lineEnd = idx
			break
		}
	}

	commentStart := lineEnd
	for idx := start; idx < lineEnd; idx++ {
		if data[idx] == '#' && (idx == start || isYAMLHorizontalSpace(data[idx-1])) {
			commentStart = idx
			break
		}
	}

	end := commentStart
	for end > start && isYAMLHorizontalSpace(data[end-1]) {
		end--
	}
	return end
}

func isYAMLHorizontalSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r'
}

func yamlStringLiteral(value string, style yaml.Style, path []string) ([]byte, error) {
	switch {
	case (style & yaml.DoubleQuotedStyle) != 0:
		return []byte(strconv.Quote(value)), nil
	case (style & yaml.SingleQuotedStyle) != 0:
		return []byte("'" + strings.ReplaceAll(value, "'", "''") + "'"), nil
	case (style&yaml.LiteralStyle) != 0 || (style&yaml.FoldedStyle) != 0:
		return nil, fmt.Errorf("byte-exact YAML update unsupported at path %s: multiline scalar style", pathx.Join(path))
	default:
		return []byte(value), nil
	}
}

func yamlNumericLiteral(value int64, style yaml.Style, path []string) ([]byte, error) {
	num := strconv.FormatInt(value, 10)
	switch {
	case (style & yaml.DoubleQuotedStyle) != 0:
		return []byte(strconv.Quote(num)), nil
	case (style & yaml.SingleQuotedStyle) != 0:
		return []byte("'" + num + "'"), nil
	case (style&yaml.LiteralStyle) != 0 || (style&yaml.FoldedStyle) != 0:
		return nil, fmt.Errorf("byte-exact YAML update unsupported at path %s: multiline scalar style", pathx.Join(path))
	default:
		return []byte(num), nil
	}
}

func replaceSpan(data []byte, start int, end int, replacement []byte) []byte {
	out := make([]byte, 0, len(data)-(end-start)+len(replacement))
	out = append(out, data[:start]...)
	out = append(out, replacement...)
	out = append(out, data[end:]...)
	return out
}

func numericScalarToInt64(val any, path []string) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0, fmt.Errorf("value at path %s is out of int64 range (got %v)", pathx.Join(path), v)
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("value at path %s is out of int64 range (got %v)", pathx.Join(path), v)
		}
		return int64(v), nil
	case float32:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, fmt.Errorf("value at path %s is not a finite number", pathx.Join(path))
		}
		if math.Trunc(f) != f {
			return 0, fmt.Errorf("value at path %s must be an integer (got %v)", pathx.Join(path), v)
		}
		i := int64(f)
		if float64(i) != f {
			return 0, fmt.Errorf("value at path %s is out of int64 range (got %v)", pathx.Join(path), v)
		}
		return i, nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, fmt.Errorf("value at path %s is not a finite number", pathx.Join(path))
		}
		if math.Trunc(v) != v {
			return 0, fmt.Errorf("value at path %s must be an integer (got %v)", pathx.Join(path), v)
		}
		i := int64(v)
		if float64(i) != v {
			return 0, fmt.Errorf("value at path %s is out of int64 range (got %v)", pathx.Join(path), v)
		}
		return i, nil
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, nil
		}
		f, err := v.Float64()
		if err != nil {
			return 0, fmt.Errorf("value at path %s is not numeric (got %T)", pathx.Join(path), val)
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, fmt.Errorf("value at path %s is not a finite number", pathx.Join(path))
		}
		if math.Trunc(f) != f {
			return 0, fmt.Errorf("value at path %s must be an integer (got %v)", pathx.Join(path), f)
		}
		i := int64(f)
		if float64(i) != f {
			return 0, fmt.Errorf("value at path %s is out of int64 range (got %v)", pathx.Join(path), f)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("value at path %s is not numeric (got %T)", pathx.Join(path), val)
	}
}

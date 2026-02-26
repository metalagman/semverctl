package formats

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/metalagman/semverctl/internal/pathx"
)

// Codec handles encoding/decoding of specific file formats.
type Codec interface {
	Decode(data []byte) (any, error)
	Encode(data any) ([]byte, error)
	GetVersion(data any, path []string) (string, error)
	SetVersion(data any, path []string, version string) error
	GetNumericScalar(data any, path []string) (int64, error)
	SetNumericScalar(data any, path []string, value int64) error
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

// Decode parses JSON data.
func (c *JSONCodec) Decode(data []byte) (any, error) {
	var result any
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return result, nil
}

// Encode serializes to JSON.
func (c *JSONCodec) Encode(data any) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// GetVersion retrieves a version string at the given path.
func (c *JSONCodec) GetVersion(data any, path []string) (string, error) {
	val, err := pathx.Get(data, path)
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
func (c *JSONCodec) SetVersion(data any, path []string, version string) error {
	return pathx.Set(data, path, version)
}

// GetNumericScalar retrieves a numeric value at the given path.
func (c *JSONCodec) GetNumericScalar(data any, path []string) (int64, error) {
	val, err := pathx.Get(data, path)
	if err != nil {
		return 0, err
	}
	return numericScalarToInt64(val, path)
}

// SetNumericScalar updates a numeric value at the given path.
func (c *JSONCodec) SetNumericScalar(data any, path []string, value int64) error {
	return pathx.Set(data, path, float64(value))
}

// YAMLCodec handles YAML files.
type YAMLCodec struct{}

// Decode parses YAML data.
func (c *YAMLCodec) Decode(data []byte) (any, error) {
	var result any
	err := yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return normalizeYAML(result), nil
}

// normalizeYAML converts yaml.v3 specific types to standard Go types.
func normalizeYAML(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, v := range val {
			val[k] = normalizeYAML(v)
		}
		return val
	case map[any]any:
		m := make(map[string]any)
		for k, v := range val {
			if keyStr, ok := k.(string); ok {
				m[keyStr] = normalizeYAML(v)
			}
		}
		return m
	case []any:
		for i, v := range val {
			val[i] = normalizeYAML(v)
		}
		return val
	default:
		return v
	}
}

// Encode serializes to YAML.
func (c *YAMLCodec) Encode(data any) ([]byte, error) {
	return yaml.Marshal(data)
}

// GetVersion retrieves a version string at the given path.
func (c *YAMLCodec) GetVersion(data any, path []string) (string, error) {
	val, err := pathx.Get(data, path)
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
func (c *YAMLCodec) SetVersion(data any, path []string, version string) error {
	return pathx.Set(data, path, version)
}

// GetNumericScalar retrieves a numeric value at the given path.
func (c *YAMLCodec) GetNumericScalar(data any, path []string) (int64, error) {
	val, err := pathx.Get(data, path)
	if err != nil {
		return 0, err
	}
	return numericScalarToInt64(val, path)
}

// SetNumericScalar updates a numeric value at the given path.
func (c *YAMLCodec) SetNumericScalar(data any, path []string, value int64) error {
	return pathx.Set(data, path, value)
}

func numericScalarToInt64(val any, path []string) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
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
	default:
		return 0, fmt.Errorf("value at path %s is not numeric (got %T)", pathx.Join(path), val)
	}
}

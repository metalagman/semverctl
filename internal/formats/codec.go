package formats

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/metalagman/semverctl/internal/pathx"
	"gopkg.in/yaml.v3"
)

// Codec handles encoding/decoding of specific file formats
type Codec interface {
	Decode(data []byte) (interface{}, error)
	Encode(data interface{}) ([]byte, error)
	GetVersion(data interface{}, path []string) (string, error)
	SetVersion(data interface{}, path []string, version string) error
	GetNumericScalar(data interface{}, path []string) (int64, error)
	SetNumericScalar(data interface{}, path []string, value int64) error
}

// GetCodec returns the appropriate codec for a file based on extension
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

// JSONCodec handles JSON files
type JSONCodec struct{}

// Decode parses JSON data
func (c *JSONCodec) Decode(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return result, nil
}

// Encode serializes to JSON
func (c *JSONCodec) Encode(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// GetVersion retrieves a version string at the given path
func (c *JSONCodec) GetVersion(data interface{}, path []string) (string, error) {
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

// SetVersion updates a version string at the given path
func (c *JSONCodec) SetVersion(data interface{}, path []string, version string) error {
	return pathx.Set(data, path, version)
}

// GetNumericScalar retrieves a numeric value at the given path
func (c *JSONCodec) GetNumericScalar(data interface{}, path []string) (int64, error) {
	val, err := pathx.Get(data, path)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("value at path %s is not numeric (got %T)", pathx.Join(path), val)
	}
}

// SetNumericScalar updates a numeric value at the given path
func (c *JSONCodec) SetNumericScalar(data interface{}, path []string, value int64) error {
	return pathx.Set(data, path, float64(value))
}

// YAMLCodec handles YAML files
type YAMLCodec struct{}

// Decode parses YAML data
func (c *YAMLCodec) Decode(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return normalizeYAML(result), nil
}

// normalizeYAML converts yaml.v3 specific types to standard Go types
func normalizeYAML(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v := range val {
			val[k] = normalizeYAML(v)
		}
		return val
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			if keyStr, ok := k.(string); ok {
				m[keyStr] = normalizeYAML(v)
			}
		}
		return m
	case []interface{}:
		for i, v := range val {
			val[i] = normalizeYAML(v)
		}
		return val
	default:
		return v
	}
}

// Encode serializes to YAML
func (c *YAMLCodec) Encode(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

// GetVersion retrieves a version string at the given path
func (c *YAMLCodec) GetVersion(data interface{}, path []string) (string, error) {
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

// SetVersion updates a version string at the given path
func (c *YAMLCodec) SetVersion(data interface{}, path []string, version string) error {
	return pathx.Set(data, path, version)
}

// GetNumericScalar retrieves a numeric value at the given path
func (c *YAMLCodec) GetNumericScalar(data interface{}, path []string) (int64, error) {
	val, err := pathx.Get(data, path)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("value at path %s is not numeric (got %T)", pathx.Join(path), val)
	}
}

// SetNumericScalar updates a numeric value at the given path
func (c *YAMLCodec) SetNumericScalar(data interface{}, path []string, value int64) error {
	return pathx.Set(data, path, value)
}

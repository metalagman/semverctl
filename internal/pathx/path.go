package pathx

import (
	"fmt"
	"strings"
)

// Parse splits a dot-path into components
// Supports formats like: .version, .app.version, version, app.version
func Parse(path string) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Remove leading dot if present
	path = strings.TrimPrefix(path, ".")

	if path == "" {
		return nil, fmt.Errorf("path cannot be just a dot")
	}

	// Split by dots
	parts := strings.Split(path, ".")

	// Validate each part
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty path component at position %d", i)
		}
	}

	return parts, nil
}

// Get traverses a nested map/slice structure following the path
// Returns the value at the path or an error if path doesn't exist
func Get(data interface{}, path []string) (interface{}, error) {
	current := data

	for i, key := range path {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[key]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", strings.Join(path[:i+1], "."))
			}
			current = val
		case map[interface{}]interface{}:
			val, ok := v[key]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", strings.Join(path[:i+1], "."))
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot traverse %T at path %s", current, strings.Join(path[:i], "."))
		}
	}

	return current, nil
}

// Set updates a value at the given path in a nested map structure
// Returns an error if the path doesn't exist or if intermediate values are not maps
func Set(data interface{}, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	current := data

	// Navigate to parent of target
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		switch v := current.(type) {
		case map[string]interface{}:
			next, ok := v[key]
			if !ok {
				return fmt.Errorf("path not found: %s", strings.Join(path[:i+1], "."))
			}
			current = next
		case map[interface{}]interface{}:
			next, ok := v[key]
			if !ok {
				return fmt.Errorf("path not found: %s", strings.Join(path[:i+1], "."))
			}
			current = next
		default:
			return fmt.Errorf("cannot traverse %T at path %s", current, strings.Join(path[:i], "."))
		}
	}

	// Set the value
	targetKey := path[len(path)-1]
	switch v := current.(type) {
	case map[string]interface{}:
		if _, ok := v[targetKey]; !ok {
			return fmt.Errorf("path not found: %s", strings.Join(path, "."))
		}
		v[targetKey] = value
	case map[interface{}]interface{}:
		if _, ok := v[targetKey]; !ok {
			return fmt.Errorf("path not found: %s", strings.Join(path, "."))
		}
		v[targetKey] = value
	default:
		return fmt.Errorf("cannot set value in %T at path %s", current, strings.Join(path[:len(path)-1], "."))
	}

	return nil
}

// Join joins path components into a dot-path string
func Join(parts []string) string {
	return strings.Join(parts, ".")
}

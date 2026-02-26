package targets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Resolver handles file target resolution.
type Resolver struct {
	roots []string
}

// NewResolver creates a new target resolver with the given roots.
func NewResolver(roots []string) *Resolver {
	if len(roots) == 0 {
		roots = []string{"."}
	}
	return &Resolver{roots: roots}
}

// Resolve resolves targets from a single file path.
func (r *Resolver) Resolve(file string) ([]string, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", file, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", file)
	}

	if !isSupportedExtension(file) {
		return nil, fmt.Errorf("unsupported file extension: %s", file)
	}

	return []string{file}, nil
}

// ResolveGlob resolves targets from a glob pattern across all roots.
func (r *Resolver) ResolveGlob(pattern string) ([]string, error) {
	var allMatches []string
	seen := make(map[string]bool)

	for _, root := range r.roots {
		matches, err := doublestar.Glob(os.DirFS(root), pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			fullPath := filepath.Join(root, match)

			// Check if it's a directory
			info, err := os.Stat(fullPath)
			if err != nil {
				return nil, fmt.Errorf("stat %q: %w", fullPath, err)
			}
			if info.IsDir() {
				continue // Skip directories
			}

			// Check extension
			if !isSupportedExtension(fullPath) {
				continue // Skip unsupported extensions
			}

			// Deduplicate
			absPath, err := filepath.Abs(fullPath)
			if err != nil {
				absPath = fullPath
			}

			if !seen[absPath] {
				seen[absPath] = true
				allMatches = append(allMatches, fullPath)
			}
		}
	}

	if len(allMatches) == 0 {
		return nil, fmt.Errorf("glob pattern %q matched no files", pattern)
	}

	// Sort for consistent ordering
	sort.Strings(allMatches)

	return allMatches, nil
}

// isSupportedExtension checks if the file has a supported extension.
func isSupportedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

// ValidateSingleTarget validates that exactly one of file or glob is specified.
func ValidateSingleTarget(file, glob string) error {
	hasFile := file != ""
	hasGlob := glob != ""

	if hasFile && hasGlob {
		return fmt.Errorf("cannot specify both --file and --glob")
	}
	if !hasFile && !hasGlob {
		return fmt.Errorf("must specify either --file or --glob")
	}
	return nil
}

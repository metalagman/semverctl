package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/metalagman/semverctl/internal/formats"
	"github.com/metalagman/semverctl/internal/pathx"
	"github.com/metalagman/semverctl/internal/semverx"
	"github.com/metalagman/semverctl/internal/targets"
)

// Operation represents the type of operation to perform.
type Operation string

const (
	OpBump Operation = "bump"
	OpSet  Operation = "set"
)

// Config holds the application configuration.
type Config struct {
	Operation   Operation
	Version     string // For OpSet
	BumpType    string // For OpBump (major, minor, patch)
	Path        []string
	File        string
	Glob        string
	Roots       []string
	DryRun      bool
	NumericBump bool // For numeric scalar bumps
}

// Result holds the result of processing a single file.
type Result struct {
	File    string
	OldVer  string
	NewVer  string
	Changed bool
	Error   error
}

// Runner orchestrates the version bumping/setting process.
type Runner struct {
	config         *Config
	suppressOutput bool
}

// NewRunner creates a new app runner.
func NewRunner(config *Config) *Runner {
	return &Runner{config: config}
}

// Run executes the version bumping/setting operation.
func (r *Runner) Run() error {
	prevSuppressOutput := r.suppressOutput
	r.suppressOutput = false
	defer func() { r.suppressOutput = prevSuppressOutput }()

	_, err := r.runWithOptionalOutput()
	return err
}

// RunWithResults executes the operation and returns per-file results without
// printing human-readable output.
func (r *Runner) RunWithResults() ([]Result, error) {
	prevSuppressOutput := r.suppressOutput
	r.suppressOutput = true
	defer func() { r.suppressOutput = prevSuppressOutput }()

	return r.runWithOptionalOutput()
}

func (r *Runner) runWithOptionalOutput() ([]Result, error) {
	// Validate config
	if err := r.validate(); err != nil {
		return nil, err
	}

	// Resolve targets
	resolver := targets.NewResolver(r.config.Roots)
	var files []string
	var err error

	if r.config.File != "" {
		files, err = resolver.Resolve(r.config.File)
	} else {
		files, err = resolver.ResolveGlob(r.config.Glob)
	}
	if err != nil {
		return nil, err
	}

	// Process files
	results := make([]Result, 0, len(files))
	hasErrors := false

	for _, file := range files {
		result := r.processFile(file)
		results = append(results, result)
		if result.Error != nil {
			hasErrors = true
		}
	}

	if !r.suppressOutput {
		r.printResults(results)
	}

	if hasErrors {
		return results, fmt.Errorf("one or more files failed to process")
	}

	return results, nil
}

func (r *Runner) validate() error {
	// Validate operation-specific config
	switch r.config.Operation {
	case OpBump:
		if !r.config.NumericBump && r.config.BumpType == "" {
			return fmt.Errorf("bump type required (major, minor, patch)")
		}
		if r.config.NumericBump && r.config.BumpType != "" {
			return fmt.Errorf("cannot use --major/--minor/--patch with numeric bump")
		}
	case OpSet:
		if r.config.Version == "" {
			return fmt.Errorf("version required for set operation")
		}
		if r.config.NumericBump {
			return fmt.Errorf("numeric bump not supported for set operation")
		}
		if !semverx.IsValid(r.config.Version) {
			return fmt.Errorf("invalid semantic version: %s", r.config.Version)
		}
	default:
		return fmt.Errorf("unknown operation: %s", r.config.Operation)
	}

	// Validate path
	if len(r.config.Path) == 0 {
		r.config.Path = []string{"version"}
	}

	// Validate file/glob
	err := targets.ValidateSingleTarget(r.config.File, r.config.Glob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) processFile(file string) Result {
	result := Result{File: file}

	// Read file
	data, err := os.ReadFile(file)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		return result
	}

	// Get codec
	codec, err := formats.GetCodec(file)
	if err != nil {
		result.Error = err
		return result
	}

	// Decode
	doc, err := codec.Decode(data)
	if err != nil {
		result.Error = err
		return result
	}

	// Process based on operation type
	if r.config.NumericBump {
		return r.processNumericBump(file, data, doc, codec)
	}

	return r.processSemver(file, data, doc, codec)
}

func (r *Runner) processSemver(file string, data []byte, doc any, codec formats.Codec) Result {
	result := Result{File: file}

	// Get current version
	oldVer, err := codec.GetVersion(doc, r.config.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to get version: %w", err)
		return result
	}
	result.OldVer = oldVer

	// Calculate new version
	var newVer string
	switch r.config.Operation {
	case OpBump:
		version, err := semverx.Parse(oldVer)
		if err != nil {
			result.Error = fmt.Errorf("failed to parse version %q: %w", oldVer, err)
			return result
		}
		newVersion, err := version.Bump(r.config.BumpType)
		if err != nil {
			result.Error = err
			return result
		}
		newVer = newVersion.String()
	case OpSet:
		newVer = r.config.Version
	}
	result.NewVer = newVer

	// Check if changed
	if oldVer == newVer {
		result.Changed = false
		return result
	}
	result.Changed = true

	// Update document
	if err := codec.SetVersion(doc, r.config.Path, newVer); err != nil {
		result.Error = fmt.Errorf("failed to set version: %w", err)
		return result
	}

	// Encode
	newData, err := codec.Encode(doc)
	if err != nil {
		result.Error = fmt.Errorf("failed to encode: %w", err)
		return result
	}

	// Output
	if r.config.DryRun {
		if !r.suppressOutput {
			r.printDiff(file, string(data), string(newData))
		}
	} else {
		err := os.WriteFile(file, newData, 0o644)
		if err != nil {
			result.Error = fmt.Errorf("failed to write file: %w", err)
			return result
		}
	}

	return result
}

func (r *Runner) processNumericBump(file string, data []byte, doc any, codec formats.Codec) Result {
	result := Result{File: file}

	// Get current numeric value
	oldVal, err := codec.GetNumericScalar(doc, r.config.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to get numeric value: %w", err)
		return result
	}
	result.OldVer = fmt.Sprintf("%d", oldVal)

	// Bump by 1
	newVal := oldVal + 1
	result.NewVer = fmt.Sprintf("%d", newVal)

	// Check if changed
	if oldVal == newVal {
		result.Changed = false
		return result
	}
	result.Changed = true

	// Update document
	if err := codec.SetNumericScalar(doc, r.config.Path, newVal); err != nil {
		result.Error = fmt.Errorf("failed to set numeric value: %w", err)
		return result
	}

	// Encode
	newData, err := codec.Encode(doc)
	if err != nil {
		result.Error = fmt.Errorf("failed to encode: %w", err)
		return result
	}

	// Output
	if r.config.DryRun {
		if !r.suppressOutput {
			r.printDiff(file, string(data), string(newData))
		}
	} else {
		err := os.WriteFile(file, newData, 0o644)
		if err != nil {
			result.Error = fmt.Errorf("failed to write file: %w", err)
			return result
		}
	}

	return result
}

func (r *Runner) printResults(results []Result) {
	changed := 0
	unchanged := 0
	errors := 0

	for _, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", result.File, result.Error)
			errors++
		} else if result.Changed {
			action := "updated"
			if r.config.DryRun {
				action = "would update"
			}
			fmt.Printf("✓ %s: %s → %s (%s)\n", result.File, result.OldVer, result.NewVer, action)
			changed++
		} else {
			fmt.Printf("- %s: no change (%s)\n", result.File, result.OldVer)
			unchanged++
		}
	}

	fmt.Printf("\nSummary: %d changed, %d unchanged", changed, unchanged)
	if errors > 0 {
		fmt.Printf(", %d errors", errors)
	}
	fmt.Println()
}

func (r *Runner) printDiff(filename, oldContent, newContent string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldContent, newContent, false)

	fmt.Printf("--- %s\n+++ %s\n", filename, filename)

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			for line := range strings.SplitSeq(diff.Text, "\n") {
				if line != "" {
					fmt.Printf("-%s\n", line)
				}
			}
		case diffmatchpatch.DiffInsert:
			for line := range strings.SplitSeq(diff.Text, "\n") {
				if line != "" {
					fmt.Printf("+%s\n", line)
				}
			}
		case diffmatchpatch.DiffEqual:
			// Skip unchanged lines for cleaner output
		}
	}
	fmt.Println()
}

// JoinPath joins path components for display.
func JoinPath(path []string) string {
	return pathx.Join(path)
}

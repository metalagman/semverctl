package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/metalagman/semverctl/internal/app"
	"github.com/metalagman/semverctl/internal/pathx"
)

// MinimumNArgsWithHelp returns an args validator that prints help before
// returning an error when too few positional args are provided.
func MinimumNArgsWithHelp(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			_ = cmd.Help()
			return fmt.Errorf("requires at least %d arg(s)", n)
		}
		return nil
	}
}

// ExactNArgsWithHelp returns an args validator that prints help before
// returning an error when positional arg count does not match exactly.
func ExactNArgsWithHelp(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			_ = cmd.Help()
			return fmt.Errorf("requires exactly %d arg(s)", n)
		}
		return nil
	}
}

func init() {
	// Bump command
	bumpCmd := &cobra.Command{
		Use:   "bump",
		Short: "Bump semantic version in JSON/YAML files",
		Long: `Bump the semantic version at the specified path in JSON or YAML files.

By default, bumps the patch version in package.json. Use --file or --glob to
target another file set.

Examples:
  semverctl bump                                              # Bump patch in package.json
  semverctl bump --minor                                      # Bump minor in package.json
  semverctl bump --file config.yaml --path .app.version       # Bump at custom path
  semverctl bump --glob "**/*.json"                           # Bump all JSON files
  semverctl bump --numeric --path .version.Patch --file config.json # Bump numeric field`,
		Args: ExactNArgsWithHelp(0),
		RunE: runBump,
	}

	bumpCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	bumpCmd.Flags().String("file", "", "Specific file to process (default: package.json)")
	bumpCmd.Flags().String("glob", "", "Glob pattern for matching files under current directory (e.g., '**/*.json')")
	bumpCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")
	bumpCmd.Flags().Bool("major", false, "Bump major version")
	bumpCmd.Flags().Bool("minor", false, "Bump minor version")
	bumpCmd.Flags().Bool("patch", false, "Bump patch version (default)")
	bumpCmd.Flags().Bool("numeric", false, "Bump numeric scalar by +1 (for object-style versions)")

	rootCmd.AddCommand(bumpCmd)

	// Set command
	setCmd := &cobra.Command{
		Use:   "set VERSION",
		Short: "Set semantic version in JSON/YAML files",
		Long: `Set the semantic version at the specified path in JSON or YAML files to an explicit value.

Examples:
  semverctl set 1.2.3                                      # Set version in package.json
  semverctl set 2.0.0 --glob "**/*.json"                  # Set version in all JSON files
  semverctl set 1.0.0 --path .app.version --file config.yaml # Set version at custom path`,
		Args: ExactNArgsWithHelp(1),
		RunE: runSet,
	}

	setCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	setCmd.Flags().String("file", "", "Specific file to process (default: package.json)")
	setCmd.Flags().String("glob", "", "Glob pattern for matching files under current directory (e.g., '**/*.json')")
	setCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")

	rootCmd.AddCommand(setCmd)
}

type bumpFlags struct {
	path        string
	file        string
	glob        string
	dryRun      bool
	major       bool
	minor       bool
	patch       bool
	numericBump bool
}

func loadBumpFlags(cmd *cobra.Command) (bumpFlags, error) {
	var f bumpFlags
	var err error

	if f.path, err = cmd.Flags().GetString("path"); err != nil {
		return f, fmt.Errorf("get --path: %w", err)
	}
	if f.file, err = cmd.Flags().GetString("file"); err != nil {
		return f, fmt.Errorf("get --file: %w", err)
	}
	if f.glob, err = cmd.Flags().GetString("glob"); err != nil {
		return f, fmt.Errorf("get --glob: %w", err)
	}
	if f.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return f, fmt.Errorf("get --dry-run: %w", err)
	}
	if f.major, err = cmd.Flags().GetBool("major"); err != nil {
		return f, fmt.Errorf("get --major: %w", err)
	}
	if f.minor, err = cmd.Flags().GetBool("minor"); err != nil {
		return f, fmt.Errorf("get --minor: %w", err)
	}
	if f.patch, err = cmd.Flags().GetBool("patch"); err != nil {
		return f, fmt.Errorf("get --patch: %w", err)
	}
	if f.numericBump, err = cmd.Flags().GetBool("numeric"); err != nil {
		return f, fmt.Errorf("get --numeric: %w", err)
	}
	return f, nil
}

type setFlags struct {
	path   string
	file   string
	glob   string
	dryRun bool
}

const defaultTargetFile = "package.json"

func loadSetFlags(cmd *cobra.Command) (setFlags, error) {
	var f setFlags
	var err error

	if f.path, err = cmd.Flags().GetString("path"); err != nil {
		return f, fmt.Errorf("get --path: %w", err)
	}
	if f.file, err = cmd.Flags().GetString("file"); err != nil {
		return f, fmt.Errorf("get --file: %w", err)
	}
	if f.glob, err = cmd.Flags().GetString("glob"); err != nil {
		return f, fmt.Errorf("get --glob: %w", err)
	}
	if f.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return f, fmt.Errorf("get --dry-run: %w", err)
	}
	return f, nil
}

func runBump(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("bump accepts no positional args; use --file or --glob")
	}

	flags, err := loadBumpFlags(cmd)
	if err != nil {
		return err
	}

	// Validate flags
	if flags.numericBump {
		if flags.major || flags.minor || flags.patch {
			return fmt.Errorf("cannot use --major/--minor/--patch with --numeric")
		}
	} else {
		majorSet := cmd.Flags().Changed("major")
		minorSet := cmd.Flags().Changed("minor")
		patchSet := cmd.Flags().Changed("patch")

		if !majorSet && !minorSet && !patchSet {
			flags.patch = true
		}

		bumpTypes := 0
		if flags.major {
			bumpTypes++
		}
		if flags.minor {
			bumpTypes++
		}
		if flags.patch {
			bumpTypes++
		}
		if bumpTypes > 1 {
			return fmt.Errorf("can only specify one of --major, --minor, or --patch")
		}
	}

	bumpType := ""
	if !flags.numericBump {
		bumpType = "patch"
		if flags.major {
			bumpType = "major"
		} else if flags.minor {
			bumpType = "minor"
		}
	}

	path, err := pathx.Parse(flags.path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flags.path, err)
	}

	file := flags.file
	glob := flags.glob
	if file == "" && glob == "" {
		file = defaultTargetFile
	}

	config := &app.Config{
		Operation:   app.OpBump,
		BumpType:    bumpType,
		Path:        path,
		File:        file,
		Glob:        glob,
		DryRun:      flags.dryRun,
		NumericBump: flags.numericBump,
	}

	return app.NewRunner(config).Run()
}

func runSet(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires exactly 1 arg(s)")
	}
	if len(args) > 1 {
		return fmt.Errorf("set accepts exactly 1 positional arg (VERSION); use --file or --glob")
	}

	flags, err := loadSetFlags(cmd)
	if err != nil {
		return err
	}

	version := args[0]

	path, err := pathx.Parse(flags.path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flags.path, err)
	}

	file := flags.file
	glob := flags.glob
	if file == "" && glob == "" {
		file = defaultTargetFile
	}

	config := &app.Config{
		Operation: app.OpSet,
		Version:   version,
		Path:      path,
		File:      file,
		Glob:      glob,
		DryRun:    flags.dryRun,
	}

	return app.NewRunner(config).Run()
}

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/metalagman/semverctl/internal/app"
	"github.com/metalagman/semverctl/internal/pathx"
)

func MinimumNArgsWithHelp(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			cmd.Help()
			fmt.Println()
			os.Exit(1)
		}
		return nil
	}
}

var (
	flagPath        string
	flagFile        string
	flagGlob        string
	flagDryRun      bool
	flagMajor       bool
	flagMinor       bool
	flagPatch       bool
	flagNumericBump bool
)

func init() {
	// Bump command
	bumpCmd := &cobra.Command{
		Use:   "bump [roots...]",
		Short: "Bump semantic version in JSON/YAML files",
		Long: `Bump the semantic version at the specified path in JSON or YAML files.

By default, bumps the patch version. Use --major, --minor, or --patch to specify
the version component to increment.

Examples:
  semverctl bump package.json                    # Bump patch in package.json
  semverctl bump --minor package.json            # Bump minor in package.json
  semverctl bump --glob "**/*.json" .            # Bump all JSON files
  semverctl bump --path .app.version config.yaml # Bump at custom path
  semverctl bump --numeric --path .version.Patch config.yaml # Bump numeric field`,
		RunE: runBump,
	}

	bumpCmd.Flags().StringVar(&flagPath, "path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	bumpCmd.Flags().StringVar(&flagFile, "file", "", "Specific file to process")
	bumpCmd.Flags().StringVar(&flagGlob, "glob", "", "Glob pattern for matching files (e.g., '**/*.json')")
	bumpCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show diffs without modifying files")
	bumpCmd.Flags().BoolVar(&flagMajor, "major", false, "Bump major version")
	bumpCmd.Flags().BoolVar(&flagMinor, "minor", false, "Bump minor version")
	bumpCmd.Flags().BoolVar(&flagPatch, "patch", false, "Bump patch version (default)")
	bumpCmd.Flags().BoolVar(&flagNumericBump, "numeric", false, "Bump numeric scalar by +1 (for object-style versions)")

	rootCmd.AddCommand(bumpCmd)

	// Set command
	setCmd := &cobra.Command{
		Use:   "set VERSION [roots...]",
		Short: "Set semantic version in JSON/YAML files",
		Long: `Set the semantic version at the specified path in JSON or YAML files to an explicit value.

Examples:
  semverctl set 1.2.3 package.json                    # Set version in package.json
  semverctl set 2.0.0 --glob "**/*.json" .            # Set version in all JSON files
  semverctl set 1.0.0 --path .app.version config.yaml # Set version at custom path`,
		Args: MinimumNArgsWithHelp(1),
		RunE: runSet,
	}

	setCmd.Flags().StringVar(&flagPath, "path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	setCmd.Flags().StringVar(&flagFile, "file", "", "Specific file to process")
	setCmd.Flags().StringVar(&flagGlob, "glob", "", "Glob pattern for matching files (e.g., '**/*.json')")
	setCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show diffs without modifying files")

	rootCmd.AddCommand(setCmd)
}

func runBump(cmd *cobra.Command, args []string) error {
	// Validate flags
	if flagNumericBump {
		if flagMajor || flagMinor || flagPatch {
			return fmt.Errorf("cannot use --major/--minor/--patch with --numeric")
		}
	} else {
		// Check if any bump type flag was explicitly set
		majorSet := cmd.Flags().Changed("major")
		minorSet := cmd.Flags().Changed("minor")
		patchSet := cmd.Flags().Changed("patch")

		// Default to patch if no bump type explicitly specified
		if !majorSet && !minorSet && !patchSet {
			flagPatch = true
		}

		// Validate only one bump type
		bumpTypes := 0
		if flagMajor {
			bumpTypes++
		}
		if flagMinor {
			bumpTypes++
		}
		if flagPatch {
			bumpTypes++
		}
		if bumpTypes > 1 {
			return fmt.Errorf("can only specify one of --major, --minor, or --patch")
		}
	}

	// Determine bump type (only for non-numeric bumps)
	bumpType := ""
	if !flagNumericBump {
		bumpType = "patch"
		if flagMajor {
			bumpType = "major"
		} else if flagMinor {
			bumpType = "minor"
		}
	}

	// Parse path
	path, err := pathx.Parse(flagPath)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flagPath, err)
	}

	// Validate file/glob
	if flagFile == "" && flagGlob == "" {
		// Use positional args as roots with default glob
		if len(args) == 0 {
			args = []string{"."}
		}
		flagGlob = "**/*.json"
		if len(args) > 0 {
			flagFile = args[0]
			args = args[1:]
		}
	}

	// Build config
	config := &app.Config{
		Operation:   app.OpBump,
		BumpType:    bumpType,
		Path:        path,
		File:        flagFile,
		Glob:        flagGlob,
		Roots:       args,
		DryRun:      flagDryRun,
		NumericBump: flagNumericBump,
	}

	runner := app.NewRunner(config)
	return runner.Run()
}

func runSet(cmd *cobra.Command, args []string) error {
	version := args[0]
	roots := args[1:]

	// Parse path
	path, err := pathx.Parse(flagPath)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flagPath, err)
	}

	// Validate file/glob
	if flagFile == "" && flagGlob == "" {
		// Use positional args
		if len(roots) == 0 {
			return fmt.Errorf("must specify a file or glob pattern")
		}
		flagFile = roots[0]
		roots = roots[1:]
	}

	// Build config
	config := &app.Config{
		Operation: app.OpSet,
		Version:   version,
		Path:      path,
		File:      flagFile,
		Glob:      flagGlob,
		Roots:     roots,
		DryRun:    flagDryRun,
	}

	runner := app.NewRunner(config)
	return runner.Run()
}

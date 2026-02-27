package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/metalagman/semverctl/internal/app"
	"github.com/metalagman/semverctl/internal/gitx"
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

type tagService interface {
	EnsureClean() error
	NextTag(component string) (string, error)
	NormalizeTag(versionOrTag string) (string, error)
	CreateAnnotatedTag(tag string) error
	PushTag(tag string) error
}

var newTagService = func() tagService {
	return gitx.NewService()
}

func init() {
	bumpCmd := &cobra.Command{
		Use:   "bump",
		Short: "Bump semantic version in file content or git tags",
		Long: `Bump semantic versions either in files or git tags.

Use 'bump file' to update a JSON/YAML file.
Use 'bump tag' to create a new git tag based on the latest stable vX.Y.Z tag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return fmt.Errorf("requires a subcommand: file or tag")
		},
	}

	bumpFileCmd := &cobra.Command{
		Use:   "file PATH",
		Short: "Bump semantic version in a JSON/YAML file",
		Long: `Bump the semantic version at the specified path in a JSON or YAML file.

Examples:
  semverctl bump file package.json
  semverctl bump file config.yaml --minor --path .app.version
  semverctl bump file config.json --numeric --path .version.Patch`,
		Args: ExactNArgsWithHelp(1),
		RunE: runBumpFile,
	}
	bumpFileCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	bumpFileCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")
	bumpFileCmd.Flags().Bool("major", false, "Bump major version")
	bumpFileCmd.Flags().Bool("minor", false, "Bump minor version")
	bumpFileCmd.Flags().Bool("patch", false, "Bump patch version (default)")
	bumpFileCmd.Flags().Bool("numeric", false, "Bump numeric scalar by +1 (for object-style versions)")

	bumpTagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Bump semantic version using git tags",
		Long: `Create the next git release tag from the highest stable existing tag.

By default this command creates the local annotated tag. Use --dry-run to preview.

Examples:
  semverctl bump tag
  semverctl bump tag --minor
  semverctl bump tag --dry-run
  semverctl bump tag --push`,
		Args: ExactNArgsWithHelp(0),
		RunE: runBumpTag,
	}
	bumpTagCmd.Flags().Bool("major", false, "Bump major version")
	bumpTagCmd.Flags().Bool("minor", false, "Bump minor version")
	bumpTagCmd.Flags().Bool("patch", false, "Bump patch version (default)")
	bumpTagCmd.Flags().Bool("push", false, "Push created tag to origin")
	bumpTagCmd.Flags().Bool("dry-run", false, "Preview resulting tag without creating it")

	bumpCmd.AddCommand(bumpFileCmd, bumpTagCmd)
	rootCmd.AddCommand(bumpCmd)

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set semantic version in file content or git tags",
		Long: `Set semantic versions explicitly either in files or git tags.

Use 'set file' to set a version value in JSON/YAML files.
Use 'set tag' to create an explicit git tag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return fmt.Errorf("requires a subcommand: file or tag")
		},
	}

	setFileCmd := &cobra.Command{
		Use:   "file VERSION PATH",
		Short: "Set semantic version in a JSON/YAML file",
		Long: `Set the semantic version at the specified path in a JSON or YAML file.

Examples:
  semverctl set file 1.2.3 package.json
  semverctl set file 2.0.0 config.yaml --path .app.version`,
		Args: ExactNArgsWithHelp(2),
		RunE: runSetFile,
	}
	setFileCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	setFileCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")

	setTagCmd := &cobra.Command{
		Use:   "tag VERSION",
		Short: "Create an explicit semantic git tag",
		Long: `Create an explicit annotated git tag.

VERSION accepts either 1.2.3 or v1.2.3.
By default this command creates the local tag. Use --dry-run to preview.

Examples:
  semverctl set tag 1.2.3
  semverctl set tag v2.0.0
  semverctl set tag 1.2.3 --dry-run
  semverctl set tag 1.2.3 --push`,
		Args: ExactNArgsWithHelp(1),
		RunE: runSetTag,
	}
	setTagCmd.Flags().Bool("push", false, "Push created tag to origin")
	setTagCmd.Flags().Bool("dry-run", false, "Preview resulting tag without creating it")

	setCmd.AddCommand(setFileCmd, setTagCmd)
	rootCmd.AddCommand(setCmd)
}

type bumpFileFlags struct {
	path        string
	dryRun      bool
	major       bool
	minor       bool
	patch       bool
	numericBump bool
}

func loadBumpFileFlags(cmd *cobra.Command) (bumpFileFlags, error) {
	var f bumpFileFlags
	var err error

	if f.path, err = cmd.Flags().GetString("path"); err != nil {
		return f, fmt.Errorf("get --path: %w", err)
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

type setFileFlags struct {
	path   string
	dryRun bool
}

func loadSetFileFlags(cmd *cobra.Command) (setFileFlags, error) {
	var f setFileFlags
	var err error

	if f.path, err = cmd.Flags().GetString("path"); err != nil {
		return f, fmt.Errorf("get --path: %w", err)
	}
	if f.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return f, fmt.Errorf("get --dry-run: %w", err)
	}

	return f, nil
}

type bumpTagFlags struct {
	major  bool
	minor  bool
	patch  bool
	push   bool
	dryRun bool
}

func loadBumpTagFlags(cmd *cobra.Command) (bumpTagFlags, error) {
	var f bumpTagFlags
	var err error

	if f.major, err = cmd.Flags().GetBool("major"); err != nil {
		return f, fmt.Errorf("get --major: %w", err)
	}
	if f.minor, err = cmd.Flags().GetBool("minor"); err != nil {
		return f, fmt.Errorf("get --minor: %w", err)
	}
	if f.patch, err = cmd.Flags().GetBool("patch"); err != nil {
		return f, fmt.Errorf("get --patch: %w", err)
	}
	if f.push, err = cmd.Flags().GetBool("push"); err != nil {
		return f, fmt.Errorf("get --push: %w", err)
	}
	if f.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return f, fmt.Errorf("get --dry-run: %w", err)
	}

	return f, nil
}

type setTagFlags struct {
	push   bool
	dryRun bool
}

func loadSetTagFlags(cmd *cobra.Command) (setTagFlags, error) {
	var f setTagFlags
	var err error

	if f.push, err = cmd.Flags().GetBool("push"); err != nil {
		return f, fmt.Errorf("get --push: %w", err)
	}
	if f.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return f, fmt.Errorf("get --dry-run: %w", err)
	}

	return f, nil
}

func resolveBumpComponent(cmd *cobra.Command, major, minor, patch bool) (string, error) {
	majorSet := cmd.Flags().Changed("major")
	minorSet := cmd.Flags().Changed("minor")
	patchSet := cmd.Flags().Changed("patch")

	if !majorSet && !minorSet && !patchSet {
		patch = true
	}

	bumpTypes := 0
	if major {
		bumpTypes++
	}
	if minor {
		bumpTypes++
	}
	if patch {
		bumpTypes++
	}
	if bumpTypes > 1 {
		return "", fmt.Errorf("can only specify one of --major, --minor, or --patch")
	}

	switch {
	case major:
		return "major", nil
	case minor:
		return "minor", nil
	default:
		return "patch", nil
	}
}

func runBumpFile(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("requires exactly 1 arg(s)")
	}

	flags, err := loadBumpFileFlags(cmd)
	if err != nil {
		return err
	}

	if flags.numericBump {
		if flags.major || flags.minor || flags.patch {
			return fmt.Errorf("cannot use --major/--minor/--patch with --numeric")
		}
	}

	bumpType := ""
	if !flags.numericBump {
		bumpType, err = resolveBumpComponent(cmd, flags.major, flags.minor, flags.patch)
		if err != nil {
			return err
		}
	}

	path, err := pathx.Parse(flags.path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flags.path, err)
	}

	config := &app.Config{
		Operation:   app.OpBump,
		BumpType:    bumpType,
		Path:        path,
		File:        args[0],
		DryRun:      flags.dryRun,
		NumericBump: flags.numericBump,
	}

	return app.NewRunner(config).Run()
}

func runSetFile(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("requires exactly 2 arg(s)")
	}

	flags, err := loadSetFileFlags(cmd)
	if err != nil {
		return err
	}

	path, err := pathx.Parse(flags.path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", flags.path, err)
	}

	config := &app.Config{
		Operation: app.OpSet,
		Version:   args[0],
		Path:      path,
		File:      args[1],
		DryRun:    flags.dryRun,
	}

	return app.NewRunner(config).Run()
}

func runBumpTag(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("requires exactly 0 arg(s)")
	}

	flags, err := loadBumpTagFlags(cmd)
	if err != nil {
		return err
	}

	bumpType, err := resolveBumpComponent(cmd, flags.major, flags.minor, flags.patch)
	if err != nil {
		return err
	}

	svc := newTagService()
	if err := svc.EnsureClean(); err != nil {
		return err
	}

	tag, err := svc.NextTag(bumpType)
	if err != nil {
		return err
	}

	if flags.dryRun {
		if flags.push {
			fmt.Printf("would create and push tag %s\n", tag)
		} else {
			fmt.Printf("would create tag %s\n", tag)
		}
		return nil
	}

	if err := svc.CreateAnnotatedTag(tag); err != nil {
		return err
	}
	fmt.Printf("created tag %s\n", tag)

	if flags.push {
		if err := svc.PushTag(tag); err != nil {
			return fmt.Errorf("tag %s created locally but push failed: %w", tag, err)
		}
		fmt.Printf("pushed tag %s to origin\n", tag)
	}

	return nil
}

func runSetTag(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("requires exactly 1 arg(s)")
	}

	flags, err := loadSetTagFlags(cmd)
	if err != nil {
		return err
	}

	svc := newTagService()
	if err := svc.EnsureClean(); err != nil {
		return err
	}

	tag, err := svc.NormalizeTag(args[0])
	if err != nil {
		return err
	}

	if flags.dryRun {
		if flags.push {
			fmt.Printf("would create and push tag %s\n", tag)
		} else {
			fmt.Printf("would create tag %s\n", tag)
		}
		return nil
	}

	if err := svc.CreateAnnotatedTag(tag); err != nil {
		return err
	}
	fmt.Printf("created tag %s\n", tag)

	if flags.push {
		if err := svc.PushTag(tag); err != nil {
			return fmt.Errorf("tag %s created locally but push failed: %w", tag, err)
		}
		fmt.Printf("pushed tag %s to origin\n", tag)
	}

	return nil
}

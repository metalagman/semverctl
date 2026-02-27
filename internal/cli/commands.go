package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

type jsonFileResultSummary struct {
	Changed   int `json:"changed"`
	Unchanged int `json:"unchanged"`
	Errors    int `json:"errors"`
}

type jsonCommandOutput struct {
	OK            bool                   `json:"ok"`
	Command       string                 `json:"command"`
	DryRun        bool                   `json:"dry_run"`
	Error         string                 `json:"error,omitempty"`
	File          string                 `json:"file,omitempty"`
	Path          string                 `json:"path,omitempty"`
	Bump          string                 `json:"bump,omitempty"`
	Numeric       bool                   `json:"numeric,omitempty"`
	TargetVersion string                 `json:"target_version,omitempty"`
	OldVersion    string                 `json:"old_version,omitempty"`
	NewVersion    string                 `json:"new_version,omitempty"`
	Result        *jsonFileResultSummary `json:"result,omitempty"`
	Tag           string                 `json:"tag,omitempty"`
	Version       string                 `json:"version,omitempty"`
	Push          bool                   `json:"push,omitempty"`
	Created       bool                   `json:"created,omitempty"`
	Pushed        bool                   `json:"pushed,omitempty"`
	Action        string                 `json:"action,omitempty"`
}

func emitJSON(output jsonCommandOutput) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(output)
}

func emitJSONError(output jsonCommandOutput, err error) error {
	output.OK = false
	output.Error = err.Error()
	emitJSON(output)
	return err
}

func jsonRequested(cmd *cobra.Command) bool {
	enabled, err := cmd.Flags().GetBool("json")
	return err == nil && enabled
}

func summarizeResults(results []app.Result) jsonFileResultSummary {
	summary := jsonFileResultSummary{}
	for _, result := range results {
		switch {
		case result.Error != nil:
			summary.Errors++
		case result.Changed:
			summary.Changed++
		default:
			summary.Unchanged++
		}
	}
	return summary
}

func bestResultError(results []app.Result, fallback error) error {
	for _, result := range results {
		if result.Error != nil {
			return result.Error
		}
	}
	return fallback
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
  semverctl bump file config.json --numeric --path .version.Patch
  semverctl bump file package.json --dry-run --json`,
		Args: ExactNArgsWithHelp(1),
		RunE: runBumpFile,
	}
	bumpFileCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	bumpFileCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")
	bumpFileCmd.Flags().Bool("major", false, "Bump major version")
	bumpFileCmd.Flags().Bool("minor", false, "Bump minor version")
	bumpFileCmd.Flags().Bool("patch", false, "Bump patch version (default)")
	bumpFileCmd.Flags().Bool("numeric", false, "Bump numeric scalar by +1 (for object-style versions)")
	bumpFileCmd.Flags().Bool("json", false, "Output machine-readable JSON")

	bumpTagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Bump semantic version using git tags",
		Long: `Create the next git release tag from the highest stable existing tag.

By default this command creates the local annotated tag. Use --dry-run to preview.

Examples:
  semverctl bump tag
  semverctl bump tag --minor
  semverctl bump tag --dry-run
  semverctl bump tag --push
  semverctl bump tag --dry-run --json`,
		Args: ExactNArgsWithHelp(0),
		RunE: runBumpTag,
	}
	bumpTagCmd.Flags().Bool("major", false, "Bump major version")
	bumpTagCmd.Flags().Bool("minor", false, "Bump minor version")
	bumpTagCmd.Flags().Bool("patch", false, "Bump patch version (default)")
	bumpTagCmd.Flags().Bool("push", false, "Push created tag to origin")
	bumpTagCmd.Flags().Bool("dry-run", false, "Preview resulting tag without creating it")
	bumpTagCmd.Flags().Bool("json", false, "Output machine-readable JSON")

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
  semverctl set file 2.0.0 config.yaml --path .app.version
  semverctl set file 1.2.3 package.json --json`,
		Args: ExactNArgsWithHelp(2),
		RunE: runSetFile,
	}
	setFileCmd.Flags().String("path", "version", "Dot-path to version field (e.g., .version, .app.version)")
	setFileCmd.Flags().Bool("dry-run", false, "Show diffs without modifying files")
	setFileCmd.Flags().Bool("json", false, "Output machine-readable JSON")

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
  semverctl set tag 1.2.3 --push
  semverctl set tag 1.2.3 --json`,
		Args: ExactNArgsWithHelp(1),
		RunE: runSetTag,
	}
	setTagCmd.Flags().Bool("push", false, "Push created tag to origin")
	setTagCmd.Flags().Bool("dry-run", false, "Preview resulting tag without creating it")
	setTagCmd.Flags().Bool("json", false, "Output machine-readable JSON")

	setCmd.AddCommand(setFileCmd, setTagCmd)
	rootCmd.AddCommand(setCmd)
}

type bumpFileFlags struct {
	path        string
	dryRun      bool
	json        bool
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
	if f.json, err = cmd.Flags().GetBool("json"); err != nil {
		return f, fmt.Errorf("get --json: %w", err)
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
	json   bool
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
	if f.json, err = cmd.Flags().GetBool("json"); err != nil {
		return f, fmt.Errorf("get --json: %w", err)
	}

	return f, nil
}

type bumpTagFlags struct {
	major  bool
	minor  bool
	patch  bool
	push   bool
	dryRun bool
	json   bool
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
	if f.json, err = cmd.Flags().GetBool("json"); err != nil {
		return f, fmt.Errorf("get --json: %w", err)
	}

	return f, nil
}

type setTagFlags struct {
	push   bool
	dryRun bool
	json   bool
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
	if f.json, err = cmd.Flags().GetBool("json"); err != nil {
		return f, fmt.Errorf("get --json: %w", err)
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
		err := fmt.Errorf("requires exactly 1 arg(s)")
		if jsonRequested(cmd) {
			return emitJSONError(jsonCommandOutput{Command: "bump file"}, err)
		}
		return err
	}

	flags, err := loadBumpFileFlags(cmd)
	if err != nil {
		return err
	}

	jsonOut := jsonCommandOutput{
		Command: "bump file",
		DryRun:  flags.dryRun,
		File:    args[0],
		Numeric: flags.numericBump,
	}

	if flags.numericBump {
		if flags.major || flags.minor || flags.patch {
			err := fmt.Errorf("cannot use --major/--minor/--patch with --numeric")
			if flags.json {
				return emitJSONError(jsonOut, err)
			}
			return err
		}
	}

	bumpType := ""
	if !flags.numericBump {
		bumpType, err = resolveBumpComponent(cmd, flags.major, flags.minor, flags.patch)
		if err != nil {
			if flags.json {
				return emitJSONError(jsonOut, err)
			}
			return err
		}
		jsonOut.Bump = bumpType
	}

	path, err := pathx.Parse(flags.path)
	if err != nil {
		err = fmt.Errorf("invalid path %q: %w", flags.path, err)
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Path = app.JoinPath(path)

	config := &app.Config{
		Operation:   app.OpBump,
		BumpType:    bumpType,
		Path:        path,
		File:        args[0],
		DryRun:      flags.dryRun,
		NumericBump: flags.numericBump,
	}

	runner := app.NewRunner(config)
	if !flags.json {
		return runner.Run()
	}

	results, runErr := runner.RunWithResults()
	if len(results) > 0 {
		jsonOut.File = results[0].File
		jsonOut.OldVersion = results[0].OldVer
		jsonOut.NewVersion = results[0].NewVer
	}
	summary := summarizeResults(results)
	jsonOut.Result = &summary

	if runErr != nil {
		return emitJSONError(jsonOut, bestResultError(results, runErr))
	}

	jsonOut.OK = true
	emitJSON(jsonOut)
	return nil
}

func runSetFile(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		err := fmt.Errorf("requires exactly 2 arg(s)")
		if jsonRequested(cmd) {
			return emitJSONError(jsonCommandOutput{Command: "set file"}, err)
		}
		return err
	}

	flags, err := loadSetFileFlags(cmd)
	if err != nil {
		return err
	}

	jsonOut := jsonCommandOutput{
		Command:       "set file",
		DryRun:        flags.dryRun,
		File:          args[1],
		TargetVersion: args[0],
	}

	path, err := pathx.Parse(flags.path)
	if err != nil {
		err = fmt.Errorf("invalid path %q: %w", flags.path, err)
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Path = app.JoinPath(path)

	config := &app.Config{
		Operation: app.OpSet,
		Version:   args[0],
		Path:      path,
		File:      args[1],
		DryRun:    flags.dryRun,
	}

	runner := app.NewRunner(config)
	if !flags.json {
		return runner.Run()
	}

	results, runErr := runner.RunWithResults()
	if len(results) > 0 {
		jsonOut.File = results[0].File
		jsonOut.OldVersion = results[0].OldVer
		jsonOut.NewVersion = results[0].NewVer
	}
	summary := summarizeResults(results)
	jsonOut.Result = &summary

	if runErr != nil {
		return emitJSONError(jsonOut, bestResultError(results, runErr))
	}

	jsonOut.OK = true
	emitJSON(jsonOut)
	return nil
}

func runBumpTag(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		err := fmt.Errorf("requires exactly 0 arg(s)")
		if jsonRequested(cmd) {
			return emitJSONError(jsonCommandOutput{Command: "bump tag"}, err)
		}
		return err
	}

	flags, err := loadBumpTagFlags(cmd)
	if err != nil {
		return err
	}

	bumpType, err := resolveBumpComponent(cmd, flags.major, flags.minor, flags.patch)
	if err != nil {
		return err
	}

	jsonOut := jsonCommandOutput{
		Command: "bump tag",
		DryRun:  flags.dryRun,
		Push:    flags.push,
		Bump:    bumpType,
	}

	svc := newTagService()
	if err := svc.EnsureClean(); err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}

	tag, err := svc.NextTag(bumpType)
	if err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Tag = tag
	jsonOut.Version = strings.TrimPrefix(tag, "v")

	if flags.dryRun {
		jsonOut.Action = "planned"
		if flags.json {
			jsonOut.OK = true
			emitJSON(jsonOut)
			return nil
		}
		if flags.push {
			fmt.Printf("would create and push tag %s\n", tag)
		} else {
			fmt.Printf("would create tag %s\n", tag)
		}
		return nil
	}

	if err := svc.CreateAnnotatedTag(tag); err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Created = true
	jsonOut.Action = "created"
	if !flags.json {
		fmt.Printf("created tag %s\n", tag)
	}

	if flags.push {
		if err := svc.PushTag(tag); err != nil {
			err = fmt.Errorf("tag %s created locally but push failed: %w", tag, err)
			if flags.json {
				return emitJSONError(jsonOut, err)
			}
			return err
		}
		jsonOut.Pushed = true
		jsonOut.Action = "created_and_pushed"
		if !flags.json {
			fmt.Printf("pushed tag %s to origin\n", tag)
		}
	}

	if flags.json {
		jsonOut.OK = true
		emitJSON(jsonOut)
		return nil
	}

	return nil
}

func runSetTag(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		err := fmt.Errorf("requires exactly 1 arg(s)")
		if jsonRequested(cmd) {
			return emitJSONError(jsonCommandOutput{Command: "set tag"}, err)
		}
		return err
	}

	flags, err := loadSetTagFlags(cmd)
	if err != nil {
		return err
	}

	jsonOut := jsonCommandOutput{
		Command: "set tag",
		DryRun:  flags.dryRun,
		Push:    flags.push,
	}

	svc := newTagService()
	if err := svc.EnsureClean(); err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}

	tag, err := svc.NormalizeTag(args[0])
	if err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Tag = tag
	jsonOut.Version = strings.TrimPrefix(tag, "v")

	if flags.dryRun {
		jsonOut.Action = "planned"
		if flags.json {
			jsonOut.OK = true
			emitJSON(jsonOut)
			return nil
		}
		if flags.push {
			fmt.Printf("would create and push tag %s\n", tag)
		} else {
			fmt.Printf("would create tag %s\n", tag)
		}
		return nil
	}

	if err := svc.CreateAnnotatedTag(tag); err != nil {
		if flags.json {
			return emitJSONError(jsonOut, err)
		}
		return err
	}
	jsonOut.Created = true
	jsonOut.Action = "created"
	if !flags.json {
		fmt.Printf("created tag %s\n", tag)
	}

	if flags.push {
		if err := svc.PushTag(tag); err != nil {
			err = fmt.Errorf("tag %s created locally but push failed: %w", tag, err)
			if flags.json {
				return emitJSONError(jsonOut, err)
			}
			return err
		}
		jsonOut.Pushed = true
		jsonOut.Action = "created_and_pushed"
		if !flags.json {
			fmt.Printf("pushed tag %s to origin\n", tag)
		}
	}

	if flags.json {
		jsonOut.OK = true
		emitJSON(jsonOut)
		return nil
	}

	return nil
}

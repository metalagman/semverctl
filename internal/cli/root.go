package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
		version = info.Main.Version
	}
}

var rootCmd = &cobra.Command{
	Use:   "semverctl",
	Short: "CLI for bumping and setting SemVer values in JSON/YAML files",
	Long: `semverctl is a CLI tool for managing semantic version values in JSON and YAML files.

It supports bumping versions (major, minor, patch) and setting explicit version values
using dot-path selectors (e.g., .version, .app.version).

Examples:
  semverctl bump package.json
  semverctl bump --path .app.version config.yaml
  semverctl set 1.2.3 --file package.json`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("semverctl version %s (git: %s, built: %s)\n", version, gitCommit, buildDate)
	},
}

// GetVersion returns the current version string.
func GetVersion() string {
	return version
}

// GetGitCommit returns the git commit hash.
func GetGitCommit() string {
	return gitCommit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return buildDate
}

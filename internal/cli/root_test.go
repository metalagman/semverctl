package cli

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Test that GetVersion returns the expected value
	v := GetVersion()
	if v != version {
		t.Errorf("GetVersion() = %v, want %v", v, version)
	}
}

func TestGetGitCommit(t *testing.T) {
	v := GetGitCommit()
	if v != gitCommit {
		t.Errorf("GetGitCommit() = %v, want %v", v, gitCommit)
	}
}

func TestGetBuildDate(t *testing.T) {
	v := GetBuildDate()
	if v != buildDate {
		t.Errorf("GetBuildDate() = %v, want %v", v, buildDate)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic
	// This is a basic smoke test - in production, you'd test actual command execution
	// but that requires more setup (temp files, etc.)

	// We can't easily test Execute() without mocking, but we can ensure
	// the command structure is valid by checking that the root command exists
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}
}

func TestRootCmd(t *testing.T) {
	// Verify command structure
	if rootCmd.Use != "semverctl" {
		t.Errorf("rootCmd.Use = %v, want semverctl", rootCmd.Use)
	}

	// Check that subcommands exist
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Error("rootCmd should have subcommands")
	}

	// Verify version command exists
	hasVersion := false
	hasBump := false
	hasSet := false
	for _, cmd := range subcommands {
		switch cmd.Name() {
		case "version":
			hasVersion = true
		case "bump":
			hasBump = true
		case "set":
			hasSet = true
		}
	}

	if !hasVersion {
		t.Error("rootCmd should have 'version' subcommand")
	}
	if !hasBump {
		t.Error("rootCmd should have 'bump' subcommand")
	}
	if !hasSet {
		t.Error("rootCmd should have 'set' subcommand")
	}
}

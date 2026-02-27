package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func setFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set flag %q=%q: %v", name, value, err)
	}
}

func newBumpFileTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("numeric", false, "")
	return cmd
}

func newSetFileTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

func newBumpTagTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("push", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

func newSetTagTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("push", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

type fakeTagService struct {
	ensureCleanErr error
	nextTag        string
	nextTagErr     error
	normalizedTag  string
	normalizeErr   error
	createErr      error
	pushErr        error
	created        []string
	pushed         []string
}

func (f *fakeTagService) EnsureClean() error { return f.ensureCleanErr }

func (f *fakeTagService) NextTag(component string) (string, error) {
	if f.nextTagErr != nil {
		return "", f.nextTagErr
	}
	return f.nextTag, nil
}

func (f *fakeTagService) NormalizeTag(versionOrTag string) (string, error) {
	if f.normalizeErr != nil {
		return "", f.normalizeErr
	}
	return f.normalizedTag, nil
}

func (f *fakeTagService) CreateAnnotatedTag(tag string) error {
	f.created = append(f.created, tag)
	return f.createErr
}

func (f *fakeTagService) PushTag(tag string) error {
	f.pushed = append(f.pushed, tag)
	return f.pushErr
}

func useFakeTagService(t *testing.T, svc *fakeTagService) {
	t.Helper()
	old := newTagService
	newTagService = func() tagService { return svc }
	t.Cleanup(func() { newTagService = old })
}

func TestRunBumpFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	t.Run("dry run success", func(t *testing.T) {
		cmd := newBumpFileTestCmd()
		setFlag(t, cmd, "dry-run", "true")

		if err := runBumpFile(cmd, []string{testFile}); err != nil {
			t.Fatalf("runBumpFile() error = %v", err)
		}
	})

	t.Run("numeric with bump type", func(t *testing.T) {
		cmd := newBumpFileTestCmd()
		setFlag(t, cmd, "numeric", "true")
		setFlag(t, cmd, "patch", "true")

		if err := runBumpFile(cmd, []string{testFile}); err == nil {
			t.Fatal("runBumpFile() should fail when --numeric and bump type are combined")
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		cmd := newBumpFileTestCmd()
		setFlag(t, cmd, "path", "..")

		if err := runBumpFile(cmd, []string{testFile}); err == nil {
			t.Fatal("runBumpFile() should fail on invalid path")
		}
	})

	t.Run("requires file arg", func(t *testing.T) {
		cmd := newBumpFileTestCmd()
		if err := runBumpFile(cmd, nil); err == nil {
			t.Fatal("runBumpFile() should require one file arg")
		}
	})
}

func TestRunSetFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	t.Run("dry run success", func(t *testing.T) {
		cmd := newSetFileTestCmd()
		setFlag(t, cmd, "dry-run", "true")

		if err := runSetFile(cmd, []string{"1.2.3", testFile}); err != nil {
			t.Fatalf("runSetFile() error = %v", err)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		cmd := newSetFileTestCmd()
		setFlag(t, cmd, "path", "..")

		if err := runSetFile(cmd, []string{"1.2.3", testFile}); err == nil {
			t.Fatal("runSetFile() should fail on invalid path")
		}
	})

	t.Run("requires two args", func(t *testing.T) {
		cmd := newSetFileTestCmd()
		if err := runSetFile(cmd, []string{"1.2.3"}); err == nil {
			t.Fatal("runSetFile() should require VERSION and PATH")
		}
	})
}

func TestRunBumpTag(t *testing.T) {
	t.Run("create by default", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4"}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		if err := runBumpTag(cmd, nil); err != nil {
			t.Fatalf("runBumpTag() error = %v", err)
		}

		if len(svc.created) != 1 || svc.created[0] != "v1.2.4" {
			t.Fatalf("created tags = %v, want [v1.2.4]", svc.created)
		}
		if len(svc.pushed) != 0 {
			t.Fatalf("pushed tags = %v, want none", svc.pushed)
		}
	})

	t.Run("dry run does not create", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4"}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "dry-run", "true")

		if err := runBumpTag(cmd, nil); err != nil {
			t.Fatalf("runBumpTag() error = %v", err)
		}
		if len(svc.created) != 0 || len(svc.pushed) != 0 {
			t.Fatalf("dry-run should not mutate, created=%v pushed=%v", svc.created, svc.pushed)
		}
	})

	t.Run("push", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4"}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "push", "true")

		if err := runBumpTag(cmd, nil); err != nil {
			t.Fatalf("runBumpTag() error = %v", err)
		}
		if len(svc.created) != 1 || len(svc.pushed) != 1 {
			t.Fatalf("expected one create and one push, created=%v pushed=%v", svc.created, svc.pushed)
		}
	})

	t.Run("clean repo required", func(t *testing.T) {
		svc := &fakeTagService{ensureCleanErr: errors.New("dirty")}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		if err := runBumpTag(cmd, nil); err == nil {
			t.Fatal("runBumpTag() should fail when repo is dirty")
		}
	})

	t.Run("bump flags conflict", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4"}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "major", "true")
		setFlag(t, cmd, "minor", "true")

		if err := runBumpTag(cmd, nil); err == nil {
			t.Fatal("runBumpTag() should fail on conflicting bump flags")
		}
	})
}

func TestRunSetTag(t *testing.T) {
	t.Run("create by default", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3"}
		useFakeTagService(t, svc)

		cmd := newSetTagTestCmd()
		if err := runSetTag(cmd, []string{"1.2.3"}); err != nil {
			t.Fatalf("runSetTag() error = %v", err)
		}

		if len(svc.created) != 1 || svc.created[0] != "v1.2.3" {
			t.Fatalf("created tags = %v, want [v1.2.3]", svc.created)
		}
	})

	t.Run("dry run does not create", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3"}
		useFakeTagService(t, svc)

		cmd := newSetTagTestCmd()
		setFlag(t, cmd, "dry-run", "true")

		if err := runSetTag(cmd, []string{"1.2.3"}); err != nil {
			t.Fatalf("runSetTag() error = %v", err)
		}
		if len(svc.created) != 0 || len(svc.pushed) != 0 {
			t.Fatalf("dry-run should not mutate, created=%v pushed=%v", svc.created, svc.pushed)
		}
	})

	t.Run("push", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3"}
		useFakeTagService(t, svc)

		cmd := newSetTagTestCmd()
		setFlag(t, cmd, "push", "true")

		if err := runSetTag(cmd, []string{"1.2.3"}); err != nil {
			t.Fatalf("runSetTag() error = %v", err)
		}
		if len(svc.created) != 1 || len(svc.pushed) != 1 {
			t.Fatalf("expected one create and one push, created=%v pushed=%v", svc.created, svc.pushed)
		}
	})

	t.Run("normalize error", func(t *testing.T) {
		svc := &fakeTagService{normalizeErr: errors.New("invalid version")}
		useFakeTagService(t, svc)

		cmd := newSetTagTestCmd()
		if err := runSetTag(cmd, []string{"nope"}); err == nil {
			t.Fatal("runSetTag() should fail on invalid version")
		}
	})

	t.Run("requires one arg", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3"}
		useFakeTagService(t, svc)

		cmd := newSetTagTestCmd()
		if err := runSetTag(cmd, nil); err == nil {
			t.Fatal("runSetTag() should require VERSION")
		}
	})
}

func TestLoadBumpFileFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Int("path", 0, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("numeric", false, "")

	if _, err := loadBumpFileFlags(cmd); err == nil {
		t.Fatal("loadBumpFileFlags() should error when path flag has wrong type")
	}
}

func TestLoadSetFileFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("path", "version", "")
	cmd.Flags().String("dry-run", "false", "")

	if _, err := loadSetFileFlags(cmd); err == nil {
		t.Fatal("loadSetFileFlags() should error when dry-run flag has wrong type")
	}
}

func TestCommandTree(t *testing.T) {
	var bumpCmd *cobra.Command
	var setCmd *cobra.Command

	for _, cmd := range rootCmd.Commands() {
		switch cmd.Name() {
		case "bump":
			bumpCmd = cmd
		case "set":
			setCmd = cmd
		}
	}

	if bumpCmd == nil {
		t.Fatal("bump command not found")
	}
	if setCmd == nil {
		t.Fatal("set command not found")
	}

	hasBumpFile := false
	hasBumpTag := false
	for _, cmd := range bumpCmd.Commands() {
		switch cmd.Name() {
		case "file":
			hasBumpFile = true
		case "tag":
			hasBumpTag = true
		}
	}
	if !hasBumpFile || !hasBumpTag {
		t.Fatalf("bump subcommands missing: file=%v tag=%v", hasBumpFile, hasBumpTag)
	}

	hasSetFile := false
	hasSetTag := false
	for _, cmd := range setCmd.Commands() {
		switch cmd.Name() {
		case "file":
			hasSetFile = true
		case "tag":
			hasSetTag = true
		}
	}
	if !hasSetFile || !hasSetTag {
		t.Fatalf("set subcommands missing: file=%v tag=%v", hasSetFile, hasSetTag)
	}
}

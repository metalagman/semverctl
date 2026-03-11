package cli

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/metalagman/semverctl/internal/app"
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
	cmd.Flags().Bool("json", false, "")
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
	cmd.Flags().Bool("json", false, "")
	return cmd
}

func newBumpTagTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("major", false, "")
	cmd.Flags().Bool("minor", false, "")
	cmd.Flags().Bool("patch", false, "")
	cmd.Flags().Bool("push", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("json", false, "")
	return cmd
}

func newSetTagTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("push", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("json", false, "")
	return cmd
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stdout = w

	runErr := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close write pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close read pipe: %v", err)
	}
	os.Stdout = oldStdout

	return string(out), runErr
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

func TestRunBumpFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"version": "1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	cmd := newBumpFileTestCmd()
	setFlag(t, cmd, "dry-run", "true")
	setFlag(t, cmd, "json", "true")

	out, err := captureStdout(t, func() error {
		return runBumpFile(cmd, []string{testFile})
	})
	if err != nil {
		t.Fatalf("runBumpFile() error = %v", err)
	}

	var payload jsonCommandOutput
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out)
	}

	if !payload.OK || payload.Command != "bump file" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.NewVersion != "1.0.1" {
		t.Fatalf("payload.NewVersion = %q, want 1.0.1", payload.NewVersion)
	}
	if payload.Result == nil || payload.Result.Changed != 1 {
		t.Fatalf("payload.Result = %+v, want changed=1", payload.Result)
	}
}

func TestRunBumpFile_ZeroMajorResets(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		flag string
		want string
	}{
		{name: "minor bump resets patch", flag: "minor", want: "0.1.0"},
		{name: "major bump resets minor and patch", flag: "major", want: "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "_")+".json")
			if err := os.WriteFile(testFile, []byte(`{"version": "0.0.1"}`), 0o644); err != nil {
				t.Fatalf("write test file: %v", err)
			}

			cmd := newBumpFileTestCmd()
			setFlag(t, cmd, tt.flag, "true")

			if err := runBumpFile(cmd, []string{testFile}); err != nil {
				t.Fatalf("runBumpFile() error = %v", err)
			}

			got, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("read bumped file: %v", err)
			}

			if !strings.Contains(string(got), `"`+tt.want+`"`) {
				t.Fatalf("bumped file content = %q, want version %q", string(got), tt.want)
			}
		})
	}
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

func TestRunSetFile_JSONError(t *testing.T) {
	cmd := newSetFileTestCmd()
	setFlag(t, cmd, "json", "true")
	setFlag(t, cmd, "dry-run", "true")

	out, err := captureStdout(t, func() error {
		return runSetFile(cmd, []string{"1.2.3", "missing.json"})
	})
	if err == nil {
		t.Fatal("runSetFile() should return error for missing file")
	}

	var payload jsonCommandOutput
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out)
	}
	if payload.OK {
		t.Fatalf("payload.OK = true, want false: %+v", payload)
	}
	if payload.Error == "" {
		t.Fatalf("payload.Error should be set: %+v", payload)
	}
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

func TestRunBumpTag_JSON(t *testing.T) {
	t.Run("dry run payload", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4"}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "dry-run", "true")
		setFlag(t, cmd, "json", "true")

		out, err := captureStdout(t, func() error {
			return runBumpTag(cmd, nil)
		})
		if err != nil {
			t.Fatalf("runBumpTag() error = %v", err)
		}

		var payload jsonCommandOutput
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out)
		}
		if !payload.OK || payload.Tag != "v1.2.4" || payload.Version != "1.2.4" || payload.Action != "planned" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
	})

	t.Run("error payload", func(t *testing.T) {
		svc := &fakeTagService{ensureCleanErr: errors.New("dirty repo")}
		useFakeTagService(t, svc)

		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "json", "true")

		out, err := captureStdout(t, func() error {
			return runBumpTag(cmd, nil)
		})
		if err == nil {
			t.Fatal("runBumpTag() should return error")
		}

		var payload jsonCommandOutput
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out)
		}
		if payload.OK || payload.Error == "" {
			t.Fatalf("unexpected payload: %+v", payload)
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

func TestRunSetTag_JSON(t *testing.T) {
	svc := &fakeTagService{normalizedTag: "v1.2.3"}
	useFakeTagService(t, svc)

	cmd := newSetTagTestCmd()
	setFlag(t, cmd, "dry-run", "true")
	setFlag(t, cmd, "json", "true")

	out, err := captureStdout(t, func() error {
		return runSetTag(cmd, []string{"1.2.3"})
	})
	if err != nil {
		t.Fatalf("runSetTag() error = %v", err)
	}

	var payload jsonCommandOutput
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", err, out)
	}
	if !payload.OK || payload.Tag != "v1.2.3" || payload.Version != "1.2.3" || payload.Action != "planned" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestExactNArgsWithHelp(t *testing.T) {
	cmd := &cobra.Command{}
	validator := ExactNArgsWithHelp(2)
	if err := validator(cmd, []string{"a", "b"}); err != nil {
		t.Errorf("ExactNArgsWithHelp() error = %v", err)
	}
	if err := validator(cmd, []string{"a"}); err == nil {
		t.Fatal("ExactNArgsWithHelp() should fail with 1 arg")
	}
}

func TestBestResultError(t *testing.T) {
	err1 := errors.New("err1")
	fallback := errors.New("fallback")

	results := []app.Result{
		{File: "f1", Error: nil},
		{File: "f2", Error: err1},
	}

	if err := bestResultError(results, fallback); err != err1 {
		t.Errorf("bestResultError() = %v, want %v", err, err1)
	}

	resultsClean := []app.Result{
		{File: "f1", Error: nil},
	}
	if err := bestResultError(resultsClean, fallback); err != fallback {
		t.Errorf("bestResultError() = %v, want %v", err, fallback)
	}
}

func TestLoadBumpTagFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("major", "true", "")
	if _, err := loadBumpTagFlags(cmd); err == nil {
		t.Fatal("loadBumpTagFlags() should error when major flag has wrong type")
	}
}

func TestLoadSetTagFlagsErrors(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("push", "true", "")
	if _, err := loadSetTagFlags(cmd); err == nil {
		t.Fatal("loadSetTagFlags() should error when push flag has wrong type")
	}
}

func TestRunBumpTag_Errors(t *testing.T) {
	t.Run("NextTag error", func(t *testing.T) {
		svc := &fakeTagService{nextTagErr: errors.New("next tag error")}
		useFakeTagService(t, svc)
		cmd := newBumpTagTestCmd()
		if err := runBumpTag(cmd, nil); err == nil {
			t.Fatal("runBumpTag() should fail on NextTag error")
		}
	})

	t.Run("CreateAnnotatedTag error", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4", createErr: errors.New("create error")}
		useFakeTagService(t, svc)
		cmd := newBumpTagTestCmd()
		if err := runBumpTag(cmd, nil); err == nil {
			t.Fatal("runBumpTag() should fail on CreateAnnotatedTag error")
		}
	})

	t.Run("PushTag error", func(t *testing.T) {
		svc := &fakeTagService{nextTag: "v1.2.4", pushErr: errors.New("push error")}
		useFakeTagService(t, svc)
		cmd := newBumpTagTestCmd()
		setFlag(t, cmd, "push", "true")
		if err := runBumpTag(cmd, nil); err == nil {
			t.Fatal("runBumpTag() should fail on PushTag error")
		}
	})
}

func TestRunSetTag_Errors(t *testing.T) {
	t.Run("CreateAnnotatedTag error", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3", createErr: errors.New("create error")}
		useFakeTagService(t, svc)
		cmd := newSetTagTestCmd()
		if err := runSetTag(cmd, []string{"1.2.3"}); err == nil {
			t.Fatal("runSetTag() should fail on CreateAnnotatedTag error")
		}
	})

	t.Run("PushTag error", func(t *testing.T) {
		svc := &fakeTagService{normalizedTag: "v1.2.3", pushErr: errors.New("push error")}
		useFakeTagService(t, svc)
		cmd := newSetTagTestCmd()
		setFlag(t, cmd, "push", "true")
		if err := runSetTag(cmd, []string{"1.2.3"}); err == nil {
			t.Fatal("runSetTag() should fail on PushTag error")
		}
	})
}

func TestRunBumpFile_PathJSONError(t *testing.T) {
	cmd := newBumpFileTestCmd()
	setFlag(t, cmd, "json", "true")
	setFlag(t, cmd, "path", "..")

	out, _ := captureStdout(t, func() error {
		return runBumpFile(cmd, []string{"test.json"})
	})
	var payload jsonCommandOutput
	_ = json.Unmarshal([]byte(out), &payload)
	if payload.OK || payload.Error == "" {
		t.Fatal("runBumpFile() should emit JSON error for invalid path")
	}
}

func TestRunSetFile_PathJSONError(t *testing.T) {
	cmd := newSetFileTestCmd()
	setFlag(t, cmd, "json", "true")
	setFlag(t, cmd, "path", "..")

	out, _ := captureStdout(t, func() error {
		return runSetFile(cmd, []string{"1.2.3", "test.json"})
	})
	var payload jsonCommandOutput
	_ = json.Unmarshal([]byte(out), &payload)
	if payload.OK || payload.Error == "" {
		t.Fatal("runSetFile() should emit JSON error for invalid path")
	}
}

func TestResolveBumpComponent_Multiple(t *testing.T) {
	cmd := newBumpTagTestCmd()
	setFlag(t, cmd, "major", "true")
	setFlag(t, cmd, "minor", "true")

	if _, err := resolveBumpComponent(cmd, true, true, false); err == nil {
		t.Fatal("resolveBumpComponent() should error on multiple bump types")
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

	for _, cmd := range bumpCmd.Commands() {
		if cmd.Flags().Lookup("json") == nil {
			t.Fatalf("bump subcommand %q missing --json flag", cmd.Name())
		}
	}
	for _, cmd := range setCmd.Commands() {
		if cmd.Flags().Lookup("json") == nil {
			t.Fatalf("set subcommand %q missing --json flag", cmd.Name())
		}
	}
}

func TestSummarizeResults(t *testing.T) {
	summary := summarizeResults([]app.Result{
		{Changed: true},
		{Changed: false},
		{Error: errors.New("boom")},
		{Changed: true},
	})

	if summary.Changed != 2 || summary.Unchanged != 1 || summary.Errors != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestLoadBumpFileFlags_Errors(t *testing.T) {
	t.Run("missing path flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error when --path is missing")
		}
	})

	t.Run("dry-run wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().String("dry-run", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on dry-run type mismatch")
		}
	})

	t.Run("json wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().String("json", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on json type mismatch")
		}
	})

	t.Run("major wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().String("major", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on major type mismatch")
		}
	})

	t.Run("minor wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().String("minor", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on minor type mismatch")
		}
	})

	t.Run("patch wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().String("patch", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on patch type mismatch")
		}
	})

	t.Run("numeric wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().Bool("patch", false, "")
		cmd.Flags().String("numeric", "false", "")
		if _, err := loadBumpFileFlags(cmd); err == nil {
			t.Fatal("loadBumpFileFlags() should error on numeric type mismatch")
		}
	})
}

func TestLoadTagAndSetFlags_MissingFlags(t *testing.T) {
	t.Run("loadBumpTagFlags missing", func(t *testing.T) {
		cmd := &cobra.Command{}
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error when required flags are missing")
		}
	})

	t.Run("loadSetTagFlags missing", func(t *testing.T) {
		cmd := &cobra.Command{}
		if _, err := loadSetTagFlags(cmd); err == nil {
			t.Fatal("loadSetTagFlags() should error when required flags are missing")
		}
	})
}

func TestLoadBumpTagFlags_ErrorsByFlag(t *testing.T) {
	t.Run("minor wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().String("minor", "false", "")
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error on minor type mismatch")
		}
	})

	t.Run("patch wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().String("patch", "false", "")
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error on patch type mismatch")
		}
	})

	t.Run("push wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().Bool("patch", false, "")
		cmd.Flags().String("push", "false", "")
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error on push type mismatch")
		}
	})

	t.Run("dry-run wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().Bool("patch", false, "")
		cmd.Flags().Bool("push", false, "")
		cmd.Flags().String("dry-run", "false", "")
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error on dry-run type mismatch")
		}
	})

	t.Run("json wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("major", false, "")
		cmd.Flags().Bool("minor", false, "")
		cmd.Flags().Bool("patch", false, "")
		cmd.Flags().Bool("push", false, "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().String("json", "false", "")
		if _, err := loadBumpTagFlags(cmd); err == nil {
			t.Fatal("loadBumpTagFlags() should error on json type mismatch")
		}
	})
}

func TestLoadSetTagFlags_ErrorsByFlag(t *testing.T) {
	t.Run("dry-run wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("push", false, "")
		cmd.Flags().String("dry-run", "false", "")
		if _, err := loadSetTagFlags(cmd); err == nil {
			t.Fatal("loadSetTagFlags() should error on dry-run type mismatch")
		}
	})

	t.Run("json wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("push", false, "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().String("json", "false", "")
		if _, err := loadSetTagFlags(cmd); err == nil {
			t.Fatal("loadSetTagFlags() should error on json type mismatch")
		}
	})
}

func TestLoadSetFileFlags_ErrorsByFlag(t *testing.T) {
	t.Run("missing path", func(t *testing.T) {
		cmd := &cobra.Command{}
		if _, err := loadSetFileFlags(cmd); err == nil {
			t.Fatal("loadSetFileFlags() should error when --path is missing")
		}
	})

	t.Run("json wrong type", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("path", "version", "")
		cmd.Flags().Bool("dry-run", false, "")
		cmd.Flags().String("json", "false", "")
		if _, err := loadSetFileFlags(cmd); err == nil {
			t.Fatal("loadSetFileFlags() should error on json type mismatch")
		}
	})
}

func TestRunSetFile_ArgCountJSONError(t *testing.T) {
	cmd := newSetFileTestCmd()
	setFlag(t, cmd, "json", "true")

	out, err := captureStdout(t, func() error {
		return runSetFile(cmd, []string{"1.2.3"})
	})
	if err == nil {
		t.Fatal("runSetFile() should error on missing PATH argument")
	}
	var payload jsonCommandOutput
	if unmarshalErr := json.Unmarshal([]byte(out), &payload); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", unmarshalErr, out)
	}
	if payload.OK || payload.Error == "" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestRunBumpFile_ArgCountJSONError(t *testing.T) {
	cmd := newBumpFileTestCmd()
	setFlag(t, cmd, "json", "true")

	out, err := captureStdout(t, func() error {
		return runBumpFile(cmd, nil)
	})
	if err == nil {
		t.Fatal("runBumpFile() should error on missing PATH argument")
	}
	var payload jsonCommandOutput
	if unmarshalErr := json.Unmarshal([]byte(out), &payload); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal() error = %v, out=%q", unmarshalErr, out)
	}
	if payload.OK || payload.Error == "" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

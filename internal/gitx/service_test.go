package gitx

import (
	"errors"
	"testing"

	"github.com/metalagman/semverctl/internal/semverx"
)

type fakeRunner struct {
	responses map[string]fakeResponse
}

type fakeResponse struct {
	out string
	err error
}

func (f fakeRunner) Run(args ...string) (string, error) {
	key := joinArgs(args)
	response, ok := f.responses[key]
	if !ok {
		return "", errors.New("unexpected git call: " + key)
	}
	return response.out, response.err
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	out := args[0]
	for i := 1; i < len(args); i++ {
		out += "\x00" + args[i]
	}
	return out
}

func TestNormalizeTag(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain version", input: "1.2.3", want: "v1.2.3"},
		{name: "tag version", input: "v1.2.3", want: "v1.2.3"},
		{name: "invalid", input: "invalid", wantErr: true},
		{name: "prerelease not allowed", input: "1.2.3-rc.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.NormalizeTag(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeTag() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != "" && got != tt.want {
				t.Fatalf("NormalizeTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureClean(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"status", "--porcelain"}): {out: "\n"},
		}})
		if err := svc.EnsureClean(); err != nil {
			t.Fatalf("EnsureClean() error = %v", err)
		}
	})

	t.Run("dirty", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"status", "--porcelain"}): {out: " M README.md\n"},
		}})
		if err := svc.EnsureClean(); err == nil {
			t.Fatal("EnsureClean() should error for dirty repo")
		}
	})
}

func TestLatestStableTag(t *testing.T) {
	svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
		joinArgs([]string{"tag", "--list"}): {out: "v0.0.9\nv1.0.0-rc.1\nv1.2.0\nfoo\nv1.1.9\n"},
	}})

	tag, err := svc.LatestStableTag()
	if err != nil {
		t.Fatalf("LatestStableTag() error = %v", err)
	}
	if tag != "v1.2.0" {
		t.Fatalf("LatestStableTag() = %q, want v1.2.0", tag)
	}
}

func TestLatestStableTag_NoStable(t *testing.T) {
	svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
		joinArgs([]string{"tag", "--list"}): {out: "v1.0.0-rc.1\nfoo\n"},
	}})

	if _, err := svc.LatestStableTag(); err == nil {
		t.Fatal("LatestStableTag() should error when no stable tags exist")
	}
}

func TestNextTag(t *testing.T) {
	svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
		joinArgs([]string{"tag", "--list"}): {out: "v1.2.3\n"},
	}})

	tests := []struct {
		component string
		want      string
	}{
		{component: "patch", want: "v1.2.4"},
		{component: "minor", want: "v1.3.0"},
		{component: "major", want: "v2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.component, func(t *testing.T) {
			got, err := svc.NextTag(tt.component)
			if err != nil {
				t.Fatalf("NextTag() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NextTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextTag_FromZeroMajor(t *testing.T) {
	svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
		joinArgs([]string{"tag", "--list"}): {out: "v0.0.1\n"},
	}})

	tests := []struct {
		component string
		want      string
	}{
		{component: "minor", want: "v0.1.0"},
		{component: "major", want: "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.component, func(t *testing.T) {
			got, err := svc.NextTag(tt.component)
			if err != nil {
				t.Fatalf("NextTag() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NextTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateAnnotatedTag(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list", "v1.2.3"}): {out: "v1.2.3\n"},
		}})

		if err := svc.CreateAnnotatedTag("v1.2.3"); err == nil {
			t.Fatal("CreateAnnotatedTag() should error for existing tag")
		}
	})

	t.Run("create", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list", "v1.2.3"}):                     {out: ""},
			joinArgs([]string{"tag", "-a", "v1.2.3", "-m", "Release v1.2.3"}): {out: ""},
		}})

		if err := svc.CreateAnnotatedTag("v1.2.3"); err != nil {
			t.Fatalf("CreateAnnotatedTag() error = %v", err)
		}
	})
}

func TestPushTag(t *testing.T) {
	svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
		joinArgs([]string{"push", "origin", "v1.2.3"}): {out: ""},
	}})

	if err := svc.PushTag("v1.2.3"); err != nil {
		t.Fatalf("PushTag() error = %v", err)
	}
}

func TestTagExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list", "v1.2.3"}): {out: "v1.2.3\n"},
		}})
		exists, err := svc.TagExists("v1.2.3")
		if err != nil {
			t.Fatalf("TagExists() error = %v", err)
		}
		if !exists {
			t.Fatal("TagExists() should return true")
		}
	})

	t.Run("not exists", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list", "v1.2.3"}): {out: ""},
		}})
		exists, err := svc.TagExists("v1.2.3")
		if err != nil {
			t.Fatalf("TagExists() error = %v", err)
		}
		if exists {
			t.Fatal("TagExists() should return false")
		}
	})

	t.Run("error", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list", "v1.2.3"}): {err: errors.New("git error")},
		}})
		_, err := svc.TagExists("v1.2.3")
		if err == nil {
			t.Fatal("TagExists() should return error")
		}
	})
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"major less", "1.0.0", "2.0.0", -1},
		{"major greater", "2.0.0", "1.0.0", 1},
		{"minor less", "1.1.0", "1.2.0", -1},
		{"minor greater", "1.2.0", "1.1.0", 1},
		{"patch less", "1.1.1", "1.1.2", -1},
		{"patch greater", "1.1.2", "1.1.1", 1},
		{"equal", "1.1.1", "1.1.1", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va, _ := semverx.Parse(tt.a)
			vb, _ := semverx.Parse(tt.b)
			got := compare(va, vb)
			if got != tt.want {
				t.Fatalf("compare(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestServiceErrors(t *testing.T) {
	t.Run("EnsureClean error", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"status", "--porcelain"}): {err: errors.New("git error")},
		}})
		if err := svc.EnsureClean(); err == nil {
			t.Fatal("EnsureClean() should return error")
		}
	})

	t.Run("LatestStableTag error", func(t *testing.T) {
		svc := newServiceWithRunner(fakeRunner{responses: map[string]fakeResponse{
			joinArgs([]string{"tag", "--list"}): {err: errors.New("git error")},
		}})
		if _, err := svc.LatestStableTag(); err == nil {
			t.Fatal("LatestStableTag() should return error")
		}
	})

	t.Run("NormalizeTag empty", func(t *testing.T) {
		svc := NewService()
		if _, err := svc.NormalizeTag("  "); err == nil {
			t.Fatal("NormalizeTag() should return error for empty input")
		}
	})
}

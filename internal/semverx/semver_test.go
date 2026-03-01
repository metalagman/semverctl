package semverx

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "version with prerelease",
			input: "1.2.3-alpha",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"},
		},
		{
			name:  "version with metadata",
			input: "1.2.3+build.123",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Metadata: "build.123"},
		},
		{
			name:  "version with prerelease and metadata",
			input: "1.2.3-alpha+build.123",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Metadata: "build.123"},
		},
		{
			name:  "version with complex prerelease",
			input: "1.2.3-alpha.1.2",
			want:  &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1.2"},
		},
		{
			name:    "invalid - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid - leading zero in major",
			input:   "01.2.3",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid - letters only",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("Parse() = %v, want %v", got, tt.want)
				}
				if got.Prerelease != tt.want.Prerelease {
					t.Errorf("Parse() Prerelease = %v, want %v", got.Prerelease, tt.want.Prerelease)
				}
				if got.Metadata != tt.want.Metadata {
					t.Errorf("Parse() Metadata = %v, want %v", got.Metadata, tt.want.Metadata)
				}
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    *Version
		want string
	}{
		{
			name: "simple",
			v:    &Version{Major: 1, Minor: 2, Patch: 3},
			want: "1.2.3",
		},
		{
			name: "with prerelease",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha"},
			want: "1.2.3-alpha",
		},
		{
			name: "with metadata",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Metadata: "build.123"},
			want: "1.2.3+build.123",
		},
		{
			name: "with both",
			v:    &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Metadata: "build.123"},
			want: "1.2.3-alpha+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_BumpMajor(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Metadata: "build"}
	got := v.BumpMajor()
	want := &Version{Major: 2, Minor: 0, Patch: 0}
	if got.Major != want.Major || got.Minor != want.Minor || got.Patch != want.Patch {
		t.Errorf("BumpMajor() = %v, want %v", got, want)
	}
	if got.Prerelease != "" || got.Metadata != "" {
		t.Errorf("BumpMajor() should clear prerelease and metadata")
	}
}

func TestVersion_BumpMinor(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Metadata: "build"}
	got := v.BumpMinor()
	want := &Version{Major: 1, Minor: 3, Patch: 0}
	if got.Major != want.Major || got.Minor != want.Minor || got.Patch != want.Patch {
		t.Errorf("BumpMinor() = %v, want %v", got, want)
	}
	if got.Prerelease != "" || got.Metadata != "" {
		t.Errorf("BumpMinor() should clear prerelease and metadata")
	}
}

func TestVersion_BumpPatch(t *testing.T) {
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha", Metadata: "build"}
	got := v.BumpPatch()
	want := &Version{Major: 1, Minor: 2, Patch: 4}
	if got.Major != want.Major || got.Minor != want.Minor || got.Patch != want.Patch {
		t.Errorf("BumpPatch() = %v, want %v", got, want)
	}
	if got.Prerelease != "" || got.Metadata != "" {
		t.Errorf("BumpPatch() should clear prerelease and metadata")
	}
}

func TestVersion_Bump(t *testing.T) {
	tests := []struct {
		name      string
		v         *Version
		component string
		want      *Version
		wantErr   bool
	}{
		{
			name:      "major",
			v:         &Version{Major: 1, Minor: 2, Patch: 3},
			component: "major",
			want:      &Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:      "minor",
			v:         &Version{Major: 1, Minor: 2, Patch: 3},
			component: "minor",
			want:      &Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name:      "patch",
			v:         &Version{Major: 1, Minor: 2, Patch: 3},
			component: "patch",
			want:      &Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:      "minor from zero major resets patch",
			v:         &Version{Major: 0, Minor: 0, Patch: 1},
			component: "minor",
			want:      &Version{Major: 0, Minor: 1, Patch: 0},
		},
		{
			name:      "major from zero major resets minor and patch",
			v:         &Version{Major: 0, Minor: 0, Patch: 1},
			component: "major",
			want:      &Version{Major: 1, Minor: 0, Patch: 0},
		},
		{
			name:      "invalid",
			v:         &Version{Major: 1, Minor: 2, Patch: 3},
			component: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.Bump(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("Bump() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "1.2.3", true},
		{"valid with prerelease", "1.2.3-alpha", true},
		{"valid with metadata", "1.2.3+build", true},
		{"invalid", "1.2", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.input); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

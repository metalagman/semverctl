package semverx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semVerRegex = regexp.MustCompile(
	`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
)

// Version represents a semantic version.
type Version struct {
	Major      uint64
	Minor      uint64
	Patch      uint64
	Prerelease string
	Metadata   string
}

// Parse validates and parses a semantic version string.
func Parse(s string) (*Version, error) {
	matches := semVerRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid semantic version: %q", s)
	}

	major, _ := strconv.ParseUint(matches[1], 10, 64)
	minor, _ := strconv.ParseUint(matches[2], 10, 64)
	patch, _ := strconv.ParseUint(matches[3], 10, 64)

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Metadata:   matches[5],
	}, nil
}

// String returns the string representation of the version.
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}
	if v.Metadata != "" {
		result += "+" + v.Metadata
	}
	return result
}

// BumpMajor increments the major version and resets minor/patch, clears prerelease/metadata.
func (v *Version) BumpMajor() *Version {
	return &Version{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
	}
}

// BumpMinor increments the minor version and resets patch, clears prerelease/metadata.
func (v *Version) BumpMinor() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
	}
}

// BumpPatch increments the patch version, clears prerelease/metadata.
func (v *Version) BumpPatch() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

// Bump bumps the version according to the given component.
func (v *Version) Bump(component string) (*Version, error) {
	switch strings.ToLower(component) {
	case "major":
		return v.BumpMajor(), nil
	case "minor":
		return v.BumpMinor(), nil
	case "patch":
		return v.BumpPatch(), nil
	default:
		return nil, fmt.Errorf("invalid bump component: %q (must be major, minor, or patch)", component)
	}
}

// IsValid checks if a string is a valid semantic version.
func IsValid(s string) bool {
	return semVerRegex.MatchString(s)
}

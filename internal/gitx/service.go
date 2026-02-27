package gitx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/metalagman/semverctl/internal/semverx"
)

type gitRunner interface {
	Run(args ...string) (string, error)
}

type execGitRunner struct{}

func (r execGitRunner) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return stdout.String(), nil
}

// Service provides git-tag operations for release flows.
type Service struct {
	runner gitRunner
}

// NewService creates a new git service.
func NewService() *Service {
	return &Service{runner: execGitRunner{}}
}

func newServiceWithRunner(runner gitRunner) *Service {
	return &Service{runner: runner}
}

// EnsureClean verifies the repository has no staged, unstaged, or untracked files.
func (s *Service) EnsureClean() error {
	out, err := s.runner.Run("status", "--porcelain")
	if err != nil {
		return err
	}

	if strings.TrimSpace(out) != "" {
		return fmt.Errorf("repository has uncommitted changes; commit, stash, or clean before tagging")
	}

	return nil
}

// LatestStableTag returns the highest stable tag in vX.Y.Z format.
func (s *Service) LatestStableTag() (string, error) {
	out, err := s.runner.Run("tag", "--list")
	if err != nil {
		return "", err
	}

	tags := strings.Split(out, "\n")
	bestTag := ""
	var bestVersion *semverx.Version

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		version, ok := stableVersionFromTag(tag)
		if !ok {
			continue
		}

		if bestVersion == nil || compare(version, bestVersion) > 0 {
			bestVersion = version
			bestTag = tag
		}
	}

	if bestVersion == nil {
		return "", fmt.Errorf("no stable tags found; create an initial tag like v0.0.1")
	}

	return bestTag, nil
}

// NextTag returns the next stable tag based on the latest stable existing tag.
func (s *Service) NextTag(component string) (string, error) {
	current, err := s.LatestStableTag()
	if err != nil {
		return "", err
	}

	version, ok := stableVersionFromTag(current)
	if !ok {
		return "", fmt.Errorf("latest tag %q is not a stable semantic version", current)
	}

	next, err := version.Bump(component)
	if err != nil {
		return "", err
	}

	return "v" + next.String(), nil
}

// NormalizeTag validates and normalizes input to vX.Y.Z.
func (s *Service) NormalizeTag(versionOrTag string) (string, error) {
	trimmed := strings.TrimSpace(versionOrTag)
	if trimmed == "" {
		return "", fmt.Errorf("version is required")
	}

	version := strings.TrimPrefix(trimmed, "v")
	if !isStableSemver(version) {
		return "", fmt.Errorf("invalid stable semantic version: %q (expected X.Y.Z or vX.Y.Z)", versionOrTag)
	}

	return "v" + version, nil
}

// CreateAnnotatedTag creates an annotated git tag.
func (s *Service) CreateAnnotatedTag(tag string) error {
	exists, err := s.TagExists(tag)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("tag already exists: %s", tag)
	}

	_, err = s.runner.Run("tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag))
	return err
}

// PushTag pushes a tag to origin.
func (s *Service) PushTag(tag string) error {
	_, err := s.runner.Run("push", "origin", tag)
	return err
}

// TagExists checks whether an exact tag exists.
func (s *Service) TagExists(tag string) (bool, error) {
	out, err := s.runner.Run("tag", "--list", tag)
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == tag {
			return true, nil
		}
	}

	return false, nil
}

func stableVersionFromTag(tag string) (*semverx.Version, bool) {
	if !strings.HasPrefix(tag, "v") {
		return nil, false
	}

	version := strings.TrimPrefix(tag, "v")
	if !isStableSemver(version) {
		return nil, false
	}

	parsed, err := semverx.Parse(version)
	if err != nil {
		return nil, false
	}

	return parsed, true
}

func isStableSemver(version string) bool {
	if version == "" {
		return false
	}

	parsed, err := semverx.Parse(version)
	if err != nil {
		return false
	}

	return parsed.Prerelease == "" && parsed.Metadata == ""
}

func compare(a, b *semverx.Version) int {
	if a.Major < b.Major {
		return -1
	}
	if a.Major > b.Major {
		return 1
	}

	if a.Minor < b.Minor {
		return -1
	}
	if a.Minor > b.Minor {
		return 1
	}

	if a.Patch < b.Patch {
		return -1
	}
	if a.Patch > b.Patch {
		return 1
	}

	return 0
}

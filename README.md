# semverctl

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/semverctl)](https://goreportcard.com/report/github.com/metalagman/semverctl)
[![lint](https://github.com/metalagman/semverctl/actions/workflows/lint.yml/badge.svg)](https://github.com/metalagman/semverctl/actions/workflows/lint.yml)
[![test](https://github.com/metalagman/semverctl/actions/workflows/test.yml/badge.svg)](https://github.com/metalagman/semverctl/actions/workflows/test.yml)
[![codecov](https://codecov.io/github/metalagman/semverctl/graph/badge.svg)](https://codecov.io/github/metalagman/semverctl)
[![version](https://img.shields.io/github/v/release/metalagman/semverctl?sort=semver)](https://github.com/metalagman/semverctl/releases)
[![npm](https://img.shields.io/npm/v/%40metalagman%2Fsemverctl)](https://www.npmjs.com/package/@metalagman/semverctl)
[![PyPI](https://img.shields.io/pypi/v/semverctl)](https://pypi.org/project/semverctl/)
[![license](https://img.shields.io/github/license/metalagman/semverctl)](LICENSE)

CLI for bumping and setting SemVer values in JSON/YAML files

## Features

- ✨ **Semantic Versioning** - Strict SemVer 2.0.0 compliance with prerelease and build metadata support
- 📁 **Multiple Formats** - JSON and YAML file support
- 🎯 **Path Navigation** - Dot-notation paths for nested version fields (e.g., `.app.version`)
- 🔢 **Numeric Bumping** - Bump individual numeric fields for object-style versions
- 🧪 **Dry-Run Mode** - Preview changes with unified diff output
- 🌐 **Cross-Platform** - Linux, macOS, and Windows support (CGO-free)
- ⚡ **Zero Dependencies** - Pure Go implementation with no runtime dependencies
- 📦 **Multiple Distribution Channels** - Go install, npm, uv/PyPI, and GitHub Releases

## Installation

### Go

```bash
go install github.com/metalagman/semverctl/cmd/semverctl@latest
```

### npm (Node.js)

```bash
npx @metalagman/semverctl version

# Or install globally
npm install -g @metalagman/semverctl
semverctl version
```

### uv/Pip (Python)

```bash
uvx semverctl version

# Or install
uv pip install semverctl
semverctl version
```

### Pre-built Binaries

Download pre-built binaries from [GitHub Releases](https://github.com/metalagman/semverctl/releases):

```bash
# Linux/macOS
curl -L https://github.com/metalagman/semverctl/releases/latest/download/semverctl-linux-amd64 -o semverctl
chmod +x semverctl
sudo mv semverctl /usr/local/bin/

# Verify checksum (recommended)
curl -L https://github.com/metalagman/semverctl/releases/latest/download/checksums.txt -o checksums.txt
sha256sum -c checksums.txt
```

## Usage

### Bump Version

Bump the semantic version at the specified path in JSON or YAML files:

```bash
# Bump patch version in package.json (default target)
semverctl bump

# Bump specific version component in package.json
semverctl bump --minor
semverctl bump --major
semverctl bump --patch

# Bump version in a specific file at a custom path
semverctl bump --file config.yaml --path .app.version

# Bump all matching files in current directory tree
semverctl bump --glob "**/*.json"

# Preview changes without modifying files
semverctl bump --dry-run
```

### Set Version

Set an explicit version value:

```bash
# Set version to 1.2.3 in package.json (default target)
semverctl set 1.2.3

# Set version in a specific file at a custom path
semverctl set 2.0.0 --file config.yaml --path .app.version

# Set version in all matching files under current directory
semverctl set 2.0.0 --glob "**/*.json"

# Preview changes
semverctl set 1.0.0 --dry-run
```

### Numeric Bump

For object-style version fields (e.g., `{ "Major": 1, "Minor": 2, "Patch": 3 }`),
you can bump numeric scalar values:

```bash
semverctl bump --numeric --path .version.Patch --file config.json
```

This increments the numeric value at the specified path by 1.

## Path Syntax

Paths use dot notation to navigate nested structures:

- `.version` - Top-level version field
- `.app.version` - Nested version field
- `.package.version` - Deeply nested field

The leading dot is optional: `version` and `.version` are equivalent.

## File Formats

Supported formats:

- **JSON** (`.json`)
- **YAML** (`.yaml`, `.yml`)

## Strict SemVer

semverctl follows the [Semantic Versioning 2.0.0](https://semver.org/) specification:

- Versions must be in format `MAJOR.MINOR.PATCH`
- Prerelease and build metadata are supported: `1.0.0-alpha+build.123`
- Leading zeros are not allowed in numeric components
- When bumping, prerelease and build metadata are cleared

## Dry-Run Mode

Use `--dry-run` to preview changes without modifying files:

```bash
semverctl bump --dry-run package.json
```

This outputs a unified diff showing what would change.

## Exit Codes

- `0` - Success
- `1` - Error (invalid arguments, file not found, parse error, etc.)

## License

MIT License - see [LICENSE](LICENSE) file for details.

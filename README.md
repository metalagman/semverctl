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
- 📁 **File + Tag Workflows** - Explicit `bump/set file` and `bump/set tag` commands
- 🏷️ **Git Tag Releases** - Bump from latest stable `vX.Y.Z` tag or set explicit tags
- 🎯 **Path Navigation** - Dot-notation paths for nested version fields (e.g., `.app.version`)
- 🔢 **Numeric Bumping** - Bump individual numeric fields for object-style versions
- 🧪 **Dry-Run Mode** - Preview changes with unified diff output
- 🤖 **Automation-Friendly JSON** - `--json` output for machine-readable success/error payloads
- 🔒 **Safety Checks** - Tag operations require a clean repository state
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

### Bump File Version

Bump the semantic version in a specific JSON or YAML file:

```bash
# Bump patch in a file
semverctl bump file package.json

# Bump specific version component
semverctl bump file package.json --minor
semverctl bump file package.json --major

# Bump version at a custom path
semverctl bump file config.yaml --path .app.version

# Preview changes without modifying the file
semverctl bump file package.json --dry-run

# Machine-readable output for automation
semverctl bump file package.json --dry-run --json
```

### Set File Version

Set an explicit version value in a specific file:

```bash
# Set version to 1.2.3 in a file
semverctl set file 1.2.3 package.json

# Set version at a custom path
semverctl set file 2.0.0 config.yaml --path .app.version

# Preview changes
semverctl set file 1.0.0 package.json --dry-run

# Machine-readable output for automation
semverctl set file 1.2.3 package.json --json
```

### Bump Tag

Bump from the latest stable git tag (`vX.Y.Z`) and create a new annotated tag:

```bash
# Create next patch tag
semverctl bump tag

# Create next minor tag
semverctl bump tag --minor

# Preview tag creation
semverctl bump tag --dry-run

# Create and push tag to origin
semverctl bump tag --push

# Machine-readable output for automation
semverctl bump tag --dry-run --json
```

### Set Tag

Create an explicit annotated git tag:

```bash
# Accepts 1.2.3 or v1.2.3
semverctl set tag 1.2.3
semverctl set tag v2.0.0

# Preview tag creation
semverctl set tag 2.1.0 --dry-run

# Create and push tag to origin
semverctl set tag 2.1.0 --push

# Machine-readable output for automation
semverctl set tag 2.1.0 --json
```

### Numeric Bump

For object-style version fields (e.g., `{ "Major": 1, "Minor": 2, "Patch": 3 }`),
you can bump numeric scalar values:

```bash
semverctl bump file config.json --numeric --path .version.Patch
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

Use `--dry-run` to preview file changes or tag actions without mutation:

```bash
semverctl bump file package.json --dry-run
semverctl bump tag --dry-run
```

For file commands, this outputs a unified diff showing what would change.

## JSON Output

Use `--json` for machine-readable automation output:

```bash
semverctl bump tag --dry-run --json
semverctl set tag 1.2.3 --json
semverctl bump file package.json --dry-run --json
```

When `--json` is enabled, both success and failure responses are printed as JSON to stdout.
Failures still return a non-zero exit code.

## Exit Codes

- `0` - Success
- `1` - Error (invalid arguments, file not found, parse error, etc.)

## License

MIT License - see [LICENSE](LICENSE) file for details.

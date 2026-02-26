# semverctl

[![Tests](https://github.com/metalagman/semverctl/workflows/Test/badge.svg)](https://github.com/metalagman/semverctl/actions/workflows/test.yml)
[![Lint](https://github.com/metalagman/semverctl/workflows/Lint/badge.svg)](https://github.com/metalagman/semverctl/actions/workflows/lint.yml)
[![Release](https://img.shields.io/github/v/release/metalagman/semverctl)](https://github.com/metalagman/semverctl/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

CLI for bumping and setting SemVer values in JSON/YAML files

## Features

- ✨ **Semantic Versioning** - Strict SemVer 2.0.0 compliance with prerelease and build metadata support
- 📁 **Multiple Formats** - JSON and YAML file support
- 🎯 **Path Navigation** - Dot-notation paths for nested version fields (e.g., `.app.version`)
- 🔢 **Numeric Bumping** - Bump individual numeric fields for object-style versions
- 🧪 **Dry-Run Mode** - Preview changes with unified diff output
- 🌐 **Cross-Platform** - Linux, macOS, and Windows support
- 📦 **Multiple Distribution Channels** - Go install, npm, uv/PyPI, and GitHub Releases

## Installation

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

### Go

```bash
go install github.com/metalagman/semverctl/cmd/semverctl@latest
```

```bash
go install github.com/metalagman/semverctl/cmd/semverctl@latest
```

Or download a pre-built binary from the [releases page](https://github.com/metalagman/semverctl/releases).

## Usage

### Bump Version

Bump the semantic version at the specified path in JSON or YAML files:

```bash
# Bump patch version (default)
semverctl bump package.json

# Bump specific version component
semverctl bump --minor package.json
semverctl bump --major package.json
semverctl bump --patch package.json

# Bump version at a custom path
semverctl bump --path .app.version config.yaml

# Bump all JSON files in a directory
semverctl bump --glob "**/*.json" .

# Preview changes without modifying files
semverctl bump --dry-run package.json
```

### Set Version

Set an explicit version value:

```bash
# Set version to 1.2.3
semverctl set 1.2.3 package.json

# Set version at a custom path
semverctl set 2.0.0 --path .app.version config.yaml

# Preview changes
semverctl set 1.0.0 --dry-run package.json
```

### Numeric Bump

For object-style version fields (e.g., `{ "Major": 1, "Minor": 2, "Patch": 3 }`),
you can bump numeric scalar values:

```bash
semverctl bump --numeric --path .version.Patch config.json
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

## Development

### Building

```bash
go build -o semverctl ./cmd/semverctl/main.go
```

### Testing

```bash
go test ./...
```

### Linting

```bash
golangci-lint run ./...
```

## Releasing

This project uses [Omnidist](https://github.com/omnidist/omnidist) for cross-platform distribution.

### Creating a Release

1. Tag the release:
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. The GitHub Actions workflow will automatically:
   - Build binaries for all platforms
   - Stage artifacts
   - Publish to npm as `@metalagman/semverctl`
   - Create a GitHub Release

### Installation from Package Managers

**npm:**
```bash
npx @metalagman/semverctl version
```

**uv (Python):**
```bash
uvx semverctl version
```

### Build from Source with Omnidist

```bash
# Build for all platforms
omnidist build

# Stage artifacts
omnidist stage

# Verify
omnidist verify
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Run linter (`golangci-lint run ./...`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Security

For security issues, please email the maintainer directly rather than opening a public issue.

## License

MIT License - see [LICENSE](LICENSE) file for details.

# semverctl

CLI for bumping and setting SemVer values in JSON/YAML files

## Installation

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

**Local build with Omnidist:**
```bash
# Build for all platforms
omnidist build

# Stage artifacts
omnidist stage

# Verify
omnidist verify
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

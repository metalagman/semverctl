# Contributing to semverctl

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Run linter (`golangci-lint run ./...`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

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

### Build from Source with Omnidist

```bash
# Build for all platforms
omnidist build

# Stage artifacts
omnidist stage

# Verify
omnidist verify
```

## Security

For security issues, please email the maintainer directly rather than opening a public issue.

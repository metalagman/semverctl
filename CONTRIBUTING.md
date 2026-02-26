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

## Security

For security issues, please email the maintainer directly rather than opening a public issue.

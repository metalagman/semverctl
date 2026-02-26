# Omnidist Publish Documentation

## NPM Publish Behavior

### Authentication Requirements

Even with `--dry-run` flag, npm requires authentication to:
- Verify package name availability
- Check user permissions
- Validate registry access

**Error without authentication:**
```
NPM authentication failed: npm error code E401
npm error 401 Unauthorized - GET https://registry.npmjs.org/-/whoami
```

### Token Requirements

To publish to npm, you need:
1. An npm account with appropriate permissions
2. An access token with `publish` scope
3. The token set as `NPM_TOKEN` environment variable in GitHub Actions

### Creating an NPM Token

1. Login to npmjs.com
2. Go to Access Tokens → Generate New Token
3. Select "Automation" token type
4. Add token to GitHub repository secrets as `NPM_TOKEN`

### Dry-Run Behavior

When authenticated, `omnidist npm publish --dry-run` will:
- Show what packages would be published
- Display version numbers
- List target platforms
- NOT actually publish to registry

### Platform Artifacts

The npm package includes platform-specific binaries:
- `@metalagman/semverctl-darwin-arm64`
- `@metalagman/semverctl-darwin-x64`
- `@metalagman/semverctl-linux-arm64`
- `@metalagman/semverctl-linux-x64`
- `@metalagman/semverctl-win32-x64`

The main package `@metalagman/semverctl` acts as a wrapper that downloads the appropriate binary for the current platform.

## UV Publish

Note: UV publishing is configured in `omnidist.yaml` but requires separate setup:
- PyPI account
- API token with upload permissions
- Token configured as `PYPI_TOKEN` in GitHub Actions

## Release Checklist

Before creating a release:
1. [ ] Ensure all tests pass
2. [ ] Update version in git tag
3. [ ] Verify `NPM_TOKEN` secret is set in GitHub
4. [ ] Create and push git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. [ ] Monitor GitHub Actions workflow
6. [ ] Verify package is published to npm
7. [ ] Test installation: `npx @metalagman/semverctl version`

# Contributing to Terraform Provider for Wormly

Thank you for your interest in contributing! This guide will help you get started with development and contribution workflows.

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.24
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [GoReleaser](https://goreleaser.com/install/) (for releases)

### Building the Provider

```shell
git clone https://github.com/radarnex/terraform-provider-wormly
cd terraform-provider-wormly
go mod download
make build
```

### Testing

Run unit tests:
```shell
make test
```

Run acceptance tests (requires `WORMLY_API_KEY`):
```shell
export WORMLY_API_KEY="your-api-key"
make testacc
```

### Local Installation

Install the provider locally for testing:
```shell
make install
```

This builds and installs the provider to your local Terraform plugins directory.

### Documentation Generation

Generate documentation from code:
```shell
make generate
```

### Code Quality

Run linting and formatting:
```shell
make lint
make fmt
```

## Release Process

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

### Creating a Release

1. **Update the version**: Update `CHANGELOG.md` with the new version and changes
2. **Commit changes**: Commit all changes to the main branch
3. **Create and push a tag**:
   ```shell
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. **Automated release**: GitHub Actions will automatically create a release using GoReleaser

### Testing a Release Locally

Test the release process locally:
```shell
make release-snapshot
```

This creates a snapshot build without publishing.

### Manual Release

If needed, you can manually create a release:
```shell
export GITHUB_TOKEN="your-github-token"
goreleaser release --clean
```

### Release Checklist

- [ ] All tests pass (`make test && make testacc`)
- [ ] Documentation is up to date (`make generate`)
- [ ] `CHANGELOG.md` is updated with the new version
- [ ] Version tag follows [semantic versioning](https://semver.org/)
- [ ] Release notes are clear and complete

## Contributing Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Getting Help

- **Documentation**: See the [`docs/`](./docs/) directory
- **Examples**: See the [`examples/`](./examples/) directory  
- **Issues**: [GitHub Issues](https://github.com/radarnex/terraform-provider-wormly/issues)

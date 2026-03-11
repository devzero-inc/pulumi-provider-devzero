# Contributing to pulumi-provider-devzero

Thank you for your interest in contributing to the DevZero Pulumi provider! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/devzero-inc/pulumi-provider-devzero.git
   cd pulumi-provider-devzero
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development

### Prerequisites

- Go 1.21+
- [Pulumi CLI](https://www.pulumi.com/docs/install/) (`brew install pulumi`)
- Python 3 (for schema generation scripts)
- `golangci-lint` (for linting)

### Building

```bash
make build
```

This compiles the provider binary to `bin/pulumi-resource-devzero`.

### Installing locally

```bash
make install
```

Copies the built binary to your `$GOPATH/bin`.

### Running Tests

```bash
make test
```

### Cleaning build artifacts

```bash
make clean
```

Removes `bin/` and `sdk/` directories.

## Schema & SDK Generation

> **Note:** Schema and SDK generation requires the Pulumi CLI to be installed.

### Generate schema

```bash
make gen-schema
```

Extracts the schema from the provider binary, merges it into `schema.json`, and applies enum patches.

### Generate SDKs (TypeScript, Python, Go)

```bash
make gen-sdk
```

Or simply:

```bash
make sdk
```

This runs `gen-schema` first, then generates SDKs for all supported languages under `sdk/`.

## Proto Sync (Core Maintainers Only)

> **This target is reserved for core maintainers.** Do not run `make proto` unless you are a core maintainer with access to the internal `services` repository.

Proto and generated files are synced from the internal `services` repo. If you need proto files updated, please open an issue or reach out to a core maintainer.

```bash
# Core maintainers only — requires ../services to be checked out
make proto

# Or specify a custom path:
make proto SERVICES_DIR=/path/to/services
```

## Dependency Management

```bash
make tidy
```

Runs `go mod tidy` to keep dependencies clean.

## Submitting Changes

1. Ensure your code builds and tests pass:
   ```bash
   make build
   make test
   ```
2. Commit your changes with a clear, descriptive commit message.
3. Push to your fork and open a Pull Request against `main`.
4. Describe your changes in the PR description and link any relevant issues.

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR.
- Add tests for new functionality.
- Update documentation if your changes affect user-facing behavior.
- Ensure CI passes before requesting review.
- Do not modify proto or generated files — those are managed by core maintainers.

## Reporting Issues

- Use [GitHub Issues](https://github.com/devzero-inc/pulumi-provider-devzero/issues) to report bugs or request features.
- Include steps to reproduce, expected behavior, and actual behavior for bug reports.

## Code Style

- Follow standard Go conventions and idioms.
- Use `gofmt` / `goimports` for formatting.
- Wrap errors with context using `fmt.Errorf("context: %w", err)`.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
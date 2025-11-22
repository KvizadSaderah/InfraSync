# Contributing to InfraSync

Thank you for your interest in contributing to InfraSync! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional. We're all here to build better tools together.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Basic understanding of Terraform

### Development Setup

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/infrasync.git
cd infrasync
```

3. Install dependencies:
```bash
go mod download
```

4. Run tests to verify setup:
```bash
go test ./...
```

## Development Workflow

### Making Changes

1. Create a branch for your changes:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes, following our coding standards

3. Add tests for new functionality

4. Run tests and ensure they pass:
```bash
go test -v ./...
```

5. Run linting:
```bash
go fmt ./...
go vet ./...
```

6. Commit your changes:
```bash
git add .
git commit -m "feat: add your feature description"
```

### Commit Message Format

We follow conventional commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

Examples:
```
feat: add support for Azure resources
fix: correct replace action detection
docs: update installation instructions
test: add tests for analyzer package
```

### Pull Request Process

1. Push your branch to your fork:
```bash
git push origin feature/your-feature-name
```

2. Open a Pull Request against `main` branch

3. Fill in the PR template with:
   - Description of changes
   - Related issue numbers
   - Testing performed
   - Screenshots (if UI changes)

4. Wait for review and address feedback

5. Once approved, a maintainer will merge your PR

## Project Structure

```
infrasync/
├── cmd/
│   └── infrasync/          # Main CLI application
├── pkg/
│   ├── parser/             # Terraform plan parsing
│   ├── formatter/          # Output formatting (CLI, Markdown)
│   └── analyzer/           # Security and risk analysis
├── action/                 # GitHub Action
├── examples/               # Example Terraform configurations
├── docs/                   # Documentation
└── .github/workflows/      # CI/CD workflows
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/parser/...
```

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests when appropriate
- Aim for good coverage of edge cases
- Include both positive and negative test cases

Example:
```go
func TestSomeFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SomeFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Areas for Contribution

### High Priority

- [ ] Support for more cloud providers (Azure, GCP)
- [ ] Additional security analysis rules
- [ ] Performance improvements for large plans
- [ ] Better error messages and validation
- [ ] More comprehensive examples

### Medium Priority

- [ ] HTML output format
- [ ] JSON output format
- [ ] Custom rule configuration
- [ ] Plan comparison (diff between two plans)
- [ ] Integration with other CI systems (GitLab, Bitbucket)

### Documentation

- [ ] More real-world examples
- [ ] Video tutorials
- [ ] Blog posts about usage
- [ ] Translations

### Good First Issues

Look for issues tagged with `good first issue` for beginner-friendly tasks.

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused
- Add comments for exported functions

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to parse plan: %w", err)
}

// Bad
if err != nil {
    panic(err)
}
```

### Naming Conventions

- Use descriptive names
- Avoid abbreviations unless well-known
- Use camelCase for unexported, PascalCase for exported

## Documentation

### Code Comments

- Add comments for all exported functions, types, and constants
- Explain the "why" not just the "what"
- Keep comments up-to-date with code changes

### Documentation Files

- Update relevant docs when adding features
- Include examples in documentation
- Keep README.md concise, detailed docs in docs/

## Release Process

Maintainers will handle releases:

1. Version bump in code
2. Update CHANGELOG.md
3. Create and push tag
4. GitHub Actions builds and publishes release

## Questions?

- Open a [Discussion](https://github.com/kvizadsaderah/infrasync/discussions)
- Open an [Issue](https://github.com/kvizadsaderah/infrasync/issues)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

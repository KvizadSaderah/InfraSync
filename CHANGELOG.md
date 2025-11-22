# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Custom rule configuration via YAML
- GitLab CI support
- HTML report generation
- Plan comparison (diff between two plans)
- Cost estimation integration
- Slack/Teams notifications

## [0.2.0] - 2025-11-22

### Added
- **Complete rewrite** with modular architecture
- Security and risk analysis engine
  - Critical warnings for database deletions, production resources
  - High-risk warnings for storage deletions, encryption changes
  - Backup and versioning change detection
- Markdown output format for GitHub PR comments
- GitHub Action for automatic PR commenting
- Comprehensive test suite
- CLI flags: `--format`, `--verbose`, `--compact`, `--warnings`, `--output`
- Exit codes for CI/CD integration (0=no changes, 1=changes, 2=critical warnings)
- Beautiful terminal output with colors and formatting
- Support for replace operations (previously buggy)
- Proper handling of sensitive values and unknown values

### Changed
- Migrated from monolithic `main.go` to clean package structure:
  - `pkg/parser` - Terraform plan parsing
  - `pkg/formatter` - Output formatting (CLI, Markdown)
  - `pkg/analyzer` - Security and risk analysis
- Fixed module name from `infrasearch` to `github.com/kvizadsaderah/infrasync`
- Completely rewrote README with focus on real-world value
- Improved error handling and user messages

### Fixed
- Replace operations (delete+create) now correctly identified
- No longer shows only delete for replace actions
- Better handling of nested attributes
- Proper map/slice value formatting

### Infrastructure
- GitHub Actions CI/CD pipeline
- Release workflow for multi-platform binaries
- golangci-lint configuration
- Makefile for development tasks
- Comprehensive documentation (USAGE.md, CONTRIBUTING.md)

## [0.1.0] - 2025-11-22 (Initial)

### Added
- Basic CLI tool to parse Terraform plan JSON
- Simple colored output for creates, updates, deletes
- Basic attribute diff printing
- Example Terraform plan file

### Known Issues
- Replace operations incorrectly shown as delete only
- Module name inconsistency
- No tests
- No documentation
- Monolithic code structure

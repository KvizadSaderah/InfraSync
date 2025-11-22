# InfraSync

**Beautiful Terraform Plan Analysis for Pull Requests**

[![CI](https://github.com/kvizadsaderah/infrasync/workflows/CI/badge.svg)](https://github.com/kvizadsaderah/infrasync/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kvizadsaderah/infrasync)](go.mod)

InfraSync transforms hard-to-read Terraform plans into beautiful, easy-to-review summaries. Perfect for code reviews, CI/CD pipelines, and catching dangerous infrastructure changes before they reach production.

## 🎯 The Problem

Reviewing Terraform changes in pull requests is painful:
- 😵 Long, unreadable `terraform plan` outputs buried in CI logs
- 🔍 Easy to miss critical changes like database deletions
- ⏱️ Time-consuming to understand what's actually changing
- 🚨 No automatic warnings for dangerous operations

## ✨ The Solution

InfraSync makes infrastructure changes **visible and safe**:

```diff
## 🔄 Terraform Plan Summary

### 📊 Changes Overview

| Action | Count |
|--------|-------|
| ✅ **Create** | 2 |
| 🔄 **Update** | 1 |
| ⚠️ **Replace** | 1 |

### ⚠️ Warning: Destructive Changes Detected

🚨 **CRITICAL WARNING**: Database will be REPLACED - downtime expected!
- Resource: `aws_db_instance.production`
- This causes service interruption and potential data loss
```

**Automatic PR comments** • **Security warnings** • **Beautiful formatting** • **Zero config**

## 🚀 Quick Start

### GitHub Action (Recommended)

Add to your workflow in `.github/workflows/terraform.yml`:

```yaml
name: Terraform Plan

on:
  pull_request:
    paths: ['**.tf']

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3

      - name: Terraform Init
        run: terraform init

      - name: Terraform Plan
        run: |
          terraform plan -out=tfplan
          terraform show -json tfplan > tfplan.json

      - name: Analyze with InfraSync
        uses: kvizadsaderah/infrasync/action@v0.2.0
        with:
          plan-file: tfplan.json
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

**That's it!** InfraSync will automatically comment on your PRs with beautiful plan summaries.

### CLI Usage

Install:
```bash
# macOS
brew install kvizadsaderah/tap/infrasync

# Linux/macOS (manual)
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-linux-amd64
chmod +x infrasync-linux-amd64
sudo mv infrasync-linux-amd64 /usr/local/bin/infrasync

# Or build from source
go install github.com/kvizadsaderah/infrasync/cmd/infrasync@latest
```

Use:
```bash
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
infrasync tfplan.json
```

## 📸 Screenshots

### Terminal Output
```
═══════════════════════════════════════════════════════
  Terraform Plan Summary
═══════════════════════════════════════════════════════

Changes Overview:
─────────────────
  ✓ 3 to create
  ~ 2 to update
  ⟳ 1 to replace
  ✗ 1 to destroy

✓ Resources to CREATE (3):
────────────────────────────
  + aws_instance.web_server[0]
    Type: aws_instance
  + aws_instance.web_server[1]
    Type: aws_instance
  + aws_s3_bucket.logs
    Type: aws_s3_bucket

🚨 CRITICAL WARNINGS (1):
─────────────────────────
  • Database will be DESTROYED - data loss risk!
    Resource: aws_db_instance.production
    Ensure backups are in place before deleting databases
```

### GitHub PR Comment
See [example output](docs/example-pr-comment.md)

## 🎁 Key Features

### 1. **Smart Security Analysis**

Automatically detects dangerous operations:
- 🚨 **Critical**: Database deletions, encryption disabled, production resource destruction
- ⚠️ **High Risk**: Storage deletions, security group weakening, backup disabled
- ℹ️ **Medium**: Other potentially risky changes

### 2. **Beautiful Output**

- **Terminal**: Colored, organized, easy to scan
- **Markdown**: Perfect for GitHub/GitLab PR comments
- **Collapsible sections**: Keep PRs clean

### 3. **Zero Configuration**

Works out of the box with any Terraform project. No config files needed.

### 4. **CI/CD Ready**

- GitHub Actions support
- Exit codes for pipeline control
- Artifact-friendly output

### 5. **Developer Friendly**

```bash
# Verbose mode for debugging
infrasync --verbose tfplan.json

# Save to file
infrasync --format markdown --output plan.md tfplan.json

# Compact summary
infrasync --compact tfplan.json
```

## 📚 Documentation

- **[Usage Guide](docs/USAGE.md)** - Detailed CLI and GitHub Action usage
- **[Contributing](CONTRIBUTING.md)** - How to contribute
- **[Examples](examples/)** - Real-world examples

## 🔧 Advanced Usage

### Multiple Environments

```yaml
strategy:
  matrix:
    environment: [prod, staging, dev]
steps:
  - name: Plan ${{ matrix.environment }}
    working-directory: terraform/${{ matrix.environment }}
    run: |
      terraform init
      terraform plan -out=tfplan
      terraform show -json tfplan > tfplan.json

  - uses: kvizadsaderah/infrasync/action@v0.2.0
    with:
      plan-file: terraform/${{ matrix.environment }}/tfplan.json
```

### Custom Security Rules

Coming soon! Vote on [this issue](https://github.com/kvizadsaderah/infrasync/issues/1).

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

**Areas where we need help:**
- Azure and GCP resource type detection
- More security analysis rules
- Performance improvements
- Documentation and examples

## 🛣️ Roadmap

- [x] CLI tool with colored output
- [x] Markdown generation for PRs
- [x] GitHub Action
- [x] Security analysis (databases, encryption, etc.)
- [ ] Custom rule configuration (YAML)
- [ ] GitLab CI support
- [ ] HTML report generation
- [ ] Plan comparison (diff between two plans)
- [ ] Slack/Teams notifications
- [ ] Cost estimation integration

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/kvizadsaderah/infrasync/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kvizadsaderah/infrasync/discussions)

## ⭐ Show Your Support

If InfraSync helps you catch bugs or speeds up your PR reviews, give it a ⭐️!

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Built with:
- [terraform-json](https://github.com/hashicorp/terraform-json) - Terraform plan parsing
- [color](https://github.com/fatih/color) - Terminal colors

---

**Made with ❤️ for the DevOps community**

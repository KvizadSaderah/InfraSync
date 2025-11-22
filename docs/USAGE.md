# InfraSync Usage Guide

## Installation

### Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/kvizadsaderah/infrasync/releases).

```bash
# Linux (amd64)
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-linux-amd64
chmod +x infrasync-linux-amd64
sudo mv infrasync-linux-amd64 /usr/local/bin/infrasync

# macOS (Apple Silicon)
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-darwin-arm64
chmod +x infrasync-darwin-arm64
sudo mv infrasync-darwin-arm64 /usr/local/bin/infrasync

# Verify installation
infrasync --version
```

### Build from Source

```bash
git clone https://github.com/kvizadsaderah/infrasync.git
cd infrasync
go build -o infrasync ./cmd/infrasync
sudo mv infrasync /usr/local/bin/
```

## CLI Usage

### Basic Plan Analysis

1. Generate a Terraform plan:
```bash
terraform plan -out=tfplan
```

2. Convert to JSON:
```bash
terraform show -json tfplan > tfplan.json
```

3. Analyze with InfraSync:
```bash
infrasync tfplan.json
```

### Output Formats

#### Terminal Output (Default)
```bash
infrasync tfplan.json
```

Outputs a beautiful colored terminal display with:
- Change statistics
- Resources to create (green ✓)
- Resources to update (yellow ~)
- Resources to replace (magenta ⟳)
- Resources to destroy (red ✗)
- Security warnings

#### Markdown Output
```bash
infrasync --format markdown tfplan.json
```

Generates GitHub-flavored markdown suitable for PR comments.

Save to file:
```bash
infrasync --format markdown --output plan.md tfplan.json
```

### Options

```bash
# Show all options
infrasync --help

# Verbose output with all attribute changes
infrasync --verbose tfplan.json

# Compact output
infrasync --compact tfplan.json

# Show unchanged resources
infrasync --show-unchanged tfplan.json

# Disable security warnings
infrasync --warnings=false tfplan.json

# Output to file
infrasync --output report.md --format markdown tfplan.json
```

### Exit Codes

- `0`: No changes detected
- `1`: Changes detected (normal)
- `2`: Changes detected with critical security warnings

Use in CI/CD:
```bash
infrasync tfplan.json
EXIT_CODE=$?

if [ $EXIT_CODE -eq 2 ]; then
  echo "Critical warnings detected! Manual review required."
  exit 1
fi
```

## GitHub Action Usage

### Quick Start

Add to your workflow (`.github/workflows/terraform.yml`):

```yaml
name: Terraform Plan

on:
  pull_request:
    paths:
      - '**.tf'

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3

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

### Advanced Configuration

```yaml
- name: Analyze with InfraSync
  uses: kvizadsaderah/infrasync/action@v0.2.0
  with:
    plan-file: tfplan.json
    github-token: ${{ secrets.GITHUB_TOKEN }}
    working-directory: ./terraform
    show-details: true
    post-comment: true
    update-comment: true  # Update existing comment instead of creating new ones
```

### Multiple Terraform Directories

```yaml
jobs:
  plan-infra:
    strategy:
      matrix:
        directory: [prod, staging, dev]
    steps:
      - uses: actions/checkout@v4

      - name: Terraform Plan
        working-directory: terraform/${{ matrix.directory }}
        run: |
          terraform init
          terraform plan -out=tfplan
          terraform show -json tfplan > tfplan.json

      - name: Analyze with InfraSync
        uses: kvizadsaderah/infrasync/action@v0.2.0
        with:
          plan-file: terraform/${{ matrix.directory }}/tfplan.json
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Real-World Examples

### Example 1: Local Development

```bash
cd my-terraform-project
terraform init
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
infrasync --verbose tfplan.json
```

### Example 2: CI/CD Pipeline

```bash
#!/bin/bash
set -e

# Run terraform plan
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json

# Analyze with InfraSync
infrasync --format markdown --output plan-summary.md tfplan.json

# Upload to artifacts or post to PR
echo "Plan summary:"
cat plan-summary.md
```

### Example 3: Code Review Helper

In your PR review workflow:

```yaml
- name: Comment Plan Summary
  uses: kvizadsaderah/infrasync/action@v0.2.0
  with:
    plan-file: tfplan.json
    github-token: ${{ secrets.GITHUB_TOKEN }}
    show-details: true
```

This will automatically post a comment like:

```markdown
## 🔄 Terraform Plan Summary

### 📊 Changes Overview

| Action | Count |
|--------|-------|
| ✅ **Create** | 3 |
| 🔄 **Update** | 1 |
| **Total** | **4** |

<details>
<summary>✅ <b>Resources to CREATE (3)</b></summary>

```diff
+ aws_instance.web_server[0]
  Type: aws_instance
+ aws_instance.web_server[1]
  Type: aws_instance
+ aws_s3_bucket.logs
  Type: aws_s3_bucket
```
</details>

---
*Generated by InfraSync*
```

## Security Analysis

InfraSync automatically detects potentially dangerous operations:

### Critical Warnings
- Database deletions or replacements
- Production resource destruction
- Encryption being disabled
- Backup/versioning being disabled

### High Risk Warnings
- Storage resource deletions
- Network resource destruction
- Security group rule changes that weaken security
- Load balancer replacements

### Example Output

```
🚨 CRITICAL WARNINGS (1):
─────────────────────────
  • Database will be DESTROYED - data loss risk!
    Resource: aws_db_instance.production
    Ensure backups are in place before deleting databases

⚠️  HIGH RISK WARNINGS (1):
──────────────────────────
  • Security group rules are being relaxed
    Resource: aws_security_group.web
    Review that new permissions don't expose services unnecessarily
```

## Tips & Best Practices

1. **Always review the plan** before applying, especially when there are critical warnings

2. **Use in CI/CD** to get automatic PR comments with plan summaries

3. **Combine with policy tools** like OPA or Sentinel for additional validation

4. **Keep plans small** - easier to review and less risky

5. **Enable verbose mode** when debugging issues: `infrasync --verbose tfplan.json`

## Troubleshooting

### Issue: "Error reading plan file"
- Ensure the file is a valid Terraform plan JSON
- Generate with: `terraform show -json tfplan > tfplan.json`

### Issue: "No changes detected" but changes exist
- Verify the plan file isn't empty
- Check that `terraform plan` shows changes

### Issue: GitHub Action not posting comments
- Ensure the workflow has `pull-requests: write` permission
- Verify `github-token` is set correctly
- Check that the event is a `pull_request`

## Support

- **Issues**: https://github.com/kvizadsaderah/infrasync/issues
- **Discussions**: https://github.com/kvizadsaderah/infrasync/discussions

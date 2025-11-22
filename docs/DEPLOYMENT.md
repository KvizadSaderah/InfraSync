# Deployment Guide

## Local Installation

### Option 1: Download Pre-built Binary (Recommended)

**Linux (amd64)**
```bash
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-linux-amd64
chmod +x infrasync-linux-amd64
sudo mv infrasync-linux-amd64 /usr/local/bin/infrasync
```

**macOS (Apple Silicon)**
```bash
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-darwin-arm64
chmod +x infrasync-darwin-arm64
sudo mv infrasync-darwin-arm64 /usr/local/bin/infrasync
```

**macOS (Intel)**
```bash
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-darwin-amd64
chmod +x infrasync-darwin-amd64
sudo mv infrasync-darwin-amd64 /usr/local/bin/infrasync
```

**Windows**
```powershell
# Download from releases page
# Or use PowerShell:
Invoke-WebRequest -Uri "https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-windows-amd64.exe" -OutFile "infrasync.exe"
# Add to PATH manually
```

**Verify Installation**
```bash
infrasync --version
```

### Option 2: Install via Go

```bash
go install github.com/kvizadsaderah/infrasync/cmd/infrasync@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/kvizadsaderah/infrasync.git
cd infrasync
make build
sudo mv infrasync /usr/local/bin/
```

## GitHub Actions Deployment

### Basic Setup

Create `.github/workflows/terraform-pr.yml`:

```yaml
name: Terraform Plan Analysis

on:
  pull_request:
    paths:
      - '**.tf'
      - '**.tfvars'

permissions:
  contents: read
  pull-requests: write

jobs:
  terraform-plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.6.0

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

### Multi-Environment Setup

```yaml
name: Terraform Plan (All Environments)

on:
  pull_request:
    paths: ['terraform/**/*.tf']

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    strategy:
      matrix:
        environment: [dev, staging, prod]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3

      - name: Plan ${{ matrix.environment }}
        working-directory: terraform/${{ matrix.environment }}
        run: |
          terraform init
          terraform plan -out=tfplan
          terraform show -json tfplan > tfplan.json

      - name: Analyze ${{ matrix.environment }}
        uses: kvizadsaderah/infrasync/action@v0.2.0
        with:
          plan-file: terraform/${{ matrix.environment }}/tfplan.json
          github-token: ${{ secrets.GITHUB_TOKEN }}
          working-directory: terraform/${{ matrix.environment }}
```

### With Terraform Cloud

```yaml
- name: Download Plan from Terraform Cloud
  run: |
    terraform show -json > tfplan.json
  env:
    TF_TOKEN_app_terraform_io: ${{ secrets.TF_API_TOKEN }}

- name: Analyze Plan
  uses: kvizadsaderah/infrasync/action@v0.2.0
  with:
    plan-file: tfplan.json
    github-token: ${{ secrets.GITHUB_TOKEN }}
```

### Advanced: Fail on Critical Warnings

```yaml
- name: Analyze Plan
  id: infrasync
  uses: kvizadsaderah/infrasync/action@v0.2.0
  with:
    plan-file: tfplan.json
    github-token: ${{ secrets.GITHUB_TOKEN }}
  continue-on-error: true

- name: Check for Critical Warnings
  run: |
    if [ "${{ steps.infrasync.outcome }}" == "failure" ]; then
      echo "Critical warnings detected! Manual review required."
      exit 1
    fi
```

## GitLab CI/CD (Future)

**Note**: GitLab support is planned. For now, use CLI in script:

```yaml
terraform-plan:
  stage: plan
  script:
    - terraform init
    - terraform plan -out=tfplan
    - terraform show -json tfplan > tfplan.json
    - wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-linux-amd64
    - chmod +x infrasync-linux-amd64
    - ./infrasync-linux-amd64 --format markdown tfplan.json > plan.md
  artifacts:
    paths:
      - plan.md
```

## Docker Deployment

**Dockerfile** (if you want containerized execution):

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o infrasync ./cmd/infrasync

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/infrasync /usr/local/bin/
ENTRYPOINT ["infrasync"]
```

**Build and run**:
```bash
docker build -t infrasync:latest .
docker run -v $(pwd):/data infrasync:latest /data/tfplan.json
```

## CI/CD Best Practices

### 1. Pin Action Version
❌ **Bad**: `uses: kvizadsaderah/infrasync/action@main`
✅ **Good**: `uses: kvizadsaderah/infrasync/action@v0.2.0`

### 2. Use Terraform Lock Files
```yaml
- name: Cache Terraform
  uses: actions/cache@v3
  with:
    path: |
      .terraform
      .terraform.lock.hcl
    key: ${{ runner.os }}-terraform-${{ hashFiles('**/.terraform.lock.hcl') }}
```

### 3. Separate Plan and Apply
```yaml
jobs:
  plan:
    # Run InfraSync here

  apply:
    needs: plan
    if: github.ref == 'refs/heads/main'
    # Only apply on main branch
```

### 4. Require Approvals for Destructive Changes
Use GitHub's branch protection + CODEOWNERS:

```
# CODEOWNERS
terraform/ @infrastructure-team
```

## Enterprise Deployment

### Self-Hosted Runners

For sensitive infrastructure, use self-hosted runners:

```yaml
jobs:
  plan:
    runs-on: self-hosted
    # Rest of workflow
```

### Private Network Access

If Terraform needs private network access:

```yaml
- name: Connect to VPN
  run: |
    # Setup VPN connection

- name: Terraform Plan
  run: terraform plan -out=tfplan
  env:
    AWS_REGION: us-east-1
    # Use instance profile or temporary credentials
```

### Secrets Management

**Never commit credentials**. Use:
- GitHub Secrets
- AWS IAM roles (via OIDC)
- HashiCorp Vault
- Azure Key Vault

```yaml
- name: Configure AWS Credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::123456789012:role/github-actions
    aws-region: us-east-1
```

## Troubleshooting

### Issue: Action fails with "Permission denied"
**Solution**: Ensure `pull-requests: write` permission:
```yaml
permissions:
  pull-requests: write
```

### Issue: Binary not found
**Solution**: Use full path or verify installation:
```bash
which infrasync
ls -l /usr/local/bin/infrasync
```

### Issue: JSON parsing error
**Solution**: Verify Terraform JSON format:
```bash
terraform show -json tfplan | jq . > tfplan.json
```

### Issue: No PR comment posted
**Solution**: Check:
1. Workflow runs on `pull_request` event
2. `github-token` is set
3. Bot has write access to repo

## Monitoring and Observability

### Track Usage in GitHub Actions

Add to workflow:
```yaml
- name: Track InfraSync Usage
  run: |
    echo "::notice::InfraSync analysis complete"
    echo "CHANGES=${{ steps.infrasync.outputs.summary }}" >> $GITHUB_STEP_SUMMARY
```

### Custom Metrics (Future)

Send metrics to monitoring system:
```bash
infrasync tfplan.json --format json | \
  jq '.to_create + .to_update + .to_delete' | \
  curl -X POST https://metrics.example.com/api/infrasync \
    -d @-
```

## Rollback Plan

If issues occur:
1. **Revert to previous workflow** - Remove InfraSync action temporarily
2. **Use specific version** - Pin to last known good version
3. **Fallback to manual review** - Comment out action, review manually

## Updates and Maintenance

### Update InfraSync Version
```yaml
- uses: kvizadsaderah/infrasync/action@v0.3.0  # Update version here
```

### Check for Updates
```bash
# See latest release
gh release view --repo kvizadsaderah/infrasync

# Upgrade local binary
wget https://github.com/kvizadsaderah/infrasync/releases/latest/download/infrasync-linux-amd64
sudo mv infrasync-linux-amd64 /usr/local/bin/infrasync
```

### Auto-update via Dependabot

Create `.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

---

**Need help?** Open an issue: https://github.com/kvizadsaderah/infrasync/issues

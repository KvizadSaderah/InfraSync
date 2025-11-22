# Architecture Overview

## High-Level Design

InfraSync follows a clean, modular architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────┐
│                    CLI / GitHub Action                   │
│                  (cmd/infrasync/main.go)                │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
┌──────────┐  ┌─────────────┐  ┌──────────┐
│  Parser  │  │  Analyzer   │  │Formatter │
│          │  │             │  │          │
│ - Parse  │  │ - Detect    │  │ - CLI    │
│   plan   │  │   risks     │  │ - Markdown│
│ - Extract│  │ - Security  │  │ - Compact│
│   changes│  │   warnings  │  │          │
└──────────┘  └─────────────┘  └──────────┘
```

## Package Structure

### `pkg/parser`
**Responsibility**: Parse Terraform plan JSON files and extract resource changes.

**Key Types**:
- `PlanSummary` - High-level summary with counts
- `ResourceChange` - Individual resource change with metadata
- `ParsePlanFile()` - Entry point for parsing

**Why**: Separates the complex Terraform JSON parsing logic from business logic.

### `pkg/analyzer`
**Responsibility**: Analyze changes for security risks and operational dangers.

**Key Types**:
- `Warning` - Risk warning with severity level
- `RiskLevel` - Critical, High, Medium, Low
- `AnalyzeChanges()` - Entry point for analysis

**Detection Rules**:
- Database operations (delete/replace)
- Production resource changes
- Encryption/backup disabling
- Security group weakening
- Storage deletions

**Why**: Centralizes all risk detection logic in one place, making it easy to add new rules.

### `pkg/formatter`
**Responsibility**: Format plan summaries for different output targets.

**Formatters**:
- `CLIFormatter` - Colored terminal output
- `MarkdownFormatter` - GitHub-flavored markdown for PRs

**Why**: Separates presentation from logic, allowing easy addition of new output formats (HTML, JSON, etc.).

## Data Flow

```
1. Terraform Plan JSON File
   ↓
2. Parser (pkg/parser)
   - Reads JSON
   - Identifies action types (create/update/delete/replace)
   - Extracts before/after values
   ↓
3. PlanSummary struct
   - List of ResourceChange objects
   - Aggregate counts
   ↓
4. Analyzer (pkg/analyzer)
   - Examines each change
   - Checks against risk rules
   - Generates warnings
   ↓
5. Warnings list
   ↓
6. Formatter (pkg/formatter)
   - Takes PlanSummary + Warnings
   - Renders in chosen format
   ↓
7. Output (Terminal or File)
```

## Design Decisions

### 1. Why separate Parser from Analyzer?
**Decision**: Parse first, analyze second.

**Rationale**:
- Parsing is complex and error-prone
- Analysis rules change frequently
- Separation allows testing each independently
- Can add multiple analyzers (security, cost, compliance)

### 2. Why not use existing Terraform libraries directly in formatters?
**Decision**: Create intermediate `PlanSummary` type.

**Rationale**:
- terraform-json types are verbose and complex
- Our intermediate type is simpler to work with
- Easier to mock for testing
- Decouples from terraform-json version changes

### 3. Why colored output in CLI?
**Decision**: Use `fatih/color` for terminal colors.

**Rationale**:
- Humans parse visual information faster
- Color-coding (red=danger, green=safe) reduces cognitive load
- Standard practice in modern CLI tools (kubectl, helm, etc.)

### 4. Why exit codes?
**Decision**:
- 0 = no changes
- 1 = changes detected
- 2 = critical warnings

**Rationale**:
- Allows CI/CD pipelines to react to different scenarios
- 2 = fail fast on dangerous operations
- 1 = informational, continue
- Standard Unix convention

## Extension Points

### Adding New Output Format
1. Create new formatter in `pkg/formatter/`
2. Implement `Format(summary *parser.PlanSummary) string`
3. Add case in `cmd/infrasync/main.go`

Example:
```go
// pkg/formatter/json.go
type JSONFormatter struct{}

func (f *JSONFormatter) Format(summary *parser.PlanSummary) string {
    // Convert to JSON
}
```

### Adding New Risk Rule
1. Add detection function in `pkg/analyzer/analyzer.go`
2. Call from `AnalyzeChanges()` or relevant `analyze*()` function
3. Add tests

Example:
```go
func hasPublicAccessEnabled(before, after map[string]interface{}) bool {
    // Check if public_access changed from false to true
    return getBool(before, "public_access") == false &&
           getBool(after, "public_access") == true
}
```

### Adding New Cloud Provider
1. Add resource type detection functions in `pkg/analyzer/analyzer.go`
2. Update relevant `is*()` functions

Example:
```go
func isDatabase(resourceType string) bool {
    dbTypes := []string{
        "aws_db_instance",
        "google_sql_database",
        "azurerm_sql_database",
        // Add new provider
        "alicloud_db_instance",
    }
    return containsAny(resourceType, dbTypes)
}
```

## Testing Strategy

### Unit Tests
- `pkg/parser` - Test all change classifications
- `pkg/analyzer` - Test each risk detection rule
- `pkg/formatter` - Mock outputs (future)

### Integration Tests (Future)
- End-to-end with real Terraform plans
- Test CLI flag combinations
- Test GitHub Action behavior

### Test Data
- `examples/simple/` - Basic plan for quick tests
- `examples/aws-production/` - Realistic AWS infrastructure

## Performance Considerations

### Current
- **Small plans** (<100 resources): <100ms
- **Medium plans** (100-1000 resources): <1s
- **Large plans** (>1000 resources): <5s

### Bottlenecks
- JSON parsing (native Go, very fast)
- String matching in analyzer (acceptable for typical plans)

### Future Optimizations
- Parallel analysis of resources
- Compiled regex for patterns
- Caching of analysis results

## Security Considerations

### Input Validation
- Terraform JSON is trusted input (generated by Terraform)
- No user input parsing
- No code execution

### Secrets Handling
- Respects Terraform's `(sensitive)` markers
- Never logs sensitive values
- Doesn't persist plan data

### Dependencies
- Minimal external dependencies
- All from trusted sources (HashiCorp, stdlib)
- Regularly updated via Dependabot (future)

## GitHub Action Architecture

```
┌─────────────────────────────────────────┐
│  GitHub Actions Workflow                │
│  (.github/workflows/terraform-pr.yml)   │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│  InfraSync Action                       │
│  (action/action.yml)                    │
│                                         │
│  1. Setup Go                            │
│  2. Build infrasync binary              │
│  3. Run: infrasync --format markdown    │
│  4. Post comment via github-script      │
└─────────────────────────────────────────┘
```

### Why composite action?
- No Docker build time
- Faster execution
- Easier to debug
- Works in any runner

## Future Architecture Improvements

### 1. Plugin System
Allow users to write custom analyzers:
```go
type Plugin interface {
    Analyze(change ResourceChange) []Warning
}
```

### 2. Configuration File
YAML-based rule configuration:
```yaml
rules:
  database-deletion:
    enabled: true
    severity: critical
  public-access:
    enabled: true
    severity: high
```

### 3. Caching Layer
Cache analysis results between runs:
- Faster re-analysis of unchanged resources
- Useful for multi-environment workflows

### 4. Web UI Backend
REST API for plan storage and comparison:
```
POST /api/plans - Upload plan
GET  /api/plans/:id - View plan
GET  /api/plans/:id/compare/:id2 - Diff two plans
```

## Contributing to Architecture

When proposing changes:
1. Maintain separation of concerns
2. Add tests for new functionality
3. Update this document
4. Consider backward compatibility
5. Profile performance impact for large plans

---

**Last Updated**: 2025-11-22
**Version**: 0.2.0

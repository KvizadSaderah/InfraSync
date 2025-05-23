# InfraSync

**InfraSync** visualizes infrastructure changes by comparing the current state ("as-is") with the Terraform plan ("to-be") in a clear, delta-based format. This eliminates the need to manually review raw `terraform plan` outputs or JSON files, reducing the risk of unapplied changes, speeding up code reviews, and increasing confidence in the IaC process.

## Concept and Value

The core idea is to provide an intuitive and efficient way to understand the impact of Terraform changes before they are applied.

## Target Audience

- Terraform engineers
- DevOps teams in medium to large organizations

## Pain Points Addressed

- **Complex Plans:** Large and intricate Terraform plans make it easy to miss critical changes, such as resource deletions.
- **Tedious Code Reviews:** Reviewing pull requests often involves sifting through lengthy textual diffs.
- **Lack of Centralized Visualization:** No single tool offers a clear, visual representation of all infrastructure modifications.

## Key Features (Planned)

| Feature             | Description                                                        | Benefit                                           |
|---------------------|--------------------------------------------------------------------|---------------------------------------------------|
| Delta Matrix        | Line-by-line comparison with color-coded Add/Change/Delete markers | Instant understanding of changes                  |
| Resource Graph      | Visualization of resource dependencies as a graph diagram            | Comprehension of affected resources               |
| Filtering & Search  | Filter by resource type, tags, modules                             | Quick access to relevant objects                  |
| Git Integration     | Automatic PR comments with analysis results                        | Out-of-the-box code review enhancement            |
| Change History      | Versioning of state deltas                                         | Traceability of who changed what and when         |
| CLI + Web UI        | Local CLI tool and a user-friendly SPA dashboard                   | Flexible usage options                            |
| IDE Plugin          | Highlight changes directly in VSCode/GoLand                        | Faster reviews directly within the code           |

## Current Status: MVP - CLI Tool

The first component of InfraSync is a Command Line Interface (CLI) tool.

**Functionality:**
- Reads a Terraform plan JSON file (generated via `terraform show -json <plan_file>`).
- Parses the plan to identify resource changes.
- Outputs a colored textual diff to the terminal, highlighting:
    - **Creations** (green `+`)
    - **Deletions** (red `-`)
    - **Updates** (yellow `~`)
    - Details of attribute changes within updated resources.
    - Handles sensitive values by displaying `(sensitive)` instead of actual data.
    - Identifies values that will be known only after apply (`(known after apply)`).

**Usage:**

1.  Generate your Terraform plan and output it to a binary file:
    ```bash
    terraform plan -out=tfplan
    ```
2.  Convert the binary plan to JSON:
    ```bash
    terraform show -json tfplan > tfplan.json
    ```
3.  Run the InfraSync CLI:
    ```bash
    go run cmd/infrasyncli/main.go tfplan.json
    ```

**Example Output:**
```
~ example_resource.this (example_type)
    ~ tags:
        ~ env: dev -> prod
        ~ team: old_team -> new_team
    + new_attr: was_added
    • computed_value: (known after apply)
    - old_attr: will_be_removed
    ~ sensitive_data: (sensitive) -> (sensitive)
    ~ name: old_name -> new_name
+ another_resource.created (another_type)
- yet_another_resource.deleted (yet_another_type)
```

## Architecture and Technology Stack (Planned)

1.  **Backend:**
    *   Language: Go (Terraform SDK)
    *   Parsing: `terraform show -json`, deserialization into Go structs
    *   API: REST or gRPC to serve CLI and Web UI
    *   Storage: Simple DB like SQLite/PostgreSQL (for state snapshots, deltas)
2.  **Frontend (Web UI):**
    *   Framework: React (TypeScript)
    *   Delta Visualization: Table components + d3.js for dependency graph
    *   Auth: OAuth/GitHub Apps for PR comments
3.  **CLI Utility:**
    *   Wrapper around backend API (future), direct parsing (current)
    *   Supports local mode (without a server)
4.  **Integrations:**
    *   GitHub/GitLab/Bitbucket: Webhooks for automatic report generation
    *   CI: Pipeline step to publish `infrasync diff` in artifacts
5.  **Open Source:**
    *   Repository on GitHub with an MIT License
    *   Modular architecture for community plugins (e.g., Ansible, Pulumi)

## MVP Roadmap

1.  **CLI Tool (✅ Implemented):**
    *   Read local `terraform plan -out=tfplan` (via JSON conversion).
    *   Output a colored textual delta.
2.  **Basic Web UI:**
    *   Upload JSON plan.
    *   Display Add/Change/Delete table.
3.  **GitHub Action:**
    *   Automatic PR comment with a link to the UI report.

## Full Roadmap

| Stage | Timeline   | Goal                                                      |
|-------|------------|-----------------------------------------------------------|
| Alpha | 1–2 months | CLI demo + Basic UI + GitHub Action (MVP)                 |
| Beta  | 3–4 months | Dependency Graph + Filtering + PostgreSQL integration     |
| v1.0  | 5–6 months | IDE Plugin + GitLab/GitHub Enterprise integration         |
| v2.0  | 7–9 months | Expansion to Pulumi/Ansible + Multi-region state support  |

## OSS Model and Community

*   **License:** MIT (or Apache 2.0 - to be finalized)
*   **Contribution Areas (Future):** Support for cloud providers (Azure, GCP), plugins for other IaC tools (Pulumi, Ansible).
*   **Documentation (Planned):**
    *   Quick Start (CLI + Web)
    *   How-to guides for integrations.
*   **Release Cadence (Planned):** Monthly MVP features, weekly bug fixes.

## How to Contribute

Details on how to contribute will be added soon. In the meantime, feel free to open issues for bugs, feature requests, or suggestions.

## License

This project is planned to be licensed under the MIT License. (License file to be added) 
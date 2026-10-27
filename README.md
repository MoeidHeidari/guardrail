# Guardrail

[![View on KikPlate](https://img.shields.io/static/v1?label=KikPlate&message=golang-cli-starter&color=0366d6&style=flat-square)](https://github.com/kikplate/golang-cli-starter)

**Real-Time Engineering Standards for AI Coding Agents**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/MoeidHeidari/guardrail)](https://goreportcard.com/report/github.com/MoeidHeidari/guardrail)

---

## The Problem

AI coding agents can generate working code that violates a project's engineering standards, introducing oversized files, complex functions, unwanted comments, or architectural violations.

Files such as `AGENTS.md` provide instructions, but they don't guarantee compliance. Standards drift, and by the time a PR is opened, violations are already embedded in the codebase.

---

## The Idea

**Guardrail** is an open-source CLI that monitors code changes in real time and verifies that they follow repository-defined engineering rules.

Maintainers define their standards in a `.guardrail.yml` file. Contributors run Guardrail locally while using their preferred AI coding agent.

When a violation occurs, Guardrail reports the exact file and location, explains the violated rule, and optionally provides feedback to the agent so it can fix the issue.

---

## Example Rules

- No comments in source code.
- Maximum 300 lines per file.
- Maximum 40 lines per function.
- Prevent forbidden dependencies between architectural layers.
- Enforce project-specific coding conventions.

---

## Proposed Architecture

| Component | Description |
|---|---|
| **File Watcher** | Detects file changes and triggers incremental analysis. |
| **Rule Engine** | Loads `.guardrail.yml` and evaluates applicable rules. |
| **Language Analysis** | Uses Tree-sitter for syntax-aware checks, with tools such as ast-grep, Semgrep, and native language analyzers. |
| **Architecture Analysis** | Tracks dependencies between files and modules to detect architectural violations. |
| **Live Monitoring** | Displays violations, severity, exact locations, and resolved issues in a continuously updated terminal interface. |
| **Agent Integration** | Sends actionable feedback to supported AI agents through hooks and integrations. |
| **CI Integration** | Reuses the same rules to check pull requests and prevent new violations from being merged. |

---

## Quick Start

```bash
go mod tidy
make build
./bin/guardrail --help
```

---

## Commands

```bash
guardrail greet Alice
guardrail greet --name Bob --shout
guardrail version
guardrail watch        # Start real-time monitoring
guardrail rules        # List configured rules
```

---

## Configuration

Configuration precedence:

1. CLI flags
2. Environment variables (`GUARDRAIL_*`)
3. Config file (`$HOME/.guardrail.yml` or `--config`)
4. Defaults

### `.guardrail.yml` Example

```yaml
rules:
  - id: no-unwanted-comments
    severity: error
    match: "*.go"
    pattern: "//.*"
    message: "Comments are not allowed in source code. Use godoc for documentation."

  - id: max-file-length
    severity: warning
    match: "*.go"
    max_lines: 300
    message: "File exceeds 300 lines. Consider splitting into smaller files."

  - id: max-function-length
    severity: error
    match: "*.go"
    max_lines: 40
    message: "Function exceeds 40 lines. Consider extracting helper functions."

  - id: forbidden-dependency
    severity: error
    match: "*.go"
    from: "internal/presentation"
    to: "internal/data"
    message: "Direct dependency from presentation to data layer is forbidden."
```

---

## Rules Format

Each rule in `.guardrail.yml`:

| Field | Description |
|---|---|
| `id` | Unique rule identifier |
| `severity` | `error`, `warning`, or `info` |
| `match` | Glob pattern for files to check (e.g., `*.go`, `*.ts`) |
| `max_lines` | Maximum allowed lines (for length rules) |
| `from` / `to` | Architectural layer boundaries (for dependency rules) |
| `pattern` | Tree-sitter/ast-grep pattern to match |
| `message` | Human-readable explanation of the violation |

---

## Live Monitoring

Run Guardrail in watch mode to monitor your codebase in real time:

```bash
guardrail watch
```

The terminal will display violations as they occur, with:

- Exact file path and line number
- Rule ID and severity
- Description of the violated rule
- Optionally, auto-fix suggestions

---

## Agent Integration

Guardrail can send feedback to supported AI agents via integration hooks. Currently supported:

- **Claude Code** - through subprocess hooks
- **Cursor** - through project-level configuration
- **Custom agents** - via JSON-RPC compatible endpoints

Configure agent integration in `.guardrail.yml`:

```yaml
agent_integration:
  enabled: true
  agent: "claude-code"
  endpoint: "http://localhost:9876/feedback"
```

---

## CI Integration

The same rules used for local development are reused in CI/CD to prevent violations from being merged:

```bash
guardrail ci --pull-request 123
```

Or integrate with GitHub Actions:

```yaml
name: Guardrail CI

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  guardrail:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Guardrail
        run: guardrail ci
```

---

## Project Structure

```text
guardrail/
├ cmd/          # CLI entry points and subcommands
├ internal/     # Internal packages (config, logger)
├ pkg/          # Public packages
├ .guardrail.yml # Repository engineering standards
├ Makefile
├ go.mod
├ main.go
└ README.md
```

---

## Development

```bash
make build     # Build the binary
make test      # Run tests with race detection
make lint      # Run golangci-lint
```

---

## Roadmap

- [x] Initial CLI scaffold (Cobra + Viper + Zap)
- [x] Config file support (.guardrail.yml)
- [ ] Tree-sitter integration for syntax-aware parsing
- [ ] Rule engine with pattern matching
- [ ] File watcher for real-time monitoring
- [ ] Architecture dependency tracking
- [ ] Agent integration (Claude Code, Cursor)
- [ ] GitHub Actions CI integration
- [ ] Web dashboard for violation history
- [ ] Plugin system for custom rules

---

## Contributing

We welcome contributions! See the [ contributing guidelines ](CONTRIBUTING.md) for more details.

- Fork the repository
- Create a feature branch
- Add tests for new rules or functionality
- Run `make lint` and `make test`
- Submit a pull request

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
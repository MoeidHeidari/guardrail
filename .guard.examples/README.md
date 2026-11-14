# Rule examples

These are individual YAML rule examples and JavaScript functions, not a starter project.
They mirror a project's `.guard/` layout. Copy selected sets and scripts from
`.guard.examples/` into your project's `.guard/`, preserving their paths.

| Example | Demonstrates | Copy to your project |
| --- | --- | --- |
| [general/quality.yml](general/quality.yml) | Regex conflict-marker rule | `.guard/general/quality.yml` (general set) |
| [frontend/quality.yml](frontend/quality.yml) | Scoped regex and inline boolean JavaScript rules | `.guard/frontend/quality.yml` (frontend set) |
| [maintainability/size.yml](maintainability/size.yml) | Referenced JavaScript rule | `.guard/maintainability/size.yml` (maintainability set) |
| [source-lines.js](scripts/source-lines.js) | Structured findings with line, column, message | `.guard/scripts/source-lines.js` |
| [inline/quality.yml](inline/quality.yml) | Single YAML combining regex and inline JavaScript | `.guard/inline/quality.yml` (inline set) |

Copy only the examples you want and adjust their scopes and thresholds.
The maintainability YAML needs the referenced function at the exact destination
shown above. All paths in YAML are relative to the context root, not the YAML file.

## Rule sets

The YAML's containing folder determines the set. A YAML directly inside `.guard/`
belongs to **General**; `.guard/frontend/*.yml` belongs to **frontend**.
Multiple YAML files in one folder form one set. IDs must be unique within that
set but can be reused in a different set. Every YAML inside `.guard/` is active.
Use `description` to explain the validation and `message` for the finding.

## JavaScript functions

Functions receive `(path, change)` with a context-relative path and the saved-file
snapshot. Use `change.content` for text, `change.lineCount` for length,
the high-level helpers for common syntax, or `change.ast` for raw AST access.
Inline functions use `code: |`; external functions use `file:` and contain one
function expression without module exports or imports.

- Return `false` to pass or `true` for a file-level violation.
- Return `[]` to pass or `[{line, column, message}]` for located findings.
- Line and column positions are 1-based; columns count Unicode code points.
- Use one limit variable for the threshold, finding position, and message.
- No filesystem, networking, processes, modules, or Node APIs are available.
  Node is not required. Run only trusted functions; this is not a memory sandbox.

Regex rules use Go RE2 syntax. Text checks can match comments and strings.
Invalid results, exceptions, and timeouts are analysis errors, not passing checks.

Examples are not activated merely by being in this folder. Guardrail reads
`.guard/` under the context you select. See [the rule contract](../docs/RULES.md)
for fields, scopes, limits, and runtime restrictions.

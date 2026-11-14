# Repository hygiene

[Community catalog](../README.md) | [Rule API](../../docs/RULES.md)

Three standalone JavaScript functions for saved text files. Maintained by
Guardrail contributors under [Apache License 2.0](../../LICENSE).
Requires the version 1 YAML schema and located JavaScript findings API;
tested against the repository source containing this pack.

| Rule | Behavior |
| --- | --- |
| `no-conflict-markers` | Reports full-line seven-character Git conflict markers, including diff3 ancestor markers, at column 1 |
| `no-trailing-whitespace` | Reports each line ending in spaces or tabs at its first trailing Unicode-code-point column |
| `require-final-newline` | Reports a missing final newline at the last line's end boundary; empty files pass |

## Install

Choose the general `repository-hygiene.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/general/repository-hygiene.yml` in your context, and copy the shared
[functions](../functions/repository-hygiene/) to `.guard/functions/repository-hygiene/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Use the general rule set; its initial scope is text files. Customize include patterns.
Standard analysis exclusions remain in effect.

## Limits and customization

LF and CRLF are supported, not bare-CR line endings. These are text policies,
not syntax-aware checks: conflict examples in strings or documentation may be
flagged, and nonstandard conflict-marker widths are not recognized.
Trailing spaces can be intentional Markdown line breaks or meaningful data;
exclude such files or omit that rule. Customize the patterns in your local copy.
More than 1,000 findings from one function/file is an explicit runtime error,
not a truncated passing result.

## Tests

Synthetic fixtures live in
[the Go regression suite](../../internal/analysis/rulepacks_test.go), including
Unicode columns, CRLF, empty files, clean inputs, and multiple conflict markers.

```sh
go test ./internal/analysis -run '^TestCommunityRulePacks$'
```

The suite copies actual YAML/functions into disposable `.guard/` directories
and checks exact finding IDs, locations, and kinds through the Guardrail engine.
Do not execute community scripts you have not reviewed; Goja is not a secure
process/memory sandbox.

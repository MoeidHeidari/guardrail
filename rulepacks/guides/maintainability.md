# Maintainability

[Community catalog](../README.md) | [Rule API](../../docs/RULES.md)

The standalone `max-file-lines` JavaScript function enforces a team-selected
file-length policy. Maintained by Guardrail contributors under
[Apache License 2.0](../../LICENSE). Requires version 1 YAML and change API v1; tested against repository source containing
this pack.

## Install

Choose the general `maintainability.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/general/maintainability.yml` in your context, and copy the shared
[functions](../functions/maintainability/) to `.guard/functions/maintainability/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Use the general rule set; its initial scope is text files. Customize include patterns.
Standard analysis exclusions remain in effect.

## Policy and limits

Edit `maximum` in your installed function to customize the initial 300-line
limit. It must be a positive integer. This is an opt-in policy, not a built-in
runtime limit. Empty files have zero lines; a final LF/CRLF terminates the last
line without adding another. Blank lines count. Bare-CR endings are unsupported.

Files at the limit pass. Longer files produce one violation at line 301,
column 1 with the default setting, and a message containing actual and allowed
counts. This measures file length, not complexity or maintainability quality.

## Tests

[Synthetic regression fixtures](../../internal/analysis/rulepacks_test.go)
cover empty content, 299/300/301 lines, LF/CRLF, and files without final newlines.

```sh
go test ./internal/analysis -run '^TestCommunityRulePacks$'
```

Tests use the actual installed YAML and function through Guardrail's engine.
Review contributed scripts before execution; runtime restrictions do not
provide a secure process/memory sandbox.

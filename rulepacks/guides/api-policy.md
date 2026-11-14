# Public function documentation

[Catalog](../README.md) | [Syntax API](../../docs/RULES.md#syntax-api)

`document-public-functions` enforces documentation for explicitly public
function declarations using actual syntax, not text searches.

| Language | Declaration scope and accepted documentation |
| --- | --- |
| Go | Uppercase-named functions/methods; adjacent comment group starting with the function name |
| JavaScript / TypeScript / TSX | Directly exported named function declarations; adjacent JSDoc |
| Python | Top-level functions not starting with `_`, including decorated functions; first body statement is a string docstring |
| Rust | Explicit unrestricted `pub fn`; adjacent outer `///` or `/**` documentation |

## Install

Choose the language-specific `api-policy.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/<language>/api-policy.yml` in your context, and copy the shared
[functions](../functions/api-policy/) to `.guard/functions/api-policy/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Each language rule set includes only that language's file patterns.
Standard analysis exclusions remain in effect.

## Behavior and limits

An undocumented public function fails at its name's location. Adding the
language's accepted documentation makes it pass. Private declarations and
strings elsewhere in the body are not substitutes for public documentation.

This first rule does not cover classes/types, JS exported function-valued
variables, re-exports, class methods, Python methods or `__all__`, Rust trait
signatures or restricted `pub(crate)`, or documentation attributes separated
from Rust declarations. It checks presence, not documentation quality or
semantic API visibility. Do not mistake it for complete public API coverage.

## Validation and contribution

Run `go test ./internal/analysis -run '^TestEngineering'`.
[Tests](../../internal/analysis/engineering_packs_test.go) include documented and
undocumented declarations in every supported language. Unavailable or malformed
syntax fails explicitly.

Maintained by Guardrail contributors under [Apache License 2.0](../../LICENSE).
Requires change API version 1 with raw AST access from this repository. Review contributed scripts;
runtime limits do not provide a secure sandbox.

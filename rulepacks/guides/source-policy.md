# Source API policy

[Catalog](../README.md) | [Syntax API](../../docs/RULES.md#syntax-api)

`no-debug-output` rejects explicit debug/output calls in production source.
Edit the `prohibited` table to choose your team's forbidden APIs.

| Language | Initial prohibited spellings |
| --- | --- |
| Go | `fmt.Println`, `fmt.Printf`, `fmt.Print`, `print`, `println` |
| JavaScript / TypeScript / TSX | `console.log`, `console.debug`, `console.trace`, `debugger` |
| Python | `print`, `breakpoint`, `pdb.set_trace` |
| Rust | `println!`, `print!`, `eprintln!`, `eprint!`, `dbg!` |

## Install

Choose the language-specific `source-policy.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/<language>/source-policy.yml` in your context, and copy the shared
[functions](../functions/source-policy/) to `.guard/functions/source-policy/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Each language rule set includes only that language's file patterns.
Standard analysis exclusions remain in effect.

## Behavior and limits

Findings point to the call, macro invocation, or debugger statement.
`console.log(value)` fails; `"console.log(value)"` and comments containing it
pass. Syntax-level calls are inspected, not arbitrary text.

This is not type or symbol analysis. Shadowed identifiers with matching spellings
are flagged. Renamed imports, assigned function aliases, computed member access,
and fully qualified Rust macro paths are not recognized by the initial table.
Add exact spellings or build a symbol-resolution feature rather than claiming
this rule detects every call to an API.

## Validation and contribution

Run `go test ./internal/analysis -run '^TestEngineering'`.
[Fixtures](../../internal/analysis/engineering_packs_test.go) cover prohibited calls
and harmless strings in all supported languages. Invalid/unavailable syntax
fails explicitly.

Maintained by Guardrail contributors under [Apache License 2.0](../../LICENSE).
Requires syntax API version 1 from this repository. Review code before execution;
runtime host restrictions do not constitute a secure sandbox.

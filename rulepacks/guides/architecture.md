# Architecture boundaries

[Catalog](../README.md) | [Syntax API](../../docs/RULES.md#syntax-api)

`no-infrastructure-imports` enforces a dependency-direction convention: files
under a `domain/` or `core/` path segment must not import modules with a path
segment named `infrastructure`, `adapters`, or `transport`. Edit `domainPath`
and `forbidden` in the function to match your architecture.

## Install

Choose the language-specific `architecture.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/<language>/architecture.yml` in your context, and copy the shared
[functions](../functions/architecture/) to `.guard/functions/architecture/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Each language rule set includes only that language's file patterns.
Standard analysis exclusions remain in effect.

## Behavior and limits

The rule inspects syntax nodes for Go imports, JavaScript imports/re-exports and
literal `require()`/`import()` calls, Python imports, and Rust `use` declarations.
`domain/service.go` importing `"app/infrastructure/db"` fails at the import
source location; importing `"app/domain/model"` passes.
Comments and strings elsewhere in code are not imports.

This is an import-spelling policy, not a resolved dependency graph. It does not
resolve relative paths, aliases, wildcard imports, dynamic expressions, module
remapping, or transitive dependencies, and cannot detect cycles. For example,
an alias that hides a forbidden module name is not caught. Rename/edit patterns
to avoid false positives for unrelated modules with the same segment names.
Rust alias names in `use` declarations are also included in the spelling check.

## Validation and contribution

Run `go test ./internal/analysis -run '^TestEngineering'`.
[Regression cases](../../internal/analysis/engineering_packs_test.go) use the actual
pack through Guardrail and cover import locations for all supported grammars.
Missing or invalid syntax produces explicit errors, not a passing result.

Maintained by Guardrail contributors under [Apache License 2.0](../../LICENSE).
Requires change API version 1 with raw AST access from the repository containing this pack.
Review functions before executing them; JavaScript is not a secure sandbox.

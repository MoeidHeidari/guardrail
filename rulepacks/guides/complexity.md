# Function complexity budgets

[Catalog](../README.md) | [Syntax API](../../docs/RULES.md#syntax-api)

`function-budgets` applies three configurable limits to real function syntax:
60 lines, 5 parameters, and 3 nested control-flow constructs.
Edit `limits` in the function to adopt your team's thresholds.

## Install

Choose the language-specific `complexity.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/<language>/complexity.yml` in your context, and copy the shared
[functions](../functions/complexity/) to `.guard/functions/complexity/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Each language rule set includes only that language's file patterns.
Standard analysis exclusions remain in effect.

## Measurements

- Line span includes the declaration and its body, comments and blank lines.
- Parameters count named parameter syntax, including receivers such as Python
  `self` and Rust `self`. Go method receivers are outside the parameter list.
  Grouped Go names each count; destructured parameters count as one.
- Nesting counts nested if/loop/switch/select/try/with/match constructs. It is
  syntactic depth, not cyclomatic or cognitive complexity. Else-if syntax can
  add depth. The `controls` list documents the exact node kinds.
- Nested function bodies are measured separately, not included in their parent's
  nesting. Their lines still contribute to the parent's source span.
- Declaration-only signatures with no body are not measured.

One finding at the function's start describes every exceeded budget.
Exactly 60 lines, 5 parameters, or depth 3 passes; exceeding a limit fails.
This is a policy signal, not proof of poor design or correctness.

## Validation and contribution

Run `go test ./internal/analysis -run '^TestEngineering'`.
[Synthetic tests](../../internal/analysis/engineering_packs_test.go) cover all
supported grammars and threshold boundaries. Unsupported or invalid syntax
fails explicitly.

Maintained by Guardrail contributors under [Apache License 2.0](../../LICENSE).
Requires syntax API version 1 from this repository. Review functions before
executing them; JavaScript is not a secure process/memory sandbox.

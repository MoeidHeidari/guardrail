# Community engineering-standard rule packs

[Project README](../README.md) | [Contributing](../docs/CONTRIBUTING.md) | [Rule API](../docs/RULES.md)

This directory is the contribution home for reusable Guardrail rules: small,
documented packs of engineering standards that teams can review and adopt.
Packs are opt-in examples, not built-in policies. Nothing in this directory is
automatically enabled; Guardrail loads active YAML from the chosen context's
`.guard/` directory.

The library includes locally adoptable JavaScript starter packs. Contributions
are welcome. There is no package registry, remote installer, pack manifest
schema, or independent pack-versioning system implemented today. For now,
identify the repository commit you copied when recording a pack's provenance.

## Suggested contributions

| Category | Initial scope |
| --- | --- |
| `repository-hygiene` | Conflict markers, trailing whitespace, final newlines |
| `maintainability` | File-length policies and other bounded text checks |
| `architecture` | Domain/core import boundaries |
| `source-policy` | Explicit prohibited API calls |
| `complexity` | Function size, parameters, and nesting |
| `api-policy` | Public function documentation |

The catalog below lists available packs; other entries are suggestions. Syntax
rules can inspect real functions and imports through the read-only AST API.
Type-aware checks, resolved import graphs, and cross-file policies still need
engine features that do not exist yet. Do not approximate semantic checks with regex without
clearly documenting their limitations and false positives.

## Contribute a pack

Follow the project's [contribution workflow](../docs/CONTRIBUTING.md).
Search existing proposals before creating a pack. Open a feature issue for
substantial new scope, then submit a focused PR containing:

```text
rulepacks/
  go/
    your-category.yml
  python/
    your-category.yml
  functions/
    your-category/
      your-rule.js
  guides/
    your-category.md
```

Use language folders for YAML, shared category folders for reusable JavaScript,
and category guides for documentation. Add other language variants only when
tested; never copy functions per language when the implementation is shared.
Keep generic text-file policies in `general/`.
Use lowercase hyphenated category and rule names. Inline JavaScript or regex rules
do not need a `functions/` directory. Keep executable rules separate from
fixtures, so adopting a pack does not copy its sample inputs into active rules.
Do not put pack contributions in this repository's `.guard/`; doing so would
activate them for Guardrail's own development.
Fixtures may instead be synthetic cases in an executable regression suite, as
the starter packs demonstrate; link that suite from the category guide.

### Required documentation

Each category guide must explain:

- Purpose, intended audience, author/contact, and supported Guardrail
  version or tested commit.
- Rule IDs, descriptions, languages/file patterns, and expected diagnostics.
- Exact installation destinations, including every referenced function.
- Include/exclude scope and how to customize policy values. There is no
  runtime rule-options API yet; explain which YAML or function values to edit.
- Synthetic passing/failing fixtures (files or executable test cases), exact validation commands and results,
  expected finding counts/locations, and known false positives or limitations.
- Attribution and licensing. New contributions use the repository's
  [Apache License 2.0](../LICENSE); preserve required notices for reused material
  and only contribute code you have permission to distribute.

IDs must be unique within their rule set. Guardrail currently uses the immediate
containing folder's name as the set; a rule such as `function-budgets` installed
under `.guard/go/` is identified by set `go` and that ID. IDs across different
category YAML files within one language must not collide.
Do not change an existing rule's meaning silently; describe compatibility and
migration implications in the PR.

Add the pack to the catalog below when contributing it.

## Function contract

Use the existing [rule API](../docs/RULES.md#javascript-rules), not a new module API:

- A function receives `(path, change)` for one complete saved-file snapshot,
  including during `check`; this is not a diff.
- Referenced JavaScript is a function expression, without imports or exports.
- Return `false` or `[]` for no findings, `true` for a file-level violation, or
  an array of `{ line, column, message }` findings.
- Located findings use 1-based lines and Unicode-code-point columns.
- `change.content`, `lines`, and `lineCount` support text checks.
- `change.functions()`, `imports()`, `comments()`, and `calls()` provide common
  models, source locations, and `finding(message)`. See the [change API](../docs/RULES.md#change-api).
- Advanced rules use `change.ast`; call its `requireAST()` before raw syntax
  checks. See the [syntax API](../docs/RULES.md#syntax-api).
- There are no filesystem/network APIs, type information, or cross-file access.
- Exceptions, invalid return values, and timeouts are execution errors.

Keep checks deterministic and bounded. Respect the documented input, timeout,
and finding limits. Keep contributed source and test files at or below 300 lines.

## Start with an existing function

See the [maintainability guide](guides/maintainability.md) for a minimal standalone
function and its YAML reference. Its installation instructions show how to copy
both files into your project's active rules. Reuse that layout when contributing
another function; do not add module exports or dependencies.

## Validate and adopt

Use separate disposable contexts for passing and failing fixtures. Copy only
the pack's rule YAML and required functions into each context's `.guard/`.
Run the built Guardrail executable against each context:

```sh
guardrail check /path/to/passing-context --format json
guardrail check /path/to/failing-context --format json
```

Passing fixtures must exit `0`; failing fixtures should exit `1` with exactly the
documented policy findings. Exit `2` means an analysis/configuration/execution
error, not a successful rule test. Include Unicode, CRLF, empty files, and
boundary cases where relevant. No dedicated pack-test runner exists yet.

For example, from a Guardrail checkout, to adopt Go architecture rules:

```sh
mkdir -p /path/to/project/.guard/go /path/to/project/.guard/functions/architecture
cp rulepacks/go/architecture.yml /path/to/project/.guard/go/
cp rulepacks/functions/architecture/*.js /path/to/project/.guard/functions/architecture/
guardrail check /path/to/project --format json
```

Replace `/path/to/project` with your context. Review destinations before copying;
`cp` overwrites matching names. Select only the YAML categories you want.
Other language rule sets follow the same layout. npm dependency rules live
under `javascript/` and also apply to TypeScript projects' `package.json` files.
Generic text policies live under `general/`; customize their file patterns.

YAML references already point to `.guard/functions/<category>/<function>.js`.
Paths are relative to the context root, not the YAML file. Copy the corresponding
shared functions with each selected category. Review include/exclude scope before
enabling rules. Avoid installing a second copy of the same ID into one language
set, especially when adopting into existing configuration.

**Run only trusted rules.** JavaScript executes inside Guardrail. Restricted
host APIs and timeouts are not a secure process/memory sandbox. Community
contributions are not security certifications. Review code and changes before
copying or updating it, and use synthetic fixtures rather than private data.

## Catalog

### Language rule sets

| Folder | Rule sets |
| --- | --- |
| [Go](go/) | Architecture, source policy, complexity, API policy |
| [JavaScript](javascript/) | Architecture, source policy, complexity, API policy, npm dependencies |
| [TypeScript](typescript/) | Architecture, source policy, complexity, API policy (includes TSX) |
| [Python](python/) | Architecture, source policy, complexity, API policy |
| [Rust](rust/) | Architecture, source policy, complexity, API policy |
| [General](general/) | Repository hygiene and file-length policy; initially text files |

### Shared implementations

Each source-language folder also includes `no-comments.yml`. It rejects ordinary
comments while allowing documented language-specific exceptions. Install the
shared [comment function](functions/source-policy/no-comments.js) under
`.guard/functions/source-policy/`. See the [comment-policy guide](guides/no-comments.md).

| Guide | JavaScript functions | Compatibility |
| --- | --- | --- |
| [Repository hygiene](guides/repository-hygiene.md) | Conflict markers, trailing whitespace, final newline | Version 1 YAML and located JS findings |
| [Maintainability](guides/maintainability.md) | Configurable maximum file length | Version 1 YAML and located JS findings |
| [Architecture](guides/architecture.md) | Domain/core import boundaries | Syntax API v1; all bundled source grammars |
| [Source policy](guides/source-policy.md) | Prohibited debug-output APIs | Syntax API v1; all bundled source grammars |
| [Complexity](guides/complexity.md) | Function size, parameters, control-flow nesting | Syntax API v1; all bundled source grammars |
| [API policy](guides/api-policy.md) | Public function documentation | Syntax API v1; all bundled source grammars |
| [Dependency policy](guides/dependency-policy.md) | Configurable npm dependency denylist | Version 1 YAML and located JS findings |

Text packs are tested against repository source through
[engine regression tests](../internal/analysis/rulepacks_test.go).
Run `go test ./internal/analysis -run '^TestCommunityRulePacks$'`.
Engineering packs use [syntax-based regression tests](../internal/analysis/engineering_packs_test.go).
Run `go test ./internal/analysis -run '^Test(Engineering|Dependency)'`.
Contribute additional packs with a README link, summary, and tested version or commit.

### Helper and raw-AST examples

[Debug-output policy](functions/source-policy/no-debug-output.js) uses
`change.calls()` without grammar-specific field traversal. Its
[raw-AST counterpart](functions/source-policy/no-debug-output-ast.js) demonstrates
the same policy through `change.ast`. The default YAML selects the helper version;
replace its `file` reference with the AST counterpart to experiment, rather than
enabling both and producing duplicate findings. Regression tests check parity.
The architecture and comment policies also demonstrate combining helpers with
raw nodes for language-specific cases.

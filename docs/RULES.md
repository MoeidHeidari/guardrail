# Rules and checking

[Documentation index](README.md) | [Watch workspace](WATCH.md)

Guardrail automatically loads every `.yml` and `.yaml` file recursively under
the context's `.guard/` folder. No policy is built in.

## Rule sets

For reusable, opt-in engineering standards and contribution guidance, see the
[community rule-pack library](../rulepacks/README.md). Library files are not
automatically loaded; review and copy selected rules into your context's `.guard/`.

Each YAML file uses the schema below. Root-level YAML belongs to **General**;
nested YAML uses its immediate containing folder's name as the rule-set name.
For example:

```text
.guard/
  rules.yml                    -> General
  go/quality.yaml              -> go
  maintainability/source.yml   -> maintainability
  rules/max-source-lines.js    -> referenced script, not a rule set
```

Multiple YAML files in the same folder contribute to the same set. IDs must be
unique within a set, but different sets may reuse IDs. Files load in lexical
order. Each file retains its own include/exclude scope and execution timeout.
Paths in YAML (including JS references) are relative to the context root,
not the YAML file. Symlink YAML files are rejected; symlink folders are not
followed. Any invalid set is an explicit configuration error, not partially
successful analysis.

This project adopts the community [Go rule sets](../.guard/go/) for architecture,
source APIs, complexity, comment policy, repository hygiene, and file
length. All active sets select Go files, including tests. The installed
[file-length function](../.guard/functions/maintainability/max-file-lines.js)
uses a 250-line limit; the community default remains 300. These are opt-in
policies, not built-in engine restrictions. The npm dependency policy is not
enabled because this repository uses Go dependencies.

## Configuration

```yaml
version: 1
timeout_ms: 100
include:
  - "**/*.go"
  - "**/*.js"
rules:
  - id: conflict-markers
    type: regex
    pattern: '(?m)^(<<<<<<< |=======$|>>>>>>> )'
    message: Resolve merge conflict markers.
  - id: max-source-lines
    type: javascript
    include:
      - "**/*.go"
    file: .guard/functions/maintainability/max-file-lines.js
```

Rule IDs must be nonempty and unique within their rule set. Unknown YAML fields, multiple documents,
invalid globs, invalid regexes, and invalid JavaScript are errors. At least one
rule and `version: 1` are required. `message` is optional; the fallback names the
rule. Regex rules require `pattern`; JavaScript rules require exactly one of
`code` or `file`.
An optional `description` explains the validation independently of the violation
`message`. The Rules page shows the selected rule's description across the full
page width, wrapping onto two lines; existing rules fall back to their message.

Patterns match context-relative paths with `/` separators. `**` matches any
number of directories, including zero; `*` matches within one directory.
Rule-level `include` further restricts the global scope.

When omitted, `include` selects Go, JavaScript, TypeScript, Python, Rust, JSON,
YAML, TOML, shell, CSS, HTML, Dockerfiles, and Makefiles. Default `exclude` omits
Git internals, rule folders, `node_modules`, `vendor`, Python environments and
caches, `dist`, `build`, `target`, and `bin` directories. Explicit include/exclude
lists replace their corresponding defaults. The root `.git/`, `.guard/`,
`.guardrail/` are always excluded from analysis.

Browsing and watching include those paths unless excluded by the root
`.guardignore`. `check` honors the same ignore file. Analysis does not follow
symlink files. Matching files must be UTF-8 regular files without NUL bytes and
at most 4 MiB; unsupported matching content produces a visible read error.
Configuration is limited to 1 MiB and JavaScript source to 64 KiB.

## Regex rules

Patterns use Go's RE2 syntax, not JavaScript regex syntax. Lookbehind and
backreferences are not supported. A match is a violation, including each
occurrence in a file. Use `(?m)` for multiline anchors and `(?i)` for case
insensitivity. More than 1,000 matches per file/rule produces an execution
error rather than silently truncating the report.

## JavaScript rules

Functions receive exactly `(path, change)`: a context-relative path and the
read-only change object below. This replaces the earlier content/syntax arguments;
migrate text access to `change.content` and raw syntax access to `change.ast`.
Despite its name, `change` is a complete saved-file snapshot, not an edit diff.
It is also supplied by one-shot `check`, without requiring a change event.

A referenced file must contain one function expression, without module exports:

```javascript
(path, change) => change.lineCount > 300
```

`true` means one file-level violation; `false` means no violation. File-level
findings use line/column `0`. A function may instead return an array:

```javascript
(path, change) => [
  { line: 1, column: 1, message: "Explain the violation here." }
]
```

An empty array passes. Locations are **1-based**, with columns counted in Unicode
code points. Columns may point at the end-of-line boundary. Invalid positions,
unknown finding fields, noninteger positions, unsupported return types, promises,
exceptions, and timeouts are execution errors, not successful checks. Messages
are optional (rule fallback), limited to 4,096 bytes; at most 1,000 findings may
be returned per file/rule.

For inline functions, use a YAML block:

```yaml
version: 1
rules:
  - id: example
    type: javascript
    code: |
      (path, change) => change.content.includes("FIXME")
```

See the [community catalog](../rulepacks/README.md) for actual functions,
language-scoped YAML, installation instructions, and regression fixtures.
Library YAML lives outside `.guard/` so it is not activated. Every YAML inside
`.guard/` is active. The former root `.guardrail.yml` is not loaded.
Referenced scripts must remain inside `.guard/` or
`.guardrail/`; absolute paths, traversal, symlink files, and symlink rule
directories are rejected.

The Goja runtime is embedded; Node.js is not required. Each file/rule gets a
fresh runtime without filesystem, networking, processes, modules, or host APIs.
`timeout_ms` defaults to 100, accepts 1–10,000, and covers initialization,
function execution, and result serialization. JavaScript loops are interrupted
on timeout/cancellation; call-stack depth is limited.

**Run only trusted repository scripts.** This is not a process or memory
sandbox. Scripts can consume memory and runtime-native work is not guaranteed
to be interruptible. Host API restrictions do not make untrusted code safe.

### Change API

`change.version` is `1`. `content` holds saved source text; `language` identifies
the grammar. `lines` is a lazily built, frozen array without LF/CRLF terminators;
`lineCount` is its length. Empty content has zero lines, a final newline does not
create an extra line, and blank lines count. Bare CR is not a line separator.

`functions()`, `imports()`, `comments()`, and `calls()` expose cached, frozen
collections for Go, JavaScript/JSX, TypeScript/TSX, Python, and Rust. Call
`change.require("functions", "calls")` to require capabilities explicitly.
`capabilities` lists supported names. Unsupported helpers/capabilities and
malformed syntax throw errors, not empty passing results.

Every model has `kind` (the raw grammar kind), `text`, `rawNode`, `location`, and
`finding(message)`. Locations contain frozen `start` and exclusive `end` objects:
`{ line, column, offset }`. Lines/columns are 1-based source positions, not screen
coordinates. Columns count Unicode code points; offsets are 0-based UTF-16 for
`change.content.slice`. Tabs count as one source character.
`change.location(rawNode)` creates the same location for an advanced AST node.
`change.finding(message, position)` creates a located finding, and a model's
`finding(message)` uses its start automatically. Final validation checks that
findings stay within the file. Use boolean `true` for an unlocated file finding.

| Helper | Additional fields |
| --- | --- |
| `functions()` | `name`, `nameNode`, `body`, `parameters`, `metrics`, `isPublic`, `documentation` |
| `imports()` | `source`, `sourceNode`, `form`, `alias`, `names`, `isDynamic` |
| `comments()` | `isDocumentation`, `isDirective` |
| `calls()` | `target` (with `spelling`), `arguments`, `isMacro`, `isDebugger` |

Function metrics expose `lineCount`, `parameterCount`, and `nestingDepth`.
The line span includes comments/blank lines; nesting excludes nested functions'
bodies. Parameters include Python/Rust receivers; grouped Go names count
individually, and destructuring counts as one. Go method receivers are outside
the parameter list. These are syntactic budgets, not cyclomatic complexity.
Declaration-only signatures have `body: null`.

Public/documentation classification follows the
[public-function policy](../rulepacks/guides/api-policy.md): uppercase Go names,
directly exported JS/TS functions, top-level non-underscore Python functions,
and unrestricted Rust `pub fn`. It does not determine full semantic visibility.
`documentation` is a comment/docstring model or `null` when absent.

Imports retain source spelling, not resolved module paths. Rust grouped imports
retain their group text; they are not expanded into a dependency graph. `names`
contains named JS/TS import specifiers (`{ name, alias }`); other forms/languages
use `null` for that classification. Namespace/default names remain in raw syntax.
Nonliteral JS `import()`/`require()` calls have `source: null` and
`isDynamic: true`. Alias support is limited to direct grammar fields.

Calls expose target spelling, not resolved symbol identity. Macro arguments
have `arguments: null` when there is no ordinary grammar argument field.
Comments expose boolean `isDocumentation` and `isDirective` flags. Documentation
recognition covers adjacent Go exported-declaration comments, package documentation,
cgo preambles, JSDoc, and Rust documentation markers. Python docstrings are strings,
exposed through function documentation, not `comments()`.
Directive recognition covers Go compiler/build/line/export and nolint comments;
JavaScript/TypeScript typing, JSX, linting, formatting, coverage, and purity markers;
and Python shebang, encoding, typing, linting, coverage, and formatting markers.
Rust currently has no recognized comment directives. These are syntactic heuristics,
not validation of documentation or tool configuration; unrecognized comments have
both flags set to `false`.

```javascript
(path, change) => change.comments()
  .filter(comment => !comment.isDocumentation && !comment.isDirective)
  .map(comment => comment.finding("Remove ordinary comments."))
```

```javascript
(path, change) => change.functions()
  .filter(fn => fn.body && fn.metrics.parameterCount > 5)
  .map(fn => fn.finding("Reduce the number of parameters."))
```

### Syntax API

Advanced users retain the raw API as `change.ast`. In this section, `syntax`
means `change.ast`.

`syntax.version` is `1`. `language` identifies the grammar (`Go`, `JavaScript`,
`TypeScript`, `TSX`, `Python`, `Rust`, or `text`); `parsed` indicates parsing was
attempted successfully. `error` holds any parser/capture failure. Always call
`syntax.requireAST()` before syntax-dependent checks: it throws when the grammar
is unavailable or syntax/capture failed, so these conditions cannot appear as a
successful rule check. Text rules may still run after parser failures.

`syntax.nodes` is a frozen, preorder array of named Tree-sitter nodes. The
nodes, their `children` arrays and `fields` maps, and the API object are frozen.
Each node has:

- `type`: the grammar's node kind, not a language-independent semantic category.
- `start` / `end`: exclusive-end UTF-16 offsets for JavaScript `content.slice`.
- `line` / `column`: 1-based start location; columns count Unicode code points.
- `endLine`: 1-based line of the exclusive end position.
- `parent`: parent node index, or `-1` at the root.
- `children`: named child indices, in source order.
- `fields`: grammar field names mapped to a child index. Repeated fields retain
  their last child; use `children` when every repeated element matters.

Helpers return node objects belonging to this snapshot:

```javascript
(path, change) => {
  const syntax = change.ast;
  syntax.requireAST();
  return syntax.find("call_expression").filter(node => {
    const target = syntax.field(node, "function");
    return target && syntax.text(target) === "fmt.Println";
  }).map(node => ({
    line: node.line,
    column: node.column,
    message: "Use the project's logging abstraction.",
  }));
}
```

`find(...types)` selects node kinds. `text(node)` returns its source text;
`field(node, name)` and `parent(node)` return a node or `null`;
`children(node)` returns a new array of named children. Foreign nodes are rejected.
Node names/fields follow the bundled grammars; test each language explicitly.

The tree is parsed once per file evaluation and captured only when JavaScript
rules apply to that file. Capture is cancellation-aware and limited to 50,000 named
nodes. Exceeding that limit is an explicit parser/capture error; no partial tree
is exposed. Capture observes the analysis deadline, serialization happens once
per evaluation, and JavaScript API initialization and execution observe each
rule's deadline. Each runtime decodes and freezes nodes lazily on first node
access; text-only rules do not pay that initialization cost.

The API provides syntax only: no type checking, symbol/import resolution,
cross-file access, dependency graph, networking, or filesystem access. See the
[engineering packs](../rulepacks/README.md#catalog) for tested uses and limitations.

## Parsing

Tree-sitter parses Go, JavaScript/JSX, TypeScript/TSX, Python, and Rust before
rule evaluation. Syntax errors/missing nodes appear as parser errors with a
location; regex and JavaScript rules still run on the saved content.
Other included extensions are explicitly text-only (`parsed: false` in JSON)
and still support custom rules. Parsing is syntax analysis, not type checking,
dependency resolution, or proof that a program is valid.

## One-shot checking

```sh
guardrail check
guardrail check /path/to/project
guardrail check -c ../project --format json
```

The path defaults to the working directory, like `watch`. JSON includes checked
and skipped counts, per-file content digests, parsing status, and findings with
`path`, `set`, `source`, `rule`, `kind`, `line`, `column`, and `message`. Finding kinds are
`violation` and `error`; configuration failures also produce a report.

Exit codes:

- `0`: no findings in the configured analysis scope.
- `1`: rule violations (or invalid command usage).
- `2`: analysis/configuration/parser/read/execution errors, even when violations
  also exist.

`check` does not open a terminal or allocate filesystem watches.

## Live results

`watch` analyzes the initial context off the UI thread, reuses unaffected
results, and rechecks updated files. Deletes clear findings. Configuration and
rule-folder changes reload the engine and reanalyze the scope. Rapid saves
cancel/coalesce work; stale generations cannot replace the latest report.

Overview shows metric panels and real counts/status; Violations includes execution errors and
opens findings in Content with a click or focused `Enter`. Use `Tab` to focus
the main panel and arrows to select a finding. The violation table and selected
detail panel identify the rule,
folder-based rule set, YAML/JavaScript source, and exact file location. Rules lists
active configuration. Live activity remains in the sidebar rather than a
separate Watch page.
While analysis runs, previous results are explicitly labeled outdated.
Missing or invalid configuration does not stop browsing/watching, but is not a
clean analysis result.

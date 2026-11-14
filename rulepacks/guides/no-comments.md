# No ordinary comments

[Catalog](../README.md)

`no-comments.yml` is available in Go, JavaScript, TypeScript, Python, and Rust
folders. Copy the chosen YAML to `.guard/<language>/no-comments.yml` and
[the function](../functions/source-policy/no-comments.js) to
`.guard/functions/source-policy/no-comments.js`. The Go rule is also installed
in this repository.

The AST-backed function reports ordinary line/block comments at their start
locations. Strings containing comment-like text are not flagged.
Exceptions are intentionally explicit:

- Go: adjacent documentation groups for uppercase-named declarations, package
  documentation, cgo preambles immediately preceding `import "C"`, `go:`,
  legacy build tags, `line`, `export`, and `nolint` directives.
- JavaScript/TypeScript: JSDoc and recognized TypeScript, JSX, Flow, ESLint,
  Prettier, coverage, and pure-expression directives.
- Python: docstrings (string syntax, not comments), shebangs, encoding declarations,
  and recognized typing, linting, formatting, and coverage directives.
- Rust: inner/outer documentation comments. Rust attributes are syntax, not comments.

These are spelling/adjacency exceptions, not proof that documentation is useful
or a tool directive is valid. Go documentation groups before exported declarations
are allowed without assessing prose quality; JSDoc is allowed regardless of its
attachment. Unknown directives and license-header comments are not automatically
exempt. Customize the function before adopting it if required notices or other
tooling comments must be preserved; never remove legally required notices merely
to satisfy a rule.

Run `go test ./internal/analysis -run '^TestCommunityCommentPolicy$'`.
Synthetic fixtures cover every language, harmless strings, and exceptions.
Requires syntax API v1; maintained by Guardrail contributors under
[Apache License 2.0](../../LICENSE). Review functions before execution.

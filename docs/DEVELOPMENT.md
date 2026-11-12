# Development

[Documentation index](README.md) | [Contributing](CONTRIBUTING.md)

## Setup

Use Go 1.26.2 or newer. Clone your fork as described in the contribution guide.
Tree-sitter requires a C compiler and CGO enabled.
No Node.js frontend, database, or Kubernetes cluster is required for the current CLI.

```sh
go build -o bin/guardrail ./cmd/guardrail
go run ./cmd/guardrail watch
```

For a separate test context, pass an explicit path rather than generating noisy
fixtures in the repository:

```sh
go run ./cmd/guardrail watch /path/to/test-project
```

## Common commands

| Command | Purpose |
| --- | --- |
| `make build` | Build `bin/guardrail` |
| `make run` | Build and watch the current directory |
| `make test` | Run all tests with race detection and coverage |
| `go test ./internal/tui/screens` | Run a focused package suite |
| `go vet ./...` | Run Go's static checks |
| `make lint` | Run `golangci-lint`; requires it to be installed separately |
| `git diff --check` | Check patch whitespace |

Use `gofmt` on changed Go files. Before submitting code:

```sh
make test
go vet ./...
make build
git diff --check
```

CI pins golangci-lint v2.6.2. Install that version to reproduce `make lint`.
Pull-request quality checks and semantic GitHub releases are described in
[Releasing](RELEASING.md). Node.js 22 is used only for CI versioning scripts;
it is not required by the Guardrail runtime.
Documentation-only changes do not need the Go test suite unless code behavior or
testable examples also change; check Markdown links and instructions instead.

## Coding conventions

- Keep every source and test file at or below **300 lines**, counting imports and
  blank lines. Split by cohesive responsibility.
- Treat that limit as a contribution convention, not a runtime restriction on
  users' watched files.
- Follow existing names, formatting, helpers, and package boundaries.
- Keep changes focused; add regression tests for fixes and behavior changes.
- Keep comments concise and reserve them for non-obvious behavior.
- Return or display errors explicitly; do not silently replace failures with
  empty successful results.
- Keep filesystem work and expensive highlighting out of the rendering path.
- Preserve keyboard/mouse parity, light/dark styling, and bounded rendering.

The root [AGENTS.md](../AGENTS.md) records conventions discoverable by coding agents.
No-comment and function-length policies discussed in the product roadmap are
example future user rules, not additional enforced repository policies today.

## Testing

### Automated coverage

- Watcher tests cover recursive lifecycle changes, atomic saves, symlinks,
  cancellation, root removal, and indexed reconciliation.
- Screen tests cover navigation, layout, markers, preview safety, highlighting,
  cursor tracking, scrollbars, themes, and bounded rendering.
- Worker tests cover coalescing, outdated results, cancellation, and visible errors.
- Analysis tests cover the five language grammars, strict configuration, regex
  positions, JavaScript contracts/timeouts, script containment, and live reloads.
- CLI tests cover structured reports and violation/error exit semantics.
- The macOS terminal regression starts a pseudo-terminal through
  `/usr/bin/script`, opens a large context, checks theme output, and verifies quit.
  This platform-specific test does not run on other operating systems.

Keep unit tests deterministic and prefer logical bounds to machine-dependent
timing assertions. Existing timeout tests are safety limits, not performance SLAs.

### Manual terminal checks

For terminal or event-loop changes, exercise a real interactive session:

1. Launch on a small context and on a large disposable context.
2. Check Dashboard, Watch, and Content via both clicks and shortcuts.
3. Expand/collapse folders; test Home/End, wheel bursts, track clicks, and thumb drags.
4. Save, atomically replace, rename, and delete files while scrolling.
5. Check preview refresh, activity expiry, and idle behavior.
6. Resize through wide, normal, and narrow dimensions.
7. Restart in light and dark terminal themes.
8. Quit and verify terminal restoration; clean up only your own fixtures/processes.

ANSI diff output can split visible text across writes. Use rendered-frame
assertions or reconstruct terminal state instead of assuming every label appears
as one contiguous raw output string.

## Building screens

The screen is split by responsibility:

| File | Responsibility |
| --- | --- |
| [main.go](../internal/tui/screens/main.go) | Screen state and update application |
| [hierarchy.go](../internal/tui/screens/hierarchy.go) | Hierarchy indexing and flattening |
| [render.go](../internal/tui/screens/render.go) | Tree and workspace composition |
| [panel.go](../internal/tui/screens/panel.go) | Dashboard/Watch/Content panel composition |
| [navigation.go](../internal/tui/screens/navigation.go) | Keyboard and mouse navigation |
| [scrollbar.go](../internal/tui/screens/scrollbar.go) | Track clicks and captured thumb dragging |
| [content.go](../internal/tui/screens/content.go) | Preview loading, rendering, cursor following |
| [highlight.go](../internal/tui/screens/highlight.go) | Bounded reads, text safety, syntax highlighting |
| [preview_worker.go](../internal/tui/screens/preview_worker.go) | Background work and result validation |

Use containers to arrange children and sections to group content:

```go
theme := styles.Default()
section := components.NewSection(theme, "Diagnostics",
    tui.New(tui.WithText("No diagnostics yet.")),
)
content := components.NewContainer(theme,
    tui.WithFlexGrow(1),
    tui.WithGap(1),
)
content.AddChild(section)
```

This fragment assumes the corresponding project `styles`, `components`, and
go-tui imports. It demonstrates composition, not an existing rule-analysis UI.

Containers accept go-tui options and propagate theme styles unless overridden.
`NewLayout` accepts header, sidebar, content, and optional footer elements.
Generic sections have no padding or borders; the screen applies panel padding.
The current navbar is four rows and the footer three, with no outer margins.
The sidebar is at most 38 columns; the content fills the remaining width.

Keep palettes in [theme.go](../internal/tui/styles/theme.go). Pass the selected
theme into workers as well as UI components so syntax colors remain consistent.

## Performance changes

See [Performance and limits](PERFORMANCE.md) for benchmark commands and baseline
measurements. Include workload size and before/after results with optimization PRs.
Do not infer end-to-end responsiveness from an isolated metadata benchmark.

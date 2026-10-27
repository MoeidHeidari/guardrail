# Guardrail

A terminal workspace for repository engineering standards.

The current implementation is a TUI foundation: a header, sidebar, and empty
content area. `guardrail watch` opens this workspace only. File watching, rule
evaluation, configuration, and analysis are not implemented yet.

## Run

Requires Go 1.26.2 or newer and an interactive terminal.

```sh
make build
./bin/guardrail watch
```

Or run directly:

```sh
go run ./cmd/guardrail watch
```

Press `q` or `Ctrl+C` to quit. The layout resizes with the terminal.

## Structure

```text
guardrail/
|-- cmd/
|   `-- guardrail/
|       `-- main.go
|-- internal/
|   |-- cli/
|   |   |-- root.go
|   |   `-- watch.go
|   `-- tui/
|       |-- app/
|       |   `-- app.go
|       |-- components/
|       |   |-- container.go
|       |   |-- header.go
|       |   |-- layout.go
|       |   |-- section.go
|       |   `-- sidebar.go
|       |-- screens/
|       |   `-- main.go
|       `-- styles/
|           `-- theme.go
|-- Makefile
|-- go.mod
`-- README.md
```

Only implemented packages have directories. Domain, application, engine,
analysis, and adapters packages can be added when their functionality exists.
The same applies to schemas and examples.

## Building screens

- `cli` defines commands and delegates terminal startup to `tui/app`.
- `tui/app` owns terminal startup, event dispatch, and cleanup.
- `tui/screens` composes pages and defines screen-level key bindings.
- `tui/components` provides composable containers, borderless sections, headers,
  sidebars, and the main layout using go-tui elements.
- `tui/styles/theme.go` defines all application colors. Pass the same theme
  through the screen and its components.

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

Containers accept go-tui options, so rows, columns, spacing, sizing, and scrolling
can be composed without adding wrapper types. Child text inherits the container
theme unless explicitly styled. `NewLayout` accepts header, sidebar, and content
elements independently; replace their contents without changing app startup.
Sections use surface backgrounds and optional text headings without borders.
The layout uses a three-row header with vertically centered text, a one-row gap
below the header, and a one-column sidebar-content gap. Sections retain one-cell
internal padding, and the header retains two-column horizontal padding.
Font size follows your terminal settings.
The sidebar occupies one quarter of the body width; content fills the remaining
space through the header's right edge.

## Development

```sh
make build
make test
go vet ./...
```

`make run` builds and opens the workspace. `make lint` uses `golangci-lint` when
installed.

## License

See [LICENSE](./LICENSE).

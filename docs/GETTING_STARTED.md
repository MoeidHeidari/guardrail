# Getting started

[Documentation index](README.md)

## Requirements

- Go 1.26.2 or newer, as specified by [go.mod](../go.mod).
- Git to clone the repository.
- An interactive terminal. Mouse interaction requires terminal mouse reporting;
  keyboard navigation is also available.
- Read access to the chosen context and enough OS filesystem-watch resources.

The implementation has been exercised on macOS, including large-context terminal
workflows. Linux and Windows need platform-specific validation before equivalent
support can be claimed.

## Build from source

```sh
git clone https://github.com/MoeidHeidari/guardrail.git
cd guardrail
go build -o bin/guardrail ./cmd/guardrail
./bin/guardrail watch
```

If Make is installed, `make build` produces the same binary.
The documented installation route is a source build; these instructions do not
assume a published release binary or package-manager distribution.

For development, run directly from the repository:

```sh
go run ./cmd/guardrail watch
```

The executable path in these examples is relative to the cloned Guardrail
repository. From elsewhere, use its absolute path or a binary on your `PATH`.

## Choose a context

The context is the directory Guardrail monitors recursively.

```sh
# Watch the working directory.
./bin/guardrail watch

# Watch a relative or absolute directory.
./bin/guardrail watch ../my-project
./bin/guardrail watch /path/to/my-project

# Equivalent flag forms.
./bin/guardrail watch --context ../my-project
./bin/guardrail watch -c ../my-project
```

The default is the directory where the process starts, not the binary's directory
or an automatically discovered Git root. The path must exist and be a directory.
Use either the positional argument or the flag; combining them is an error.
A symlink supplied as the root is resolved to its target.

```sh
./bin/guardrail --help
./bin/guardrail watch --help
```

## First session

1. Start `watch` in the project you want to monitor.
2. Dashboard opens with a workspace summary.
3. The context root is expanded; its immediate files and folders are visible.
4. Expand a child folder with Right, `Enter`, or a click.
5. Select a file and press `Enter` to open Content.
6. Edit and save that file in your editor. Guardrail shows its activity and
   refreshes the preview.
7. Press `1` to inspect recent events, `0` to return to Dashboard, or `q` to quit.

Guardrail previews files but does not edit them. You do not need an AI coding agent
to use watching, and filesystem activity does not identify who authored an edit.

## Configuration status

There is no rule configuration loader yet. A `.guardrail.yml` file currently has
no effect. The proposed standards engine and configuration work are described in
[Architecture and roadmap](ARCHITECTURE.md).

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Wrong directory is displayed | Pass an explicit context; the Git root is not inferred. |
| Mouse clicks do not work | Check terminal mouse-reporting support; use keyboard controls instead. |
| Theme does not match | Restart after changing the terminal theme. Detection falls back to `COLORFGBG`, then dark mode. |
| Empty space outside the entire TUI | The terminal emulator may add window padding outside its character grid; Guardrail cannot draw there. |
| Preview is unavailable | Check the message in Content: the file may be binary, a symlink, too large, or unreadable. |
| Startup is slow | Initial discovery visits the whole context, including dependency folders and Git internals. |
| Watch setup fails | Inspect the reported permission, watch-limit, or open-file-limit error. Try a narrower context. |
| Context disappears | Removing or renaming the root ends the session with an error; restart on a valid path. |

The terminal normally restores on exit. If a crash leaves your shell in an unusable
display/input state, run your terminal's reset command (commonly `reset`).

See [Performance and limits](PERFORMANCE.md) for resource behavior and
[Support](SUPPORT.md) for reporting a reproducible problem.

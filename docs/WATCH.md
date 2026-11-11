# Watch workspace

[Documentation index](README.md) | [Getting started](GETTING_STARTED.md)

## Tabs

### Dashboard

Dashboard is the default page and is always available with `0` or a tab click.
It currently contains a workspace container showing the context path, entry
count, and changed-file count. It is the intended home for future engineering
status and rule findings; those features are not implemented yet.

### Watch

Press `1` to see the latest 200 changes, newest first, with timestamps and
context-relative paths. Green indicates creation, yellow modification, and red
deletion.

At the top, new events follow automatically. When you scroll back, your position
is retained until the relevant events leave the bounded history.

### Content

Press `2` to use the read-only preview. Clicking a file or pressing `Enter` on it
also opens Content. While Content is active, changing the tree selection previews
the newly selected file.

The breadcrumb shows the file, detected language, and size. Chroma selects syntax
highlighting from the filename, including Go, JavaScript/TypeScript, Python,
Rust, JSON, YAML, shell, and Dockerfiles. Unknown names render as plain text.

Watching continues regardless of the active tab.

## Context hierarchy

- Directories sort before files.
- The context root starts expanded one level; nested and newly discovered folders
  start collapsed.
- Explicit expansion/collapse choices survive filesystem updates.
- Collapsed folders and their descendants remain watched.
- Hidden and Git-ignored entries are included. There are no automatic exclusions
  for dependencies, generated output, or Git internals.
- Symlink entries are shown without recursively following their targets.

## Change markers and activity

| Marker | Meaning |
| --- | --- |
| `[A]` | Added during the session |
| `[M]` | Modified or changed during the session |
| `[*]` | An ancestor folder contains session changes |

Markers are independent of Git status and are not persisted between sessions.
Deleted entries leave the tree, but their deletion is recorded in Watch and
surviving ancestors remain marked.

Recently changed files have purple activity styling and dancing dots. Dots
advance every 200 ms for two seconds after the latest activity. The file's
highlight and marker remain after animation stops. Folder markers do not animate.

## Preview behavior

- File previews refresh after saves and support vertical and horizontal scrolling.
- Supported content is UTF-8 regular files up to 1 MiB and 20,000 lines.
- Binary files, symlinks, oversized files, and read errors show an explicit message.
- Tabs display as four spaces, CRLF endings are normalized, and other control
  characters are escaped rather than interpreted as terminal commands.

### Live change cursor

When previewed text changes, Content follows the latest changed region. A `>`
gutter marker identifies its line and a purple block marks the inferred edit
boundary. Deletions point at the surviving boundary.

This is computed from successive saved file snapshots, **not the editor's actual
cursor**. Unsaved changes are not visible. Multiple edits between snapshots are
treated as one region. You can scroll manually until the next text change.

## Keyboard and mouse

| Input | Action |
| --- | --- |
| `0` / `1` / `2`, or click a tab | Open Dashboard / Watch / Content |
| `Tab` | Switch focus between the context tree and the active tab |
| Up/Down or `k`/`j` | Move tree selection or scroll Watch/Content |
| Right or `l` | Expand a folder/select its first child; in Content, scroll right |
| Left or `h` | Collapse a folder/select its parent; in Content, scroll left |
| `Enter` / `Space` | Toggle the selected folder or open a file |
| `Home` / `End` | Jump to first/last tree entry, retained event, or content line |
| `PageUp` / `PageDown` | Move by a page in the active pane |
| Left click | Focus a pane, select a tree entry, toggle a folder, or open a file |
| Mouse wheel | Scroll the pane under the pointer |
| Scrollbar track click | Move one page |
| Scrollbar thumb drag | Scroll; dragging remains captured outside the track until release |
| `r` | Reload the selected preview, including retrying a read error |
| `q` / `Ctrl+C` | Quit and restore the terminal |

Dashboard currently has no scrollable content.

## Appearance

The theme is detected once at startup. Both UI colors and syntax highlighting
match the detected light/dark background. Unsupported background queries fall
back to `COLORFGBG`, then dark mode. Restart after changing terminal themes;
Guardrail does not poll for theme changes.

The workspace has a four-row navbar, a context sidebar up to 38 columns wide,
a main panel filling remaining width, and a three-row status/help footer.
Underlined tabs and accent headings identify the active page and keyboard focus.
The sidebar shrinks on narrow terminals and hints shorten as needed.

There are no outer layout margins. Backgrounds fill the terminal character grid.
Internal padding is two columns horizontally and at least one row vertically.
A one-column divider separates the sidebar and main panel. Font size and any
padding outside the character grid are controlled by your terminal application.

## Event semantics and failures

Creates, modifications, permission changes, and deletions update the workspace.
Renames appear as removal of old paths and creation of new paths. Newly created
or moved-in directory subtrees are discovered recursively.

Events are coalesced over 20 ms and reconciled with current disk state. This is a
live view, not a lossless audit log of individual writes or short-lived files.

Setup, permission, and runtime failures are reported explicitly. Removing or
renaming the context root stops watching with an error. Event overflow also ends
the session rather than silently continuing with an incomplete view.

See [Performance and limits](PERFORMANCE.md) for implementation details.

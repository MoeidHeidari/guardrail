# Watch workspace

[Documentation index](README.md) | [Getting started](GETTING_STARTED.md)

## Tabs

### Workspace pages

The sidebar uses thin horizontal divider lines between workspace navigation,
the independently scrollable Context tree, and live Watch activity below it.
Sections share the same background, without tinted cards or boxed borders.
Workspace navigation, Context rows, and Watch use compact single-line items with
horizontal text insets. Selection uses a restrained tinted background and a thin
blue left accent, without tall boxes or extra blank rows. The accent column is
reserved for every row so changing selection does not shift the text.
Navigation contains Overview, Violations,
and Rules. Choose a page with a click, `o` / `v` / `u`, or focus navigation with
`Tab` and use Up/Down then `Enter`. Page selection preserves tree state.

Overview is the default workspace page. Workspace navigation receives keyboard
focus at startup, rather than Context. Metrics at the top show violations,
execution errors, affected files, and rules, with workspace status and findings below.
The content preview stays hidden until you select a finding with a click,
Up/Down in the focused list, or Enter. It appears beside the findings on wide
terminals and below them on narrower terminals.
Rule sets appear as highlighted `[set-name]` tags beside affected-file locations.
The findings panel scrolls horizontally with Left/Right (or `h`/`l`) when focused,
so long paths, tags, and rule names remain accessible.
Filter Overview's affected findings with All / Violations / Errors or `a` / `b` / `e`,
just like the violation table. Selection and live preview follow matching results;
metrics remain totals for the entire report.
The preview follows the selected location without leaving Overview. Saved edits
refresh that preview. Its mouse wheel and scrollbar scroll content independently.
It shows the context path, entry count, changed-file count, and analysis status.
Violations shows policy violations and execution errors; Rules lists active
regex/JavaScript rules. Missing or invalid configuration is explicitly reported.
Rules groups entries by rule set, with live impact summaries for each set and
rule: unique affected files, violation occurrences and violating files, and
execution errors and files with errors.
The Rules workspace uses a compact tree on the left: rule sets are parents
and their rules are indented children. Only the selected row has an accent
border, with no background highlight. Left/Right collapses/expands sets.
The tree heading shows the total number of rule sets and rules, even when sets
are collapsed. Matching file findings appear on the right. The full-width content
preview stays hidden until you click a file finding, navigate the file list,
or press Enter there. Once opened, it gets more vertical space than the compact,
scrollable lists above.
Selecting a set includes its rules; selecting a rule narrows the file list.
Use `Tab` between the rule tree, files, and the preview when visible; arrows navigate or scroll, and clicks
select directly. File selection updates the preview without opening Content.
Narrow terminals stack the two lists above the preview. A file with several findings counts once
in each file total; set totals deduplicate across their rules. Pending analysis
is labeled outdated rather than implying the previous counts are current. Missing or invalid configuration is explicitly reported.
No policy is built in. See [Rules and checking](RULES.md).
The violation table has fixed column headings for kind, file/location, rule set,
rule, and message.
Focus the table and use Left/Right (`h`/`l`) to scroll full-width columns; headings
stay aligned with the rows. Filter with the clickable All / Violations / Errors
buttons or `a` / `b` / `e`. Up/Down and Enter operate only on matching findings.
Summary metrics continue to describe the entire report, not just the filter. A subtle highlighted row marks selection; the detail panel
below shows the selected location, message and YAML/JavaScript source without
repeating a border around every result. Headings sit left-aligned inside panels,
not in the middle of their top borders. Detail layout compacts in shorter terminals. Root-level YAML rules belong to General; nested rule sets use their containing
folder's name. Parser and infrastructure errors belong to System.

Before opening a file, Overview, Violations, and Rules show a plain page heading
without a tab underline or `0` label. Workspace pages do not repeat the heading
in a subtitle; Content retains its file breadcrumb.
Once Content is available, the heading becomes the first main tab and follows
the selected workspace page. Press `0` to return to
that page after using Content. Watch is a scrollable sidebar section below Context,
not a page or main tab. Content appears only after opening a file for the first time and
then remains available as a main tab. Watching continues throughout.

### Watch

Watch is excluded from Tab and workspace-menu navigation; it remains visible
as a passive activity section. Mouse scrolling and the explicit `1` shortcut
remain available.
Watch below Context shows the latest 200 changes, newest first, with timestamps
and context-relative paths. `A` indicates creation, `M` modification,
and `D` deletion, in green, yellow, and red respectively. Click or press `1` to
focus activity without replacing the main workspace page or Content. It supports
arrow/Page/Home/End keys, mouse wheel scrolling and scrollbar dragging.

At the top, new events follow automatically. When you scroll back, your position
is retained until the relevant events leave the bounded history.

### Content

Open a file with a click or `Enter` to reveal the read-only Content tab.
Afterward, press `2` to return to it. While Content is active, changing the tree selection previews
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
- Hidden and Git-ignored entries are included unless `.guardignore` excludes
  them. Git's `.gitignore` files are not read automatically.
- Symlink entries are shown without recursively following their targets.

## Ignore paths

Put `.guardignore` in the context root. It uses Git-style patterns:

```gitignore
# Ignore Git metadata and build output.
.git/
/bin/
*.log
!important.log
```

`*`, `?`, character classes, `**`, leading `/` root anchoring, trailing `/`
directory-only matches, comments, escaped `#`/`!`, and `!` negation are supported.
Last matching pattern wins. As in Git, you cannot re-include a file while its
parent directory is ignored; unignore the parent first.

Ignored entries are omitted from the context tree, activity feed, and `check`.
Ignored directories are pruned before watches are allocated. Git's `.gitignore`
and nested ignore files are not loaded. The root and `.guardignore` itself remain
observable so edits, replacement, creation, or deletion of the ignore file
trigger a full rescan and watch reconciliation. Paths becoming hidden appear
as removals; paths becoming visible appear as additions. Keep `.guard/` visible
if you want changes to rule sets/scripts to trigger automatic reanalysis.
Unreadable or oversized ignore files (over 1 MiB) are errors.

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
| `0` / `1` / `2` | Open workspace page / focus sidebar activity / Content (after first file open) |
| `o` / `v` / `u` | Open Overview / Violations / Rules |
| `Tab` | Cycle focus: context tree, main panel, workspace navigation |
| Up/Down or `k`/`j` | Move tree/finding selection or scroll Watch/Content/Rules |
| Right or `l` | Expand a folder/select its first child; in Content, scroll right |
| Left or `h` | Collapse a folder/select its parent; in Content, scroll left |
| `Enter` / `Space` | Toggle the selected folder, open a file, or open a focused finding |
| `Home` / `End` | Jump to first/last tree entry, retained event, or content line |
| `PageUp` / `PageDown` | Move by a page in the active pane |
| Left click | Focus a pane, select a tree entry, toggle a folder, or open a file |
| Mouse wheel | Scroll the pane under the pointer |
| Scrollbar track click | Move one page |
| Scrollbar thumb drag | Scroll; dragging remains captured outside the track until release |
| `r` | Reload the selected preview, including retrying a read error |
| `q` / `Ctrl+C` | Quit and restore the terminal |

Violations and Rules scroll independently of the tree and Watch/Content.
Click a finding, or focus the main panel and press `Enter`, to open its location.

## Appearance

The theme is detected once at startup. Both UI colors and syntax highlighting
match the detected light/dark background. Unsupported background queries fall
back to `COLORFGBG`, then dark mode. Restart after changing terminal themes;
Guardrail does not poll for theme changes.

The workspace uses a dark navy palette (or a matching light palette), thin blue-toned
panel outlines, and a padded navbar with the brand, workspace name, and watching
status aligned on a single text row.
It has a context sidebar up to 38 columns wide,
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

# Watch workspace

[Documentation index](README.md) | [Getting started](GETTING_STARTED.md)

## Tabs

### Workspace pages

The sidebar uses thin horizontal divider lines between workspace navigation,
the independently scrollable Context tree, and live Watch activity below it.
Sections share the same background, without tinted cards or boxed borders.
Workspace navigation, Context rows, and Watch use compact single-line items with
horizontal text insets. Selection uses a restrained tinted background and readable
text, without vertical markers, boxes, or extra blank rows.
Use `[` / `]` when Context or Watch is focused, or Shift+mouse wheel over either
section, to scroll long names and paths horizontally.
Left/Right navigate without changing horizontal scroll offsets. Use `[` / `]`
for horizontal scrolling, including in file and search previews.
Navigation contains Overview, Violations,
and Rules. Choose a page with a click, or focus navigation with
`Tab`, use Up/Down, then `Enter` or Right (which also moves focus into the page). Page selection preserves tree state.

Overview is the default workspace page. Workspace navigation receives keyboard
focus at startup, rather than Context. Metrics at the top show violations,
execution errors, affected files, and rules, with workspace status and findings below.
Overview and Violations share metric cards with bold counts, colored bars with
muted empty tracks, and labeled ratios. Violations and errors show their share of
all findings; affected files are compared with checked files. The Rules count is
the configured total, with its bar showing how many rules have findings.
Cards distinguish unavailable analysis (`--`) from zero and mark pending results
as updating. Bars describe result proportions, not analysis completion progress.
The content preview stays hidden until you select a finding with a click,
Up/Down in the focused list, or Enter. It appears beside the findings on wide
terminals and below them on narrower terminals.
Rule sets appear as highlighted `[set-name]` tags beside affected-file locations.
The findings panel scrolls horizontally with `[` / `]` when focused,
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
rule: unique affected files, violation occurrences, and execution errors.
The Rules workspace uses a grouped table on the left: rule sets are parents
and their rules are indented children, with columns for type, rule count, affected
files, violations, and errors. Only the selected row has an accent
border, with no background highlight. Set names use the accent color; nonzero
violation and error counts use the warning and error colors. Space collapses/expands a set; Right moves to the next pane, Left to the previous one and finally back to the sidebar.
The table heading shows the total number of rule sets and rules, even when sets
are collapsed. Use `[` / `]` to scroll its columns or the file list horizontally.
Matching file findings appear on the right. The full-width content
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
Focus the table and use `[` / `]` to scroll full-width columns; headings
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
not a page or main tab. Content appears after opening a file and remains available
as a main tab until closed. Watching continues throughout.

### Watch

Watch participates in the Tab cycle: Workspace menu, Context, main pane, Watch.
Down from the last Context row enters Watch; Up from its first row returns to
Context. Mouse interaction and the explicit `w` shortcut remain available.
Watch below Context shows the latest 200 changes, newest first, with timestamps
and context-relative paths. `A` indicates creation, `M` modification,
and `D` deletion, in green, yellow, and red respectively. Click or press `w` to
focus activity without replacing the main workspace page or Content.
Up/Down selects changed files; Enter or Right opens the selection in a content
tab. Click a change to open it. Up from the first change returns focus to Context.
Deleted files show an unavailable-file message. Page/Home/End keys, mouse wheel
scrolling and scrollbar dragging remain available; use `[` / `]` for horizontal scrolling.

The focused pane's heading uses a subtle background tint and bold contrasting text.
Inactive headings retain their quieter colors; focus adds no borders.

At the top, new events follow automatically. When you scroll back, your position
is retained until the relevant events leave the bounded history.

### Content

Open a file with a click or `Enter` to add a Content tab named after it. Each
opened file (or finding location) gets its own tab; opening an already open file
reuses its tab. Click a tab, press `1`-`9` to jump to that tab (`0` is the workspace tab), `<` / `>` to
cycle, and `x` or a tab's `[x]` to close it. When tabs do not fit, the bar scrolls to
keep the active tab visible and shows `‹N` / `N›` for hidden tabs (click to step). Each tab keeps its own scroll position, and
tree selection no longer replaces the active file. Duplicate file names show their
paths. Closing the last tab returns to the workspace page.

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
- Click `[x]` or press `x` to close the active preview; select a file or finding to reopen it.
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

The header's **Search files [/]** searches filenames and context-relative
paths (case-insensitive). Press `/` or click it, then type to filter watched
files. Up/Down selects results; Enter, Right, or a result click previews the file
in the right-hand pane without creating a tab, focuses the preview, and reveals
the file in Context. Up/Down scrolls the focused preview; Left or Esc returns to
the results. Esc from results cancels without changing the current page or file.
Search updates as files
are added or removed and respects the watched context's ignore rules.

The footer shows shortcuts for the focused pane, including filters, Rules folding,
and content tab navigation. It prioritizes the relevant actions and fits the
available terminal width; wider terminals also show sidebar jump shortcuts.

| Input | Action |
| --- | --- |
| `/` | Focus global file search in the header |
| `0` / `1`-`9` | Workspace tab / content tab N |
| `<` / `>` | Previous / next tab (wraps through the workspace tab) |
| `o` / `v` / `u` | Open Overview / Violations / Rules and focus it |
| `c` / `w` | Focus Context / Watch |
| `Tab` | Cycle focus: workspace navigation, Context, main panel, Watch |
| `Esc` | Return focus from the page or Watch to the sidebar item that opened it |
| Up/Down or `k`/`j` | Move tree/finding selection or scroll Watch/Content/Rules |
| Right or `l` | Folder: expand/select first child. Workspace item or file: open it and focus the page. In Overview/Violations: next filter. In Rules: next pane |
| Left or `h` | Collapse a folder/select its parent. In Overview/Violations: previous filter, or back to the sidebar from "All". In other panes: return to the previous pane or sidebar |
| Down / Up at an edge | Down from the last Workspace item enters Context; Up from the first Context row enters the Workspace menu |
| `[` / `]` | Scroll the focused Context, Watch, table, findings, or preview pane horizontally |
| `x` | Close the active content preview |
| `Enter` / `Space` | Toggle the selected folder, open a file, or open a focused finding |
| `Home` / `End` | Jump to first/last tree entry, retained event, or content line |
| `PageUp` / `PageDown` | Move by a page in the active pane |
| Left click | Focus a pane, select a tree entry, toggle a folder, or open a file |
| Mouse wheel | Scroll the pane under the pointer |
| Shift+mouse wheel | Scroll Context or Watch horizontally |
| Scrollbar track click | Move one page |
| Scrollbar thumb drag | Scroll; dragging remains captured outside the track until release |
| `r` | Reload the selected preview, including retrying a read error |
| `q` / `Ctrl+C` | Quit and restore the terminal |

Violations and Rules scroll independently of the tree and Watch/Content.
Click a finding, or focus the main panel and press `Enter`, to open its location.

## Appearance

The workspace inherits the terminal's default background and foreground, and uses
its configurable ANSI palette for accents, errors, warnings, and activity.
Light/dark detection at startup selects syntax highlighting and subtle selection
tints, which remain application-defined. Highlighted text uses a matching explicit
foreground so terminal accent colors cannot disappear against the selection tint.
Unsupported background queries fall
back to `COLORFGBG`, then dark mode. Restart after changing terminal themes;
Guardrail does not poll for theme changes.

Exact accent colors and their contrast depend on your terminal palette; Guardrail
does not read or replace the entire terminal theme.
The workspace uses subtle panel outlines and a padded navbar with the brand, workspace name, and watching
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

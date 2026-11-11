# Performance and limits

[Documentation index](README.md) | [Watch workspace](WATCH.md)

## Event-driven monitoring

Guardrail uses fsnotify notifications, not periodic full-directory polling.
An initial recursive discovery builds a metadata index and installs filesystem
watches. That startup cost still grows with the size of the context.

After startup:

1. Notifications are coalesced over a fixed 20 ms window.
2. An indexed path/children map reconciles touched entries.
3. New or replaced directory subtrees are discovered recursively; deleted
   subtrees are removed from the index.
4. Metadata deltas reach the UI rather than complete copies of the context.
5. Ordinary writes update indexed entries and visible rows directly.

Structural updates can still flatten the visible hierarchy. Each affected sibling
list is filtered once and sorted once per batch, rather than rescanned for every
changed path. Row storage is reused.

## Rendering and previews

- Tree and preview elements are virtualized: a bounded row window and spacers
  preserve the logical scroll range without an element for every file or line.
- Input and updates share a pending frame at most once per 16 ms. Wheel movement
  accumulates before rendering instead of resetting against the last frame.
- Frame timers stop when idle. Noninteractive mouse motion does not redraw.
- Animation updates existing elements and stops when recent activity expires.
  Hidden activity does not cause animation-only redraws.
- A bounded eight-update channel absorbs short UI stalls without dropping
  application-level updates. Sustained overload can still create backpressure.
- One preview worker reads, compares, and highlights files outside the UI loop.
  Requests coalesce for 40 ms, with one running and one latest pending request.
  Stale results cannot replace a newer request or another selected file.

The worker re-highlights the selected file rather than using an incremental lexer.
A running highlighting operation finishes before worker shutdown; newer requests
replace pending work, not the running operation.

## Bounds

| Resource | Current behavior |
| --- | --- |
| Watch history | Latest 200 changes |
| Text preview | UTF-8 regular files, maximum 1 MiB and 20,000 lines |
| Preview work | One running request and one latest pending request |
| Metadata and hierarchy indexes | Grow with tracked entries |
| Filesystem resources | Depend on OS backend and context size |
| Initial scan | Visits the entire context |

Syntax spans and text for the selected preview are cached within the preview
limits. File contents are not cached for the whole repository.

There are currently **no automatic exclusions** for dependency directories,
generated output, hidden files, or Git internals. Git ignore rules are not
analysis or watch filters. Use a narrower context when appropriate.

## Platform limits

On macOS, kqueue monitoring opens descriptors for watched files as well as
directories. Terminal input resources are initialized before watches so a large
context cannot push newly allocated terminal-reader descriptors beyond the
select-based reader's range. OS open-file limits still apply.

Other backends also impose watch/queue limits. Watch-limit, permission, and queue
overflow errors are surfaced; there is no automatic overflow rescan/recovery yet.
Removing or renaming the context root stops the session.

Guardrail does not promise constant resource usage for an arbitrarily large
repository or lossless delivery of every intermediate edit.

## Measurements

These are development measurements on an Apple M3 Pro, not latency guarantees or
cross-platform capacity claims:

| Focused workload | 1,000 indexed entries | 100,000 indexed entries |
| --- | --- | --- |
| Single-write watcher reconciliation | About 3.1 us, 712 B | About 3.0 us, 712 B |
| Previous full-copy/diff implementation | About 107 us, 222 KB | About 16 ms, 16 MB |
| UI metadata delta application | About 0.36 us, zero allocations | About 0.36 us, zero allocations |

A structural-batch microbenchmark at 10,000 entries improved from approximately
37.6 ms to 2.2 ms after replacing repeated sibling scans with batch filtering.
This trades additional temporary batch storage for lower processing time.

A terminal smoke test with 10,001 files and a burst of 100 wheel events observed
a live preview update in approximately 116 ms. That is one observed workflow,
not an automated performance threshold for all terminals and machines.

## Reproduce benchmarks

From the repository root:

```sh
go test ./internal/watcher ./internal/tui/screens -run '^$' \
  -bench 'Benchmark(SingleWriteBatch|MetadataDelta|ActivityFrame|StructuralBatch)$' \
  -benchmem
```

For meaningful comparisons, use the same Go version, machine, context shape,
terminal, and workload. Measure end-to-end rendering separately from isolated
index benchmarks.

If performance regresses, report context size, file/directory counts, edit pattern,
terminal, OS, and reproduction steps through the [support guide](SUPPORT.md).

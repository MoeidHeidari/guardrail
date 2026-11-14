package screens

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestMetadataDeltaPreservesTreeStorage(t *testing.T) {
	m := testScreen()
	before := &m.rows[0]
	entry := m.entries[2]
	entry.Size = 99
	m.Apply(watcher.Update{Upsert: []watcher.Entry{entry}, Changes: []watcher.Change{{Path: entry.Path, Kind: "modified"}}})
	if &m.rows[0] != before || m.rows[m.rowPositions[entry.Path]].entry.Size != 99 || m.changedFiles != 1 {
		t.Fatal("metadata delta must update the visible row without rebuilding the tree")
	}
	m.Apply(watcher.Update{Upsert: []watcher.Entry{entry}, Changes: []watcher.Change{{Path: entry.Path, Kind: "modified"}}})
	if m.changedFiles != 1 {
		t.Fatal("repeated modifications must not double-count changed files")
	}
	m.selected = m.rowPositions[entry.Path]
	m.collapse()
	m.expand()
	m.expand()
	if m.rows[m.selected].entry.Size != 99 {
		t.Fatal("cached hierarchy must use current metadata after collapse/expand")
	}
}

func TestStructuralDeltasMaintainHierarchy(t *testing.T) {
	m := testScreen()
	directory := filepath.Join("src", "added")
	file := filepath.Join(directory, "new.go")
	m.Apply(watcher.Update{
		Upsert:  []watcher.Entry{{Path: directory, Mode: os.ModeDir}, {Path: file}},
		Changes: []watcher.Change{{Path: directory, Kind: "created"}, {Path: file, Kind: "created"}},
	})
	if !m.collapsed[directory] {
		t.Fatal("new directories must be collapsed")
	}
	m.selected = m.rowPositions[directory]
	m.expand()
	if _, visible := m.rowPositions[file]; !visible {
		t.Fatal("new child must be reachable through the cached hierarchy")
	}
	m.selected = m.rowPositions[file]
	m.Apply(watcher.Update{
		Removed: []string{directory, file},
		Changes: []watcher.Change{{Path: directory, Kind: "deleted"}, {Path: file, Kind: "deleted"}},
	})
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("deleted selected subtree must select surviving parent")
	}
	if _, retained := m.treeChildren[directory]; retained {
		t.Fatal("deleted subtree must not remain in hierarchy cache")
	}
}

func TestTreeRenderingBoundedAndNavigable(t *testing.T) {
	m := testScreen()
	for i := range 100000 {
		m.entries = append(m.entries, watcher.Entry{Path: fmt.Sprintf("file%06d", i)})
	}
	m.rebuildRows()
	render(m, 120, 24)
	if len(m.tree.Children()) > 102 {
		t.Fatal("tree must render a bounded window, not all entries")
	}
	press(t, m, tui.KeyEnd)
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "file099999") {
		t.Fatalf("last entry must remain reachable:\n%s", frame)
	}
	rect := m.tree.ContentRect()
	_, offset := m.tree.ScrollOffset()
	expected := offset / treeRowHeight
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	if m.selected != expected {
		t.Fatalf("mouse selected %d, want %d; tree=%+v offset=%d", m.selected, expected, rect, m.treeOffset)
	}
}

func waitPreview(t *testing.T, results <-chan PreviewResult) PreviewResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("preview worker did not respond")
		return PreviewResult{}
	}
}

func TestPreviewWorkerCoalescesAndRejectsStaleResults(t *testing.T) {
	m := fileScreen(t, "package main\n")
	results, stop := m.StartPreviewWorker(context.Background())
	defer stop()
	m.contentTab = true
	render(m, 120, 24)
	if !strings.Contains(m.preview.message, "Loading") {
		t.Fatal("UI must display loading state without blocking on file IO/highlighting")
	}
	first := waitPreview(t, results)
	m.previewDirty = true
	render(m, 120, 24)
	if m.ApplyPreview(first) {
		t.Fatal("outdated worker result must not replace a newer request")
	}
	latest := waitPreview(t, results)
	if !m.ApplyPreview(latest) || !strings.Contains(render(m, 120, 24), "package main") {
		t.Fatal("latest preview must be applied on UI loop")
	}
	if err := os.WriteFile(filepath.Join(m.root, "main.go"), []byte("package main\n// live\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m.previewDirty = true
	render(m, 120, 24)
	if !m.ApplyPreview(waitPreview(t, results)) || m.preview.cursor == nil {
		t.Fatal("async updates must preserve live cursor tracking")
	}
	m.selected = 0
	if m.ApplyPreview(latest) {
		t.Fatal("selection changes must reject previous file results")
	}
}

func TestPreviewWorkerCancellation(t *testing.T) {
	m := fileScreen(t, "package main\n")
	_, stop := m.StartPreviewWorker(context.Background())
	m.contentTab = true
	render(m, 120, 24)
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("preview worker must stop while coalescing a pending request")
	}
}

func TestPreviewWorkerLatestRequestAndErrors(t *testing.T) {
	m := fileScreen(t, "package main\n")
	results, stop := m.StartPreviewWorker(context.Background())
	defer stop()
	m.contentTab = true
	for range 20 {
		m.previewDirty = true
		render(m, 120, 24)
		if len(m.previewRequests) > 1 {
			t.Fatal("pending previews must remain bounded")
		}
	}
	for {
		result := waitPreview(t, results)
		if result.id == m.previewID {
			if !m.ApplyPreview(result) {
				t.Fatal("latest coalesced request must be accepted")
			}
			break
		}
		if m.ApplyPreview(result) {
			t.Fatal("superseded preview must not be accepted")
		}
	}
	if err := os.Remove(filepath.Join(m.root, "main.go")); err != nil {
		t.Fatal(err)
	}
	m.previewDirty = true
	render(m, 120, 24)
	if !m.ApplyPreview(waitPreview(t, results)) || !strings.Contains(m.preview.message, "Cannot preview") {
		t.Fatal("background read errors must be visible rather than displaying stale content")
	}
}

func BenchmarkMetadataDelta(b *testing.B) {
	for _, count := range []int{1000, 100000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			entries := []watcher.Entry{{Path: ".", Mode: os.ModeDir}}
			for i := range count {
				entries = append(entries, watcher.Entry{Path: fmt.Sprintf("file%d", i)})
			}
			m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: entries})
			update := watcher.Update{Upsert: []watcher.Entry{{Path: "file0"}}, Changes: []watcher.Change{{Path: "file0", Kind: "modified"}}}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				m.Apply(update)
			}
		})
	}
}

package screens

import (
	"fmt"
	"os"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func BenchmarkStructuralBatch(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			entries := []watcher.Entry{{Path: ".", Mode: os.ModeDir}}
			for i := range count {
				entries = append(entries, watcher.Entry{Path: fmt.Sprintf("file%06d", i)})
			}
			m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: entries})
			update := watcher.Update{Removed: []string{"new"}, Upsert: append([]watcher.Entry{{Path: "new"}}, entries[1:]...)}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				m.Apply(update)
			}
		})
	}
}

func TestHierarchyMixedBatch(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{
		Removed: []string{"src/main.go", "README.md"},
		Upsert: []watcher.Entry{
			{Path: "src", Mode: os.ModeDir, Size: 99},
			{Path: "src/new.go"}, {Path: "README.md", Size: 12},
			{Path: "folder", Mode: os.ModeDir}, {Path: "folder/child"},
		},
	})
	for parent, children := range m.treeChildren {
		seen := make(map[string]bool)
		for _, path := range children {
			if seen[path] {
				t.Fatalf("duplicate child %s under %s", path, parent)
			}
			seen[path] = true
			if _, exists := m.entryPositions[path]; !exists {
				t.Fatalf("deleted child retained: %s", path)
			}
		}
	}
	if _, exists := m.rowPositions["src/main.go"]; exists {
		t.Fatal("removed file remains visible")
	}
	if m.rows[m.rowPositions["README.md"]].entry.Size != 12 {
		t.Fatal("replacement metadata missing")
	}
}

func TestWheelBurstAccumulatesBeforeRender(t *testing.T) {
	m := testScreen()
	for i := range 1000 {
		m.entries = append(m.entries, watcher.Entry{Path: fmt.Sprintf("file%06d", i)})
	}
	m.rebuildRows()
	render(m, 120, 32)
	rect := m.tree.ContentRect()
	for range 10 {
		m.HandleMouse(tui.MouseEvent{Button: tui.MouseWheelDown, X: rect.X, Y: rect.Y})
	}
	if m.treeOffset != 30 {
		t.Fatalf("wheel burst lost movement before rendering: offset=%d, want 30", m.treeOffset)
	}
}

package screens

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestContextStartsExpandedOneLevel(t *testing.T) {
	initial := testScreen().entries
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: initial})
	m.navigationFocus = false
	if len(m.rows) != 3 || m.collapsed["."] || !m.collapsed["src"] {
		t.Fatal("root must start expanded with child folders collapsed")
	}
	if _, visible := m.rowPositions["src/main.go"]; visible {
		t.Fatal("nested entries must not be visible initially")
	}
	m.Apply(watcher.Update{Entries: initial})
	if len(m.rows) != 3 || m.collapsed["."] {
		t.Fatal("updates must preserve initial first-level expansion")
	}
	m.toggle()
	if len(m.rows) != 1 {
		t.Fatal("root must remain manually collapsible")
	}
	m.Apply(watcher.Update{Entries: initial})
	if !m.collapsed["."] || len(m.rows) != 1 {
		t.Fatal("updates must preserve explicit collapse")
	}
	m.toggle()
	if len(m.rows) != 3 || !m.collapsed["src"] {
		t.Fatal("reopening root must leave child folders collapsed")
	}
}

func TestScrollbarTrackAndDrag(t *testing.T) {
	m := testScreen()
	for i := range 100 {
		m.entries = append(m.entries, watcher.Entry{Path: fmt.Sprintf("file%03d", i)})
	}
	m.rebuildRows()
	render(m, 120, 24)
	rect := m.tree.ContentRect()
	x := rect.X + rect.Width - 1
	if !m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: x, Y: rect.Bottom() - 1}) || m.treeOffset == 0 {
		t.Fatal("clicking scrollbar track must page, not select a file")
	}
	render(m, 120, 24)
	top, _, _, _ := scrollThumb(m.tree)
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: x, Y: rect.Y + top})
	if m.dragPane == nil {
		t.Fatal("thumb press must start drag")
	}
	render(m, 120, 24)
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, Action: tui.MouseDrag, X: x + 20, Y: rect.Bottom() + 10})
	render(m, 120, 24)
	_, maximum := m.tree.MaxScroll()
	if m.treeOffset != maximum {
		t.Fatalf("drag outside track must clamp to end: %d != %d", m.treeOffset, maximum)
	}
	m.HandleMouse(tui.MouseEvent{Action: tui.MouseRelease})
	if m.dragPane != nil {
		t.Fatal("release must stop dragging")
	}
	press(t, m, tui.KeyEnd)
	if m.selected != len(m.rows)-1 {
		t.Fatal("keyboard navigation must remain available after dragging")
	}
}

func TestActivityExpiresAndReusesTree(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "src/main.go", Kind: "modified"}}})
	root := m.Render(nil)
	if !m.AdvanceAnimation() || m.Render(nil) != root {
		t.Fatal("animation must reuse existing layout instead of rebuilding all elements")
	}
	if !m.advanceAnimation(time.Now().Add(3 * time.Second)) {
		t.Fatal("expiry must clear dots with one final frame")
	}
	m.Render(nil)
	if len(m.activeUntil) != 0 || m.AdvanceAnimation() {
		t.Fatal("idle files must stop triggering redraws")
	}
	if m.status["src/main.go"] != "M" {
		t.Fatal("persistent change badge must outlive activity animation")
	}
}

func TestContentScrollbarDrag(t *testing.T) {
	m := fileScreen(t, strings.Repeat("// content\n", 500))
	m.contentTab = true
	render(m, 120, 24)
	rect := m.feed.ContentRect()
	x := rect.Right() - 1
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: x, Y: rect.Y})
	render(m, 120, 24)
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, Action: tui.MouseDrag, X: x, Y: rect.Bottom() - 1})
	render(m, 120, 24)
	_, maximum := m.feed.MaxScroll()
	if !m.changesFocus || m.contentOffset != maximum || m.changeOffset != 0 {
		t.Fatal("dragging Content scrollbar must scroll Content, not Watch")
	}
	m.HandleMouse(tui.MouseEvent{Action: tui.MouseRelease})
}

func TestLargePreviewUsesBoundedElements(t *testing.T) {
	m := fileScreen(t, strings.Repeat("// line\n", 19000))
	m.contentTab = true
	render(m, 120, 24)
	if len(m.feed.Children()) > 102 {
		t.Fatalf("preview created %d elements for 19000 lines", len(m.feed.Children()))
	}
	m.changesFocus = true
	press(t, m, tui.KeyEnd)
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "19000") {
		t.Fatalf("virtualized preview must reach last source lines:\n%s", frame)
	}
}

func BenchmarkActivityFrame(b *testing.B) {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: []watcher.Entry{{Path: "."}}})
	m.entries = testScreen().entries
	m.collapsed["."] = false
	m.collapsed["src"] = false
	m.rebuildRows()
	m.status["src/main.go"] = "M"
	m.activeUntil["src/main.go"] = time.Now().Add(time.Hour)
	m.Render(nil)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		m.AdvanceAnimation()
		m.Render(nil)
	}
}

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

func TestSidebarHorizontalScroll(t *testing.T) {
	m := testScreen()
	name := strings.Repeat("long-name-", 8) + "tail.go"
	m.Apply(watcher.Update{Upsert: []watcher.Entry{{Path: name}},
		Changes: []watcher.Change{{Path: name, Kind: "modified", Time: time.Now()}}})
	render(m, 120, 32)
	maximum, _ := m.tree.MaxScroll()
	if maximum == 0 {
		t.Fatal("long context entries must overflow horizontally")
	}
	for range 100 {
		pressRune(t, m, ']')
	}
	buffer := tui.NewBuffer(120, 32)
	m.Render(nil).RenderTo(buffer, 120, 32)
	rect := m.tree.ContentRect()
	var row strings.Builder
	for x := rect.X; x < rect.Right(); x++ {
		row.WriteRune(buffer.Cell(x, rect.Y+m.rowPositions[name]).Rune)
	}
	if !strings.Contains(row.String(), "tail.go") || m.treeX != maximum {
		t.Fatal("context scrolling must reveal the end of long names")
	}
	for range 100 {
		pressRune(t, m, '[')
	}
	if m.treeX != 0 {
		t.Fatal("context horizontal offset must clamp at zero")
	}
	pressRune(t, m, 'w')
	render(m, 120, 32)
	maximum, _ = m.watchPane.MaxScroll()
	if maximum == 0 {
		t.Fatal("long Watch paths must overflow horizontally")
	}
	for range 100 {
		pressRune(t, m, ']')
	}
	frame := render(m, 120, 32)
	if m.watchX != maximum || !strings.Contains(frame, "tail.go") {
		t.Fatal("Watch scrolling must reveal the end of long paths")
	}
	pressRune(t, m, '[')
	if m.watchX >= maximum {
		t.Fatal("Watch must scroll left")
	}
}

func TestSidebarShiftWheelHorizontalScroll(t *testing.T) {
	m := testScreen()
	name := strings.Repeat("long-name-", 8) + "tail.go"
	m.Apply(watcher.Update{Upsert: []watcher.Entry{{Path: name}},
		Changes: []watcher.Change{{Path: name, Kind: "modified", Time: time.Now()}}})
	render(m, 120, 32)
	m.navigationFocus = true
	rect := m.tree.ContentRect()
	event := tui.MouseEvent{Button: tui.MouseWheelDown, Mod: tui.ModShift, X: rect.X, Y: rect.Y}
	if !m.HandleMouse(event) || m.treeX != 4 || m.navigationFocus || m.treeOffset != 0 {
		t.Fatal("Shift+wheel over Context must focus and scroll it horizontally")
	}
	render(m, 120, 32)
	event.Button = tui.MouseWheelUp
	m.HandleMouse(event)
	if m.treeX != 0 {
		t.Fatal("Shift+wheel up must scroll Context left")
	}
	rect = m.watchPane.ContentRect()
	event.Button, event.X, event.Y = tui.MouseWheelDown, rect.X, rect.Y
	if !m.HandleMouse(event) || !m.watchFocus || m.watchX != 4 || m.changeOffset != 0 {
		t.Fatal("Shift+wheel over Watch must scroll activity horizontally")
	}
}

func TestContextArrowHorizontalScroll(t *testing.T) {
	m := testScreen()
	name := "src/" + strings.Repeat("long-name-", 8) + "tail.go"
	m.Apply(watcher.Update{Upsert: []watcher.Entry{{Path: name}}})
	m.selected = m.rowPositions[name]
	render(m, 120, 32)
	selected := m.selected
	pressRune(t, m, ']')
	render(m, 120, 32)
	x, _ := m.tree.ScrollOffset()
	if x != 4 || m.treeX != 4 || m.selected != selected {
		t.Fatal("] on a file must horizontally scroll Context without moving selection")
	}
	pressRune(t, m, '[')
	if m.treeX != 0 || m.selected != selected {
		t.Fatal("Left must scroll back before navigating to the parent")
	}
	press(t, m, tui.KeyLeft)
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("Left at the beginning must preserve parent navigation")
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

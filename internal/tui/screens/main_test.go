package screens

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func testScreen() *Main {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: []watcher.Entry{
		{Path: ".", Mode: fs.ModeDir},
		{Path: "src", Mode: fs.ModeDir},
		{Path: "src/main.go", Size: 42},
		{Path: "README.md"},
	}})
	m.dashboardTab = false
	for path := range m.collapsed {
		m.collapsed[path] = false
	}
	m.rebuildRows()
	return m
}

func render(m *Main, width, height int) string {
	buffer := tui.NewBuffer(width, height)
	m.Render(nil).RenderTo(buffer, width, height)
	return buffer.StringTrimmed()
}

func press(t *testing.T, m *Main, key tui.Key) {
	t.Helper()
	for _, binding := range m.KeyMap() {
		if binding.Pattern.Key == key {
			binding.Handler(tui.KeyEvent{Key: key})
			return
		}
	}
	t.Fatalf("missing binding for %v", key)
}

func TestMainScreen(t *testing.T) {
	m := testScreen()
	frame := render(m, 120, 24)
	for _, text := range []string{"GUARDRAIL", "Context", "Watch", "/workspace",
		"workspace/", "src/", "main.go", "README.md", "Watching for filesystem changes", "q quit"} {
		if !strings.Contains(frame, text) {
			t.Errorf("missing %q:\n%s", text, frame)
		}
	}
	for _, size := range [][2]int{{80, 24}, {40, 12}, {20, 8}, {120, 40}} {
		render(m, size[0], size[1])
	}
}

func TestTreeKeyboardNavigation(t *testing.T) {
	m := testScreen()
	render(m, 120, 24)
	press(t, m, tui.KeyDown)
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("directories must sort before files")
	}
	press(t, m, tui.KeyLeft)
	if len(m.rows) != 3 || !m.collapsed["src"] {
		t.Fatal("left must collapse directory")
	}
	press(t, m, tui.KeyRight)
	if len(m.rows) != 4 || m.collapsed["src"] {
		t.Fatal("right must expand directory")
	}
	press(t, m, tui.KeyRight)
	if m.rows[m.selected].entry.Path != "src/main.go" {
		t.Fatal("right on expanded folder must select child")
	}
	press(t, m, tui.KeyLeft)
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("left on file must select parent")
	}
	press(t, m, tui.KeyEnter)
	if !m.collapsed["src"] {
		t.Fatal("enter must toggle directory")
	}
	press(t, m, tui.KeyEnd)
	if m.selected != len(m.rows)-1 {
		t.Fatal("end must select last row")
	}
	press(t, m, tui.KeyHome)
	if m.selected != 0 {
		t.Fatal("home must select root")
	}
}

func TestLiveUpdatesPreserveSelection(t *testing.T) {
	m := testScreen()
	m.selected = 2
	update := watcher.Update{Entries: append([]watcher.Entry(nil), m.entries...)}
	update.Entries = append(update.Entries, watcher.Entry{Path: "src/new.go"})
	update.Changes = []watcher.Change{{Path: "src/new.go", Kind: "created", Time: time.Now()}}
	m.Apply(update)
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "created") || !strings.Contains(frame, "src/new.go") {
		t.Fatalf("live change missing:\n%s", frame)
	}
	if m.rows[m.selected].entry.Path != "src/main.go" {
		t.Fatal("selection moved when another entry was inserted")
	}
	m.Apply(watcher.Update{Entries: []watcher.Entry{m.entries[0], m.entries[1]},
		Changes: []watcher.Change{{Path: "src/main.go", Kind: "deleted"}}})
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("deleting selection must select nearest surviving parent")
	}
}

func TestRightOnEmptyDirectoryDoesNotSelectSibling(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{Entries: []watcher.Entry{m.entries[0], m.entries[1], m.entries[3]}})
	m.selected = 1
	press(t, m, tui.KeyRight)
	if m.rows[m.selected].entry.Path != "src" {
		t.Fatal("right on an empty directory must not select its sibling")
	}
}

func TestChangesScrollPreservesHistoryPosition(t *testing.T) {
	m := testScreen()
	changes := make([]watcher.Change, 50)
	for i := range changes {
		changes[i] = watcher.Change{Path: fmt.Sprintf("file%d", i)}
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: changes})
	render(m, 120, 24)
	m.changesFocus = true
	m.move(5)
	previous := m.changes[len(m.changes)-1-m.changeOffset].Path
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "new"}}})
	if got := m.changes[len(m.changes)-1-m.changeOffset].Path; got != previous {
		t.Fatalf("new updates moved history position from %q to %q", previous, got)
	}
	press(t, m, tui.KeyHome)
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "newest"}}})
	if m.changeOffset != 0 {
		t.Fatal("feed must follow new changes when at top")
	}
}

func TestMouseAndScrolling(t *testing.T) {
	m := testScreen()
	for i := range 50 {
		m.entries = append(m.entries, watcher.Entry{Path: fmt.Sprintf("file%02d", i)})
	}
	m.rebuildRows()
	render(m, 120, 24)
	row := m.tree.ContentRect()
	if !m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: row.X, Y: row.Y + 1}) {
		t.Fatal("tree click not handled")
	}
	if m.selected != 1 || !m.collapsed["src"] {
		t.Fatal("click must select and toggle directory")
	}
	render(m, 120, 24)
	rect := m.tree.Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseWheelDown, X: rect.X, Y: rect.Y})
	if m.treeOffset != 3 {
		t.Fatalf("wheel did not scroll: %d", m.treeOffset)
	}
	press(t, m, tui.KeyEnd)
	render(m, 120, 24)
	_, offset := m.tree.ScrollOffset()
	if m.selected < offset || m.selected >= offset+viewport(m.tree) {
		t.Fatal("keyboard selection must scroll into view")
	}
	press(t, m, tui.KeyTab)
	if !m.changesFocus {
		t.Fatal("tab must focus change feed")
	}
	m.changes = make([]watcher.Change, 50)
	render(m, 120, 24)
	rect = m.feed.Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseWheelDown, X: rect.X, Y: rect.Y})
	if m.changeOffset != 3 {
		t.Fatal("wheel must scroll change feed")
	}
	press(t, m, tui.KeyEnd)
	if m.changeOffset != len(m.changes)-viewport(m.feed) {
		t.Fatal("end must scroll to oldest retained change")
	}
}

func TestBoundedHistoryAndControlCharacters(t *testing.T) {
	m := testScreen()
	changes := make([]watcher.Change, historyLimit+20)
	for i := range changes {
		changes[i] = watcher.Change{Path: fmt.Sprintf("file%d", i), Kind: "modified"}
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: changes})
	if len(m.changes) != historyLimit || m.changes[0].Path != "file20" {
		t.Fatal("history must retain only latest changes")
	}
	if got := display("file\n\x1b[2J"); got != `file\n\x1b[2J` {
		t.Fatalf("unsafe filename: %q", got)
	}
}

func TestMainQuitBindings(t *testing.T) {
	bindings := testScreen().KeyMap()
	if bindings[0].Pattern.Rune != 'q' || bindings[0].Handler == nil {
		t.Fatal("missing q quit binding")
	}
	if bindings[1].Pattern.Rune != 'c' || bindings[1].Pattern.Mod != tui.ModCtrl || bindings[1].Handler == nil {
		t.Fatal("missing ctrl+c quit binding")
	}
}

func TestSidebarChangeMarkers(t *testing.T) {
	m := testScreen()
	m.collapsed["src"] = true
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "src/main.go", Kind: "modified"}}})
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "[*] src/") || m.status["src/main.go"] != "M" {
		t.Fatalf("changed file and collapsed parent must be marked:\n%s", frame)
	}
	m.collapsed["src"] = false
	m.rebuildRows()
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "[M] main.go") {
		t.Fatalf("modified file badge missing:\n%s", frame)
	}
	entries := append([]watcher.Entry(nil), m.entries...)
	entries = append(entries, watcher.Entry{Path: "src/new.go"})
	m.Apply(watcher.Update{Entries: entries, Changes: []watcher.Change{{Path: "src/new.go", Kind: "created"}}})
	if m.status["src/new.go"] != "A" || m.status["src/main.go"] != "M" {
		t.Fatal("unrelated changes must not clear existing markers")
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: make([]watcher.Change, historyLimit+1)})
	if m.status["src/main.go"] != "M" {
		t.Fatal("markers must survive history eviction")
	}
	m.Apply(watcher.Update{Entries: m.entries[:2], Changes: []watcher.Change{{Path: "src/main.go", Kind: "deleted"}}})
	if _, exists := m.status["src/main.go"]; exists || m.status["src"] != "*" {
		t.Fatal("deleted file must disappear, with its parent still marked")
	}
}

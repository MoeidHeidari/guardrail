package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestChangedFilesAnimateWithDistinctColor(t *testing.T) {
	m := testScreen()
	if m.AdvanceAnimation() {
		t.Fatal("unchanged files must not animate")
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{
		{Path: "src/main.go", Kind: "modified"},
		{Path: "README.md", Kind: "created"},
	}})
	frames := make(map[string]bool)
	for range 4 {
		frame := render(m, 160, 40)
		frames[frame] = true
		for i, row := range m.rows {
			el := m.tree.Children()[i]
			switch row.entry.Path {
			case "src/main.go", "README.md":
				if !el.TextStyle().Fg.Equal(m.theme.Activity) {
					t.Fatalf("%s must use activity color", row.entry.Path)
				}
				if !strings.Contains(el.Text(), ".") {
					t.Fatalf("%s must have dancing dots", row.entry.Path)
				}
			case "src":
				if !strings.Contains(el.Text(), "[-] [*] src/") {
					t.Fatal("folder marker should remain static")
				}
			}
		}
		if !m.AdvanceAnimation() {
			t.Fatal("changed files must advance animation without input")
		}
	}
	if len(frames) != 4 {
		t.Fatalf("got %d distinct animation frames, want 4", len(frames))
	}
	m.selected = 2
	m.animationOnly = false
	render(m, 160, 24)
	selected := m.tree.Children()[m.selected]
	if !selected.Background().Bg.Equal(m.theme.ActiveSelection) ||
		!selected.TextStyle().Fg.Equal(m.theme.Activity) {
		t.Fatal("selected changed file must retain a distinct, readable activity highlight")
	}
}

func TestAnimationStopsWhenChangedFilesDisappear(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{
		{Path: "src/main.go", Kind: "modified"},
	}})
	m.collapsed["src"] = true
	m.rebuildRows()
	if m.AdvanceAnimation() {
		t.Fatal("collapsed changed files must not trigger animation redraws")
	}
	m.collapsed["src"] = false
	m.rebuildRows()
	render(m, 120, 24)
	if !m.AdvanceAnimation() {
		t.Fatal("expanding changed files must resume animation")
	}
	m.Apply(watcher.Update{Entries: []watcher.Entry{m.entries[0], m.entries[1], m.entries[3]},
		Changes: []watcher.Change{{Path: "src/main.go", Kind: "deleted"}}})
	if m.AdvanceAnimation() {
		t.Fatal("deleted files must stop animating")
	}
	render(m, 120, 24)
	if m.tree.Children()[1].TextStyle().Fg.Equal(m.theme.Activity) {
		t.Fatal("ancestor change markers must not acquire file activity styling")
	}
}

func TestAnimationDoesNotChangePreviewOrNavigation(t *testing.T) {
	m := fileScreen(t, "package main\n")
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{
		{Path: "main.go", Kind: "modified"},
	}})
	press(t, m, tui.KeyEnter)
	render(m, 120, 24)
	before := m.preview.lines
	for range 8 {
		m.AdvanceAnimation()
		render(m, 120, 24)
	}
	if m.selected != 1 || !m.contentTab || m.previewDirty || &m.preview.lines[0] != &before[0] {
		t.Fatal("animation must preserve selection, tab, and cached file preview")
	}
}

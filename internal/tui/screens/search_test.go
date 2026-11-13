package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func typeSearch(t *testing.T, m *Main, text string) {
	t.Helper()
	for _, char := range text {
		handled := false
		for _, binding := range m.KeyMap() {
			if binding.Pattern.AnyRune {
				binding.Handler(tui.KeyEvent{Key: tui.KeyRune, Rune: char})
				handled = true
				break
			}
		}
		if !handled {
			t.Fatal("search must capture typed characters")
		}
	}
}

func pressSearchShortcut(t *testing.T, m *Main) {
	t.Helper()
	for _, binding := range m.KeyMap() {
		if binding.Pattern.Rune == '/' && binding.Pattern.Mod == 0 {
			binding.Handler(tui.KeyEvent{Key: tui.KeyRune, Rune: '/'})
			return
		}
	}
	t.Fatal("missing / search binding")
}

func TestGlobalFileSearch(t *testing.T) {
	m := fileScreen(t, "package main\n")
	pressSearchShortcut(t, m)
	typeSearch(t, m, "MAIN")
	frame := render(m, 120, 32)
	if !m.searchOpen || !strings.Contains(frame, "1 matches") || !strings.Contains(frame, "main.go") {
		t.Fatalf("case-insensitive file search must render results:\n%s", frame)
	}
	press(t, m, tui.KeyRight)
	if !m.searchOpen || !m.searchPreviewFocus || len(m.contentTabs) != 1 {
		t.Fatal("Right must focus the search preview without adding a tab")
	}
	if m.rows[m.selected].entry.Path != "main.go" {
		t.Fatal("previewing must reveal the file in Context")
	}
	if frame = render(m, 120, 32); !strings.Contains(frame, "package main") {
		t.Fatalf("search result content must be visible:\n%s", frame)
	}
	if m.searchViewer.feed.Rect().X <= m.searchRows[0].Rect().X {
		t.Fatal("search content must appear to the right of results")
	}
	press(t, m, tui.KeyLeft)
	m.searchQuery = nil
	typeSearch(t, m, "qovw123")
	if string(m.searchQuery) != "qovw123" || !m.searchOpen || m.contentActive != "main.go" {
		t.Fatal("typing shortcuts in search must not trigger workspace actions")
	}
	press(t, m, tui.KeyEnter)
	if !m.searchOpen {
		t.Fatal("Enter on empty results must keep search open")
	}
	press(t, m, tui.KeyEscape)
	if m.searchOpen || m.contentActive != "main.go" {
		t.Fatal("Esc must restore the previous content")
	}
}

func TestSearchPathsUnicodeAndLiveUpdates(t *testing.T) {
	m := testScreen()
	m.openSearch()
	typeSearch(t, m, "src/")
	if string(m.searchQuery) != "src/" {
		t.Fatal("slash inside search must be typed, not reopen search")
	}
	if paths := m.searchMatches(); len(paths) != 1 || paths[0] != "src/main.go" {
		t.Fatalf("search must match paths and exclude folders: %v", paths)
	}
	m.openSearch()
	typeSearch(t, m, "é")
	press(t, m, tui.KeyBackspace)
	if len(m.searchQuery) != 0 {
		t.Fatal("backspace must remove a Unicode character")
	}
	typeSearch(t, m, "new")
	m.Apply(watcher.Update{Upsert: []watcher.Entry{{Path: "new.go"}}})
	if paths := m.searchMatches(); len(paths) != 1 || paths[0] != "new.go" {
		t.Fatal("newly discovered files must be searchable immediately")
	}
	m.Apply(watcher.Update{Removed: []string{"new.go"}})
	if len(m.searchMatches()) != 0 {
		t.Fatal("removed files must disappear from results")
	}
}

func TestSearchMouseAndResultNavigation(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	button := m.searchButton.Rect()
	m.HandleMouse(tui.MouseEvent{Action: tui.MousePress, Button: tui.MouseLeft, X: button.X, Y: button.Y})
	if !m.searchOpen {
		t.Fatal("clicking the header search must focus it")
	}
	press(t, m, tui.KeyDown)
	if m.searchSelection != 1 {
		t.Fatal("Down must select the next result")
	}
	press(t, m, tui.KeyEnd)
	press(t, m, tui.KeyUp)
	render(m, 120, 32)
	row := m.searchRows[0].Rect()
	m.HandleMouse(tui.MouseEvent{Action: tui.MousePress, Button: tui.MouseLeft, X: row.X, Y: row.Y})
	if !m.searchOpen || !m.searchPreviewFocus || m.searchViewer.rows[0].entry.Path != "README.md" {
		t.Fatal("clicking a result must preview the corresponding file")
	}
}

func TestSearchPreviewHomeEndPreserveResultSelection(t *testing.T) {
	m := fileScreen(t, strings.Repeat("package main\n", 100))
	m.openSearch()
	typeSearch(t, m, "main")
	render(m, 120, 32)
	press(t, m, tui.KeyRight)
	render(m, 120, 32)
	selection := m.searchSelection
	press(t, m, tui.KeyEnd)
	viewer := m.searchViewer
	expected := max(0, len(viewer.preview.lines)-viewport(viewer.feed))
	if expected == 0 || viewer.contentOffset != expected || m.searchSelection != selection {
		t.Fatal("End must scroll the focused preview, not navigate search results")
	}
	press(t, m, tui.KeyHome)
	if viewer.contentOffset != 0 || m.searchSelection != selection {
		t.Fatal("Home must return the focused preview to its first line")
	}
}

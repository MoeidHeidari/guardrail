package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestWatchSelectionAndOpening(t *testing.T) {
	m := fileScreen(t, "package main\n")
	m.changes = []watcher.Change{{Path: "gone.go", Kind: "deleted"}, {Path: "main.go", Kind: "modified"}}
	m.openNavigation(len(workspaceLabels))
	render(m, 120, 32)
	press(t, m, tui.KeyDown)
	if m.watchSelection != 1 {
		t.Fatal("Down must select the next changed file")
	}

	press(t, m, tui.KeyUp)
	press(t, m, tui.KeyRight)
	if m.contentActive != "main.go" || !m.changesFocus || m.watchFocus {
		t.Fatal("Right must open the selected changed file and focus its content")
	}
	if frame := render(m, 120, 32); !strings.Contains(frame, "package main") {
		t.Fatalf("changed file content not shown:\n%s", frame)
	}
	m.openNavigation(len(workspaceLabels))
	press(t, m, tui.KeyDown)
	press(t, m, tui.KeyEnter)
	if frame := render(m, 120, 32); !strings.Contains(frame, "no longer available") {
		t.Fatalf("deleted files need an explicit message:\n%s", frame)
	}
	m.openNavigation(len(workspaceLabels))
	press(t, m, tui.KeyHome)
	press(t, m, tui.KeyUp)
	if m.watchFocus || m.navigationFocus || m.changesFocus {
		t.Fatal("Up from the first Watch row must focus Context")
	}
}

func TestWatchReachableFromContext(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	m.focusContext()
	m.selected = len(m.rows) - 1
	press(t, m, tui.KeyDown)
	if !m.watchFocus {
		t.Fatal("Down from the last Context row must focus Watch")
	}
	press(t, m, tui.KeyUp)
	if m.watchFocus || m.changesFocus || m.navigationFocus {
		t.Fatal("Up from empty Watch must return to Context")
	}
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyTab)
	if !m.watchFocus {
		t.Fatal("Tab through the main pane must reach Watch")
	}
	render(m, 120, 32)
	if !m.watchTabButton.Background().Bg.Equal(m.theme.Selection) {
		t.Fatal("Watch heading must visibly indicate focus")
	}
}

func TestFocusHeadingColors(t *testing.T) {
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		m := testScreen()
		m.theme = theme
		m.openWorkspaceFocused(violationsPage)
		render(m, 120, 32)
		if !m.dashboardButton.Background().Bg.Equal(theme.Selection) ||
			!m.dashboardButton.TextStyle().Fg.Equal(theme.SelectionText) {
			t.Fatal("focused heading needs a subtle tint and readable text")
		}
		m.focusContext()
		render(m, 120, 32)
		if background := m.dashboardButton.Background(); background != nil && background.Bg.Equal(theme.Selection) {
			t.Fatal("inactive heading must lose focus highlighting")
		}
	}
}

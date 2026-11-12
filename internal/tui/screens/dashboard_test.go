package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestDashboardDefaultAndNavigation(t *testing.T) {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: testScreen().entries})
	frame := render(m, 120, 32)
	if !m.dashboardVisible() || !strings.Contains(frame, "Workspace") ||
		!m.dashboardButton.TextStyle().Fg.Equal(m.theme.Accent) {
		t.Fatal("Dashboard must be the default main tab")
	}
	for _, tab := range []*tui.Element{m.dashboardButton, m.watchTabButton} {
		if !strings.Contains(frame, tab.Text()) {
			t.Fatalf("tab %q must always be available", tab.Text())
		}
	}
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyEnd)
	if m.changeOffset != 0 || m.contentOffset != 0 {
		t.Fatal("Dashboard navigation must not scroll inactive tabs")
	}
	pressRune(t, m, '1')
	frame = render(m, 120, 32)
	if !m.dashboardVisible() || !m.watchFocus || !strings.Contains(frame, "Waiting for saved changes") {
		t.Fatal("Watch shortcut must focus sidebar activity without replacing Dashboard")
	}
	button := m.dashboardButton.Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: button.X, Y: button.Y})
	frame = render(m, 120, 32)
	if !m.dashboardVisible() || !strings.Contains(frame, "/   Overview") {
		t.Fatal("Dashboard mouse tab must restore dashboard and focus label")
	}
	pressRune(t, m, '2')
	if m.contentTabButton != nil || m.contentTab {
		t.Fatal("Content must not open before the first file is opened")
	}
	m.changesFocus = false
	m.selected = m.rowPositions["README.md"]
	press(t, m, tui.KeyEnter)
	render(m, 120, 32)
	if m.dashboardVisible() || !m.contentTab {
		t.Fatal("Content shortcut must leave Dashboard")
	}
	pressRune(t, m, '0')
	render(m, 120, 32)
	if !m.dashboardVisible() {
		t.Fatal("Dashboard keyboard shortcut must restore Dashboard")
	}
}

func TestDashboardUpdatesWhileOtherTabsRemainAvailable(t *testing.T) {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: testScreen().entries})
	m.Apply(watcher.Update{Upsert: []watcher.Entry{{Path: "src/main.go", Size: 99}},
		Changes: []watcher.Change{{Path: "src/main.go", Kind: "modified"}}})
	frame := render(m, 120, 32)
	if !strings.Contains(frame, "1 changed files") {
		t.Fatal("Dashboard must reflect live change counts")
	}
	pressRune(t, m, '1')
	if !strings.Contains(render(m, 120, 32), "src/main.go") {
		t.Fatal("Watch history must continue updating while Dashboard is active")
	}
}

func TestWorkspaceHasNoOuterMargins(t *testing.T) {
	m := testScreen()
	root := m.Render(nil)
	buffer := tui.NewBuffer(120, 32)
	root.RenderTo(buffer, 120, 32)
	header, body, footer := root.Children()[0], root.Children()[1], root.Children()[2]
	if header.Rect().X != 0 || header.Rect().Y != 0 || header.Rect().Right() != 120 ||
		header.Rect().Height != 3 || footer.Rect().X != 0 || footer.Rect().Right() != 120 ||
		footer.Rect().Bottom() != 32 || body.Rect().X != 0 || body.Rect().Right() != 120 {
		t.Fatal("workspace must fill the terminal with no outer empty bands")
	}
	for _, pane := range body.Children() {
		if pane.Style().Padding.Horizontal() != 4 || pane.Style().Padding.Top != 1 || pane.Style().Padding.Bottom != 1 {
			t.Fatal("workspace panels must retain their internal padding")
		}
		if header.Children()[0].Rect().Y < 1 || header.Children()[0].Rect().Bottom() >= header.Rect().Bottom() ||
			footer.Children()[0].Rect().Y != 30 ||
			m.watchPane.Rect().Bottom() != body.Rect().Bottom()-1 || m.feed.Rect().Bottom() != body.Rect().Bottom()-1 {
			t.Fatal("top and bottom spacing must be internal padding, not outer margins")
		}
	}
	if header.Children()[0].Rect().X != 2 || footer.Children()[0].Rect().X != 2 ||
		m.tree.Rect().X != 2 || m.feed.Rect().Right() != 118 {
		t.Fatal("content must be padded inside edge-to-edge panels")
	}
}

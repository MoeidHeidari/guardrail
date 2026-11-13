package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestTabScrollSurvivesWorkspacePreview(t *testing.T) {
	m := fileScreen(t, strings.Repeat("// "+strings.Repeat("x", 200)+"\n", 100))
	m.openTab("main.go")
	render(m, 120, 32)
	m.contentX, m.contentOffset = 12, 20
	m.openWorkspace(overviewPage)
	m.contentX, m.contentOffset = 0, 0
	m.openTab("main.go")
	render(m, 120, 32)
	if m.contentX != 12 || m.contentOffset != 20 {
		t.Fatalf("tab position lost after workspace preview: %d,%d", m.contentX, m.contentOffset)
	}
	m.focusMain()
	press(t, m, tui.KeyLeft)
	if m.changesFocus || m.contentX != 12 {
		t.Fatal("Left must return focus without changing horizontal scroll")
	}
}

func TestClosingRememberedTabPreservesWorkspace(t *testing.T) {
	m := overviewFixture(t)
	m.openTab("main.go")
	m.openTab("other.go")
	m.openWorkspace(overviewPage)
	m.contentX, m.contentOffset = 4, 8
	m.closeContent()
	if m.contentTab || m.contentActive != "main.go" || m.workspacePage != overviewPage {
		t.Fatal("closing a remembered content tab must not activate it over the workspace")
	}
	if m.contentX != 4 || m.contentOffset != 8 {
		t.Fatal("closing an inactive tab must preserve workspace preview offsets")
	}
}

func TestAsyncTabSwitchUsesRestoredOffsets(t *testing.T) {
	m := fileScreen(t, "package main\n")
	m.openTab("main.go")
	m.tabViews = map[string][2]int{"main.go": {12, 20}}
	m.preview.path = "other.go"
	m.previewRequests = make(chan previewRequest, 1)
	m.requestPreview()
	request := <-m.previewRequests
	if request.x != 12 || request.y != 20 {
		t.Fatalf("worker must receive target tab offsets, got %d,%d", request.x, request.y)
	}
}

func TestWatchSelectionSurvivesIncomingChanges(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{Changes: []watcher.Change{{Path: "first.go", Kind: "modified"}}})
	m.openNavigation(len(workspaceLabels))
	m.Apply(watcher.Update{Changes: []watcher.Change{{Path: "second.go", Kind: "modified"}}})
	if m.watchSelection != 1 || m.changes[len(m.changes)-1-m.watchSelection].Path != "first.go" {
		t.Fatal("new events must not change the selected Watch file")
	}
}

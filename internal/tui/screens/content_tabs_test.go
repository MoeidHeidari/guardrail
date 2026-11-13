package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

func TestMultipleContentTabs(t *testing.T) {
	root := t.TempDir()
	for name, text := range map[string]string{"a.go": "package a\n", "b.go": "package b\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	m := testScreen()
	m.root = root
	m.contentTabs = nil
	m.openTab("a.go")
	m.openTab("b.go")
	m.openTab("a.go")
	if len(m.contentTabs) != 2 || m.contentActive != "a.go" {
		t.Fatalf("tabs must be unique and reuse existing ones: %v", m.contentTabs)
	}
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "a.go") || !strings.Contains(frame, "b.go") || strings.Count(frame, "[x]") < 2 {
		t.Fatalf("every open file needs a tab with a close button:\n%s", frame)
	}
	m.cycleTab(1)
	if m.contentActive != "b.go" {
		t.Fatal("next tab must activate the following file")
	}
	m.closeContent()
	if m.contentActive != "a.go" || len(m.contentTabs) != 1 || !m.contentTab {
		t.Fatal("closing the active tab must activate a neighbour")
	}
	m.closeContent()
	if m.contentOpened || m.contentTab || !m.dashboardTab {
		t.Fatal("closing the last tab must return to the workspace")
	}
}

func TestManyTabsStayReachable(t *testing.T) {
	m := testScreen()
	m.contentTabs = nil
	for i := range 30 {
		m.openTab(fmt.Sprintf("dir/file-%02d.go", i))
	}
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "file-29.go") || !strings.Contains(frame, "‹") {
		t.Fatalf("active tab must stay visible with hidden tabs marked:\n%s", frame)
	}
	m.jumpTab(1)
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "file-00.go") || !strings.Contains(frame, "›") {
		t.Fatalf("jumping to the first tab must reveal it:\n%s", frame)
	}
	m.cycleTab(-1)
	m.cycleTab(-1)
	if m.contentActive != "dir/file-29.go" {
		t.Fatalf("cycling must wrap through the workspace tab: %s", m.contentActive)
	}
}

func TestArrowsTransferFocus(t *testing.T) {
	m := testScreen()
	m.contentTab, m.dashboardTab = false, true
	m.navigationFocus, m.changesFocus = true, false
	render(m, 120, 24)
	press(t, m, tui.KeyRight)
	if !m.changesFocus || m.navigationFocus {
		t.Fatal("Right on a workspace item must open it and move focus to the page")
	}
	press(t, m, tui.KeyLeft)
	if !m.navigationFocus || m.changesFocus {
		t.Fatal("Left at the page edge must return to the workspace menu")
	}
	m.navigationFocus = false
	m.selected = m.rowPositions["src/main.go"]
	press(t, m, tui.KeyRight)
	if !m.contentTab || !m.changesFocus || m.contentActive != "src/main.go" {
		t.Fatal("Right on a file must open it and focus the content")
	}
	press(t, m, tui.KeyEscape)
	if m.changesFocus || m.navigationFocus {
		t.Fatal("Esc must return focus to the Context tree")
	}
}

func TestFilterArrowsAndEdgeJumps(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	m.focusMain()
	press(t, m, tui.KeyRight)
	if m.findingFilter != "violation" {
		t.Fatalf("Right must select the next filter, got %q", m.findingFilter)
	}
	press(t, m, tui.KeyRight)
	press(t, m, tui.KeyLeft)
	press(t, m, tui.KeyLeft)
	if m.findingFilter != "" || !m.changesFocus {
		t.Fatal("Left steps back through the filters")
	}
	press(t, m, tui.KeyLeft)
	if !m.navigationFocus {
		t.Fatal("Left on the first filter must return to the menu")
	}
	m.navigationIndex = len(workspaceLabels) - 1
	press(t, m, tui.KeyDown)
	if m.navigationFocus || m.changesFocus {
		t.Fatal("Down from the last workspace item must enter Context")
	}
	m.selected = 0
	press(t, m, tui.KeyUp)
	if !m.navigationFocus || m.navigationIndex != len(workspaceLabels)-1 {
		t.Fatal("Up from the first Context row must enter the workspace menu")
	}
}

func TestSidebarSectionShortcuts(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	for key, page := range map[rune]workspacePage{'o': overviewPage, 'v': violationsPage, 'u': rulesPage} {
		pressRune(t, m, key)
		if m.workspacePage != page || !m.dashboardVisible() || !m.changesFocus {
			t.Fatalf("%c must open its page and focus it", key)
		}
	}
	pressRune(t, m, 'c')
	if m.changesFocus || m.navigationFocus || m.watchFocus {
		t.Fatal("c must focus Context")
	}
	pressRune(t, m, 'w')
	if !m.watchFocus {
		t.Fatal("w must focus Watch")
	}
}

func TestRulesLeftReturnsToSidebar(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspaceFocused(rulesPage)
	render(m, 160, 40)
	for range 6 {
		press(t, m, tui.KeyLeft)
		render(m, 160, 40)
	}
	if !m.navigationFocus || m.changesFocus {
		t.Fatal("repeated Left in Rules must end in the workspace menu")
	}
}

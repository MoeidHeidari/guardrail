package screens

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

func TestWorkspaceSidebarPages(t *testing.T) {
	m := testScreen()
	frame := render(m, 120, 32)
	for _, text := range []string{"Overview", "Violations", "Rules", "Context", "main.go"} {
		if !strings.Contains(frame, text) {
			t.Fatalf("missing sidebar content %q", text)
		}
	}
	selected, offset := m.selected, m.treeOffset
	for i, text := range []string{"Add .guard/rules.yml", "Violations", "Configure regex or JavaScript"} {
		rect := m.navigationItems[i].Rect()
		m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
		frame = render(m, 120, 32)
		if !strings.Contains(frame, text) || m.workspacePage != workspacePage(i) {
			t.Fatalf("sidebar page %d not activated:\n%s", i, frame)
		}
		if m.selected != selected || m.treeOffset != offset {
			t.Fatal("page changes must preserve context navigation")
		}
		if !strings.Contains(frame, "Watch  [w]") {
			t.Fatal("utility tabs must remain available")
		}
	}
	pressRune(t, m, 'w')
	render(m, 120, 32)
	if !m.watchFocus || !m.dashboardVisible() {
		t.Fatal("Watch must remain accessible alongside Rules")
	}
	pressRune(t, m, '0')
	render(m, 120, 32)
	if m.workspacePage != rulesPage || !m.dashboardVisible() {
		t.Fatal("main tab must remember the selected workspace page")
	}
}

func TestSidebarKeyboardFocus(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	press(t, m, tui.KeyTab)
	if !m.changesFocus {
		t.Fatal("Tab from context must reach content")
	}

	press(t, m, tui.KeyTab)
	if !m.watchFocus {
		t.Fatal("Tab from content must reach Watch")
	}
	press(t, m, tui.KeyTab)
	if !m.navigationFocus || m.changesFocus {
		t.Fatal("Tab from Watch must reach workspace navigation")
	}
	press(t, m, tui.KeyDown)
	press(t, m, tui.KeyEnter)
	if m.workspacePage != violationsPage {
		t.Fatal("keyboard menu activation must open Violations")
	}
	if !m.changesFocus {
		t.Fatal("Enter on a workspace item must move focus into the page")
	}
	press(t, m, tui.KeyEscape)
	press(t, m, tui.KeyTab)
	if m.navigationFocus || m.changesFocus {
		t.Fatal("Tab from navigation must return to context")
	}
}

func TestWorkspaceFocusIncludesWatch(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	m.navigationFocus = true
	press(t, m, tui.KeyEnd)
	if m.navigationIndex != len(workspaceLabels)-1 || m.watchFocus {
		t.Fatal("workspace navigation must stop at Rules, not Watch")
	}
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyTab)
	if !m.watchFocus {
		t.Fatal("Tab from Context must focus Watch")
	}
	press(t, m, tui.KeyTab)
	if !m.navigationFocus || m.watchFocus || len(m.navigationItems) != len(workspaceLabels) {
		t.Fatal("Tab cycle must return to workspace navigation after Watch and the main pane")
	}
}

func TestSidebarSectionsDoNotOverlap(t *testing.T) {
	m := testScreen()
	render(m, 120, 32)
	last := m.navigationItems[len(workspaceLabels)-1].Rect()
	if last.Bottom() >= m.tree.Rect().Y {
		t.Fatal("workspace navigation and context need separate vertical regions")
	}

	if viewport(m.tree) < 7 {
		t.Fatal("context tree must retain most of the sidebar height")
	}
	if m.watchTabButton.Rect().Y < m.tree.Rect().Bottom() {
		t.Fatal("Watch must sit below the scrollable context")
	}
}

func TestSidebarDividers(t *testing.T) {
	m := testScreen()
	root := m.Render(nil)
	buffer := tui.NewBuffer(120, 32)
	root.RenderTo(buffer, 120, 32)
	sidebar := root.Children()[1].Children()[0]
	children := sidebar.Children()
	if len(children) != 5 {
		t.Fatal("sidebar must contain three sections and two divider lines")
	}

	for _, index := range []int{1, 3} {
		rect := children[index].Rect()
		if rect.Height != 1 || rect.Width != sidebar.ContentRect().Width {
			t.Fatal("dividers must be single lines across the sidebar content")
		}
		cell := buffer.Cell(rect.X, rect.Y)
		if cell.Rune == ' ' || !cell.Style.Fg.Equal(m.theme.Separator) {
			t.Fatal("dividers must render in the subtle separator color")
		}
		for x := rect.X; x < rect.Right(); x++ {
			if !buffer.Cell(x, rect.Y).Style.Bg.Equal(m.theme.Surface) {
				t.Fatal("divider must use the sidebar background, not a black rectangle")
			}
		}
	}
	if !children[0].Background().Bg.Equal(m.theme.Surface) ||
		!children[4].Background().Bg.Equal(m.theme.Surface) {
		t.Fatal("sections must share the sidebar background without tinted cards")
	}
}

func TestSidebarItemPadding(t *testing.T) {
	m := testScreen()
	m.openWorkspace(overviewPage)
	render(m, 120, 32)
	for _, item := range m.navigationItems {
		padding := item.Style().Padding
		if padding.Left != 1 || padding.Right != 1 || padding.Top != 0 ||
			padding.Bottom != 0 || item.Rect().Height != 1 {
			t.Fatal("sidebar items need compact single-line spacing")
		}
	}
	for _, item := range m.tree.Children() {
		if item.Style().Padding.Left != 1 || item.Style().Padding.Right != 1 ||
			item.Style().Padding.Top != 0 || item.Style().Padding.Bottom != 0 ||
			item.Rect().Height != 1 {
			t.Fatal("context rows must use the same compact padding as navigation and Watch")
		}
	}
}

func TestReferenceSelectionAccent(t *testing.T) {
	m := testScreen()
	m.openWorkspace(overviewPage)
	root := m.Render(nil)
	buffer := tui.NewBuffer(120, 32)
	root.RenderTo(buffer, 120, 32)
	for _, item := range []*tui.Element{m.navigationItems[0], m.tree.Children()[m.selected]} {
		rect := item.Rect()
		if item.Parent() == m.tree {
			rect.X += m.tree.ContentRect().X
			rect.Y += m.tree.ContentRect().Y
		}
		if rect.Height != 1 || !item.Border().All(tui.BorderNone) {
			t.Fatalf("sidebar rows must remain borderless: rect=%+v border=%+v", rect, item.Border())
		}
		if !buffer.Cell(rect.X+1, rect.Y).Style.Bg.Equal(m.theme.Selection) {
			t.Fatal("selection inset must share the highlighted background")
		}
	}
}

func TestContentAppearsAfterOpeningFile(t *testing.T) {
	m := testScreen()
	m.contentOpened = false
	m.openWorkspace(overviewPage)
	render(m, 120, 32)
	if m.contentTabButton != nil {
		t.Fatal("Content must not be offered before a file is opened")
	}
	for _, page := range []workspacePage{overviewPage, violationsPage, rulesPage} {
		m.openWorkspace(page)
		render(m, 120, 32)
		if m.dashboardButton.Text() != workspaceLabels[page] || m.contentTabButton != nil {
			t.Fatal("a single workspace page must have a plain heading without a shortcut number")
		}
		for _, child := range m.dashboardButton.Parent().Parent().Children() {
			if child.Text() == workspaceLabels[page] {
				t.Fatal("workspace heading must not be repeated as a subtitle")
			}
		}
	}
	m.selected = m.rowPositions["src/main.go"]
	m.toggle()
	render(m, 120, 32)
	if !m.contentOpened || !m.contentTab || m.contentTabButton == nil {
		t.Fatal("opening a file must reveal Content")
	}
	if m.dashboardButton.Text() != "0 "+workspaceLabels[m.workspacePage] {
		t.Fatal("opening Content must reveal the numbered workspace tab")
	}
	m.openWorkspace(overviewPage)
	render(m, 120, 32)
	if m.contentTabButton == nil {
		t.Fatal("Content must remain available after its first use")
	}
	m.openNavigation(len(workspaceLabels))
	render(m, 120, 32)
	if !m.watchFocus || !m.dashboardVisible() || m.contentTab {
		t.Fatal("sidebar Watch must focus activity alongside workspace pages")
	}
	if len(m.watchTabButton.Parent().Children()) != 2 {
		t.Fatal("Watch must contain a heading and sidebar activity pane")
	}
}

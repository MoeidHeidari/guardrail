package screens

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

func clickClose(t *testing.T, m *Main, button *tui.Element) {
	t.Helper()
	if button == nil {
		t.Fatal("missing close button")
	}
	rect := button.Rect()
	if rect.Width < 3 || !m.HandleMouse(tui.MouseEvent{Action: tui.MousePress, Button: tui.MouseLeft, X: rect.X, Y: rect.Y}) {
		t.Fatalf("close button not clickable: %+v", rect)
	}
}

func TestPreviewCloseAndReopen(t *testing.T) {
	m := overviewFixture(t)
	render(m, 160, 40)
	rect := m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	render(m, 160, 40)
	m.overviewPreviewFocus = true
	clickClose(t, m, m.overviewClose)
	frame := render(m, 160, 40)
	if m.overviewPreview != nil || m.overviewPreviewFocus || strings.Contains(frame, "Live content") {
		t.Fatal("closing Overview preview must hide it and restore list focus")
	}
	pressRune(t, m, 'x')
	rect = m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	render(m, 160, 40)
	if m.overviewPreview == nil {
		t.Fatal("Overview preview must reopen")
	}
	pressRune(t, m, 'x')
	if m.overviewPreviewOpened {
		t.Fatal("x must close the Overview preview")
	}
	m.openWorkspace(rulesPage)
	render(m, 160, 40)
	m.changesFocus, m.rulesFocus = true, 1
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	clickClose(t, m, m.rulesClose)
	render(m, 160, 40)
	if m.rulesPreview != nil || m.rulesFocus != 1 {
		t.Fatal("closing Rules preview must restore file focus")
	}
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if m.rulesPreview == nil {
		t.Fatal("Rules preview must reopen")
	}
	pressRune(t, m, 'x')
	if m.rulesPreviewOpened {
		t.Fatal("x must close the Rules preview")
	}
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	m.changesFocus = true
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	clickClose(t, m, m.contentClose)
	render(m, 160, 40)
	if m.contentTab || m.contentOpened || m.workspacePage != violationsPage || m.contentTabButton != nil {
		t.Fatal("closing Content must restore the workspace and remove its tab")
	}
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if !m.contentTab {
		t.Fatal("Content must reopen")
	}
	pressRune(t, m, 'x')
	if m.contentTab || m.contentOpened {
		t.Fatal("x must close Content")
	}
}

func TestClosingPreviewRejectsPendingResult(t *testing.T) {
	m := overviewFixture(t)
	m.overviewPreviewOpened = true
	render(m, 160, 40)
	result := PreviewResult{id: m.previewID, preview: m.preview}
	m.closeOverviewPreview()
	if m.ApplyPreview(result) {
		t.Fatal("closing a preview must invalidate pending results")
	}
}

func TestCloseInactiveContentTab(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	m.changesFocus = true
	press(t, m, tui.KeyEnter)
	m.openWorkspace(overviewPage)
	render(m, 160, 40)
	rect := m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	render(m, 160, 40)
	clickClose(t, m, m.contentClose)
	render(m, 160, 40)
	if m.contentOpened || m.contentTab || m.workspacePage != overviewPage || m.overviewPreview == nil {
		t.Fatal("closing an inactive Content tab must preserve the workspace preview")
	}
}

func TestOverviewPreviewAppearsOnFileSelection(t *testing.T) {
	m := overviewFixture(t)
	frame := render(m, 160, 40)
	if m.overviewPreview != nil || strings.Contains(frame, "Live content") || m.overviewPath != "" {
		t.Fatal("Overview must not load or reserve a content panel before selection")
	}
	rect := m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	frame = render(m, 160, 40)
	if m.overviewPreview == nil || !strings.Contains(frame, "var first = 1") {
		t.Fatal("clicking an overview finding must reveal its live content")
	}
}

func TestRulesFocusSkipsHiddenContent(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspace(rulesPage)
	render(m, 160, 40)
	m.changesFocus = true
	press(t, m, tui.KeyTab)
	if m.rulesFocus != 1 || m.rulesPreview != nil {
		t.Fatal("Tab must reach files without opening content")
	}
	press(t, m, tui.KeyTab)
	if !m.navigationFocus || m.rulesFocus != 0 {
		t.Fatal("Tab must skip a hidden content pane")
	}
	m.navigationFocus, m.changesFocus, m.rulesFocus = false, true, 1
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if m.rulesPreview == nil || m.rulesFocus != 2 {
		t.Fatal("Enter on a finding must reveal and focus content")
	}
}

func TestRulesPreviewReopensAtSelectedFinding(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspace(rulesPage)
	render(m, 160, 40)
	m.changesFocus, m.rulesFocus, m.rulesFileSelection = true, 1, 1
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	m.closeRulesPreview()
	render(m, 160, 40)
	m.preview.cursor = nil
	m.jumpLine, m.jumpColumn = 0, 0
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if m.preview.cursor == nil || m.preview.cursor.line != 2 {
		t.Fatal("reopening Rules must restore the selected finding location")
	}
}

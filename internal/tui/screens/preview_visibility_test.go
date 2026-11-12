package screens

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

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

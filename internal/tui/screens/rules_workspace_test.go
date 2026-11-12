package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/grindlemire/go-tui"
)

func TestRulesWorkspaceSelectionPreviewAndFocus(t *testing.T) {
	m := overviewFixture(t)
	m.analysisResult.Rules = []config.Rule{{ID: "test", Set: "quality", Type: "regex"},
		{ID: "unused", Set: "quality", Type: "regex"}}
	m.openWorkspace(rulesPage)
	render(m, 160, 40)
	if m.rulesPreview != nil {
		t.Fatal("Rules must not show content before opening a file finding")
	}
	m.changesFocus, m.rulesFocus = true, 1
	press(t, m, tui.KeyEnter)
	m.rulesFocus = 0
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "Live content / main.go") || !strings.Contains(frame, "var first = 1") {
		t.Fatalf("rules must show embedded content:\n%s", frame)
	}
	if m.rulesFiles.Rect().X <= m.rulesList.Rect().Right() || m.rulesPreview.Rect().Y <= m.rulesList.Rect().Bottom() {
		t.Fatal("rules and files must be side by side above full-width content")
	}
	for _, size := range [][2]int{{160, 40}, {120, 32}, {100, 24}} {
		render(m, size[0], size[1])
		if m.rulesPreview.Rect().Height <= m.rulesList.Rect().Height ||
			m.rulesList.Rect().Height < 1 || m.rulesFiles.Rect().Height < 1 {
			t.Fatalf("content must get more height than usable rule/file lists at %dx%d: list=%+v files=%+v preview=%+v", size[0], size[1], m.rulesList.Rect(), m.rulesFiles.Rect(), m.rulesPreview.Rect())
		}
	}
	render(m, 160, 40)
	m.changesFocus = true
	press(t, m, tui.KeyTab)
	if m.rulesFocus != 1 {
		t.Fatal("Tab must reach matching files")
	}
	press(t, m, tui.KeyDown)
	render(m, 160, 40)
	if m.contentTab || m.preview.cursor == nil || m.preview.cursor.line != 2 {
		t.Fatal("file selection must preview exact finding without leaving Rules")
	}
	m.rulesFocus = 0
	press(t, m, tui.KeyEnd)
	render(m, 160, 40)
	if len(m.rulesMatches) != 0 || m.rulesPath != "" {
		t.Fatal("unaffected rule must clear file list and content")
	}
	m.rulesSelection = 0
	render(m, 160, 40)
	rect := m.rulesFiles.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X + 1, Y: rect.Y})
	render(m, 160, 40)
	if m.rulesFocus != 1 || m.rulesFileSelection != 0 || m.contentTab {
		t.Fatal("mouse must select a matching file")
	}
	render(m, 90, 40)
	if m.rulesFiles.Rect().Y < m.rulesList.Rect().Bottom() {
		t.Fatal("narrow Rules must stack lists")
	}
	m.findings = []analysis.Finding{}
	render(m, 160, 40)
	if m.rulesPath != "" {
		t.Fatal("resolved findings must clear preview")
	}
}

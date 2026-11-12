package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/grindlemire/go-tui"
)

func TestOverviewFiltersNavigationAndPreview(t *testing.T) {
	m := overviewFixture(t)
	render(m, 160, 40)
	pressRune(t, m, 'e')
	m.changesFocus = true
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if m.findingIndex != 1 || len(m.overviewList.Children()) != 1 || m.preview.cursor.line != 2 {
		t.Fatal("Errors must filter the list and preview the error location")
	}
	m.changesFocus = true
	press(t, m, tui.KeyUp)
	press(t, m, tui.KeyDown)
	if m.findingIndex != 1 {
		t.Fatal("arrows must stay within matching overview findings")
	}
	rect := m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	if m.findingIndex != 1 {
		t.Fatal("clicks must map filtered rows to original findings")
	}
	rect = m.filterButtons[1].Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	render(m, 160, 40)
	if m.findingFilter != "violation" || m.findingIndex != 0 || m.preview.cursor.line != 1 {
		t.Fatal("clicking Violations must update selection and preview")
	}
	pressRune(t, m, 'a')
	render(m, 160, 40)
	if len(m.overviewList.Children()) != 2 {
		t.Fatal("All must restore both findings")
	}
	pressRune(t, m, 'e')
	m.BeginAnalysis(2)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 2, Report: analysis.Report{
		Files: []analysis.Result{{Path: "main.go", Findings: []analysis.Finding{
			{Path: "main.go", Rule: "test", Line: 2, Column: 1, Kind: "violation"},
		}}},
	}})
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "No findings match this filter.") || m.overviewPath != "" ||
		strings.Contains(frame, "var first = 1") {
		t.Fatal("resolved filtered results must clear stale preview and show the empty state")
	}
}

func TestOverviewFilterControlsFitNarrowLayout(t *testing.T) {
	m := overviewFixture(t)
	render(m, 100, 24)
	if m.overviewList.ContentRect().Height < 1 {
		t.Fatal("filters must leave visible finding rows in narrow layout")
	}
	for _, button := range m.filterButtons {
		if button.Rect().Y >= m.overviewList.Rect().Y {
			t.Fatal("filters must stay above the scrollable rows")
		}
	}
}

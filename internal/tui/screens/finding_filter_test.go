package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/grindlemire/go-tui"
)

func TestFindingFiltersAndSelection(t *testing.T) {
	m := overviewFixture(t)
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	pressRune(t, m, 'e')
	if indices := m.visibleFindings(); len(indices) != 1 || m.findingIndex != 1 {
		t.Fatal("Errors filter must select the error, not the first unfiltered finding")
	}
	m.changesFocus = true
	press(t, m, tui.KeyDown)
	if m.findingIndex != 1 {
		t.Fatal("navigation must stay within the filter")
	}
	press(t, m, tui.KeyEnter)
	render(m, 160, 40)
	if !m.contentTab || m.preview.cursor.line != 2 {
		t.Fatal("Enter must open the filtered error location")
	}
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	rect := m.filterButtons[1].Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	if m.findingFilter != "violation" || m.findingIndex != 0 {
		t.Fatal("click filter must select violations")
	}
	m.findings = nil
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "No findings match this filter") {
		t.Fatal("empty filtered results must be explicit")
	}
	m.toggle()
}

func TestFindingTableHorizontalScroll(t *testing.T) {
	m := overviewFixture(t)
	m.findings = []analysis.Finding{{Path: strings.Repeat("long/", 30) + "main.go",
		Rule: "last-rule", Kind: "violation", Message: "end-of-table-message"}}
	m.openWorkspace(violationsPage)
	render(m, 160, 40)
	maximum, _ := m.feed.MaxScroll()
	if maximum <= 0 {
		t.Fatal("table must overflow horizontally instead of truncating columns")
	}
	m.changesFocus = true
	for range 200 {
		pressRune(t, m, ']')
	}
	render(m, 160, 40)
	x, _ := m.feed.ScrollOffset()
	if x != maximum {
		t.Fatalf("horizontal scroll must clamp to end: %d != %d", x, maximum)
	}
	if len(m.feed.Children()) != 1 {
		t.Fatal("table must retain finding rows")
	}
	for range 200 {
		pressRune(t, m, '[')
	}
	if m.findingX != 0 {
		t.Fatal("Left must return to beginning")
	}
}

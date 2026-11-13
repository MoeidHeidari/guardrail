package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
)

func TestOverviewFindingsHorizontalScroll(t *testing.T) {
	m := overviewFixture(t)
	m.findings = []analysis.Finding{{
		Path: strings.Repeat("long-folder/", 12) + "main.go",
		Rule: "rule-at-end", Set: "quality", Line: 1,
	}}
	render(m, 160, 40)
	maximum, _ := m.overviewList.MaxScroll()
	if maximum <= 0 {
		t.Fatal("long finding rows must have horizontal overflow")
	}
	m.changesFocus = true
	pressRune(t, m, ']')
	render(m, 160, 40)
	x, _ := m.overviewList.ScrollOffset()
	if x != 4 || m.overviewListX != 4 {
		t.Fatal("Right must horizontally scroll the findings panel")
	}
	for range 100 {
		pressRune(t, m, ']')
	}
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "rule-at-end") {
		t.Fatal("horizontal scrolling must reveal the end of the full finding")
	}
	for range 100 {
		pressRune(t, m, '[')
	}
	if m.overviewListX != 0 || m.contentX != 0 {
		t.Fatal("Left must clamp to the beginning without scrolling the preview")
	}
}

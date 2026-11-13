package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func TestRulesTableColumnsAndHorizontalScrolling(t *testing.T) {
	m := overviewFixture(t)
	m.analysisResult.Rules = []config.Rule{{Set: "quality", ID: strings.Repeat("long-rule-", 7) + "end", Type: "regex"}}
	m.openWorkspace(rulesPage)
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "Rule set / Rule") || !strings.Contains(frame, "[-] quality") {
		t.Fatalf("rules table must show headings and grouped rows:\n%s", frame)
	}
	maximum, _ := m.rulesList.MaxScroll()
	if maximum == 0 {
		t.Fatal("long rules must overflow horizontally")
	}
	m.changesFocus = true
	for range 100 {
		pressRune(t, m, ']')
	}
	render(m, 160, 40)
	if m.rulesListX != maximum {
		t.Fatal("rules table must scroll horizontally")
	}
	header := m.rulesList.Parent().Children()[1]
	x, _ := header.ScrollOffset()
	if x != m.rulesListX {
		t.Fatal("table header and rows must remain aligned")
	}
	press(t, m, tui.KeyLeft)
	render(m, 160, 40)
	if len(m.rulesItems) != 1 {
		t.Fatal("Left must still collapse the selected set")
	}
}

func TestRulesTableImpactColors(t *testing.T) {
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		m := overviewFixture(t)
		m.theme = theme
		m.openWorkspace(rulesPage)
		render(m, 160, 40)
		set := m.rulesList.Children()[0].Children()
		if !set[0].TextStyle().Fg.Equal(theme.Accent) ||
			!set[4].TextStyle().Fg.Equal(theme.Modified) ||
			!set[5].TextStyle().Fg.Equal(theme.Danger) {
			t.Fatal("set names and nonzero impact counts must use theme colors")
		}
		m.findings = nil
		render(m, 160, 40)
		set = m.rulesList.Children()[0].Children()
		if !set[4].TextStyle().Fg.Equal(theme.Muted) || !set[5].TextStyle().Fg.Equal(theme.Muted) {
			t.Fatal("zero impact counts must remain muted")
		}
	}
}

func TestRuleImpactUniqueFilesAndSetIsolation(t *testing.T) {
	m := testScreen()
	m.BeginAnalysis(1)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 1,
		Rules: []config.Rule{
			{ID: "same", Set: "go", Type: "regex"},
			{ID: "other", Set: "go", Type: "javascript"},
			{ID: "same", Set: "python", Type: "regex"},
		},
		Report: analysis.Report{Files: []analysis.Result{{Findings: []analysis.Finding{
			{Path: "a.go", Set: "go", Rule: "same", Kind: "violation"},
			{Path: "a.go", Set: "go", Rule: "same", Kind: "violation"},
			{Path: "a.go", Set: "go", Rule: "same", Kind: "error"},
			{Path: "b.go", Set: "go", Rule: "other", Kind: "violation"},
			{Path: "a.py", Set: "python", Rule: "same", Kind: "error"},
		}}}}})
	rules, sets := m.ruleImpacts()
	first := rules[ruleKey{"go", "same"}]
	if len(first.files) != 1 || first.violations != 2 || first.errors != 1 ||
		len(first.violatedFiles) != 1 || len(first.errorFiles) != 1 {
		t.Fatalf("rule counts must distinguish occurrences and unique files: %+v", first)
	}
	if len(sets["go"].files) != 2 || sets["go"].violations != 3 ||
		sets["python"].violations != 0 || sets["python"].errors != 1 {
		t.Fatal("set counts must deduplicate files and isolate duplicate rule IDs")
	}
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		m.theme = theme
		m.openWorkspace(rulesPage)
		render(m, 160, 40)
		m.feed = tui.New(tui.WithDirection(tui.Column))
		m.renderRuleList()
		buffer := tui.NewBuffer(160, 80)
		m.feed.RenderTo(buffer, 160, 80)
		frame := buffer.StringTrimmed()
		for _, text := range []string{"[-] go          -          2     2     3          1",
			"same        regex      -     1     2          1", "[-] python"} {
			if !strings.Contains(frame, text) {
				t.Fatalf("missing impact text %q:\n%s", text, frame)
			}
		}
	}
	m.BeginAnalysis(2)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 2, Rules: m.analysisResult.Rules})
	_, sets = m.ruleImpacts()
	if len(sets) != 0 {
		t.Fatal("resolved findings must clear rule impacts")
	}
}

package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestStartupFocusesWorkspace(t *testing.T) {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{})
	if !m.navigationFocus || m.changesFocus || m.navigationIndex != 0 {
		t.Fatal("startup must focus Overview in workspace navigation")
	}
}

func TestRuleSetsAndRulesTree(t *testing.T) {
	m := overviewFixture(t)
	m.analysisResult.Rules = []config.Rule{
		{Set: "alpha", ID: "first", Type: "regex"},
		{Set: "beta", ID: "second", Type: "javascript"},
	}
	m.openWorkspace(rulesPage)
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "2 rule sets / 2 rules") {
		t.Fatal("rule tree heading must show total sets and rules")
	}
	selected := 0
	for _, card := range m.rulesList.Children() {
		if card.BorderStyle().Fg.Equal(m.theme.Accent) {
			selected++
		}
	}
	if selected != 1 || len(m.rulesItems) != 4 || m.rulesItems[1].id != "first" {
		t.Fatal("tree must show all sets and their children with a single selection")
	}
	m.changesFocus, m.rulesFocus = true, 0
	m.move(3)
	render(m, 160, 40)
	if m.selectedRuleSet != "beta" || m.rulesItems[m.rulesSelection].id != "second" {
		t.Fatal("tree navigation must reach rules in another set")
	}
	if m.rulesList.Rect().Height < 4 {
		t.Fatal("tree must retain usable list height")
	}
	press(t, m, tui.KeyTab)
	if m.rulesFocus != 1 {
		t.Fatal("Tab must move from tree to affected files")
	}
	m.rulesFocus, m.rulesSelection = 0, 0
	press(t, m, tui.KeyLeft)
	render(m, 160, 40)
	if len(m.rulesItems) != 3 || m.rulesItems[1].set != "beta" {
		t.Fatal("collapsing a set must hide only its child rules")
	}
	press(t, m, tui.KeyRight)
	render(m, 160, 40)
	if len(m.rulesItems) != 4 {
		t.Fatal("expanding a set must restore its rules")
	}
}

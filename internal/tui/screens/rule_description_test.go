package screens

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestSelectedRuleDescription(t *testing.T) {
	m := overviewFixture(t)
	m.analysisResult.Rules = []config.Rule{{
		ID: "test", Set: "quality", Type: "regex",
		Description: "Explain what this validation checks and why it matters.",
		Message:     "Short violation message.",
	}}
	m.rulesSelection = 1
	m.openWorkspace(rulesPage)
	frame := render(m, 160, 40)
	if !strings.Contains(frame, "test: Explain what this validation checks and why it matters.") {
		t.Fatalf("selected description must be visible outside narrow rule cards:\n%s", frame)
	}
	card := m.rulesList.Children()[m.rulesSelection]
	if !card.BorderStyle().Bg.Equal(m.theme.Background) {
		t.Fatal("selected rule border must retain the normal background")
	}
	if got := ruleDescription(config.Rule{Message: "Fallback explanation"}); got != "Fallback explanation" {
		t.Fatal("existing rule messages must remain visible as descriptions")
	}
}

package screens

import (
	"fmt"
	"sort"

	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/grindlemire/go-tui"
)

type ruleImpact struct {
	files         map[string]bool
	violatedFiles map[string]bool
	errorFiles    map[string]bool
	violations    int
	errors        int
}

type ruleKey struct{ set, id string }

func normalizedSet(set string) string {
	if set == "" {
		return "General"
	}
	return set
}

func (m *Main) ruleImpacts() (map[ruleKey]*ruleImpact, map[string]*ruleImpact) {
	rules, sets := make(map[ruleKey]*ruleImpact), make(map[string]*ruleImpact)
	for _, finding := range m.findings {
		set, _ := m.findingSource(finding)
		key := ruleKey{set, finding.Rule}
		for _, target := range []*ruleImpact{impactFor(rules, key), impactFor(sets, set)} {
			target.files[finding.Path] = true
			if finding.Kind == "error" {
				target.errors++
				target.errorFiles[finding.Path] = true
			} else {
				target.violations++
				target.violatedFiles[finding.Path] = true
			}
		}
	}
	return rules, sets
}

func impactFor[K comparable](impacts map[K]*ruleImpact, key K) *ruleImpact {
	if impacts[key] == nil {
		impacts[key] = &ruleImpact{files: make(map[string]bool),
			violatedFiles: make(map[string]bool), errorFiles: make(map[string]bool)}
	}
	return impacts[key]
}

func (m *Main) renderRuleList() {
	m.feed.Apply(tui.WithGap(0))
	ruleImpacts, setImpacts := m.ruleImpacts()
	grouped := make(map[string][]config.Rule)
	for _, rule := range m.analysisResult.Rules {
		set := normalizedSet(rule.Set)
		grouped[set] = append(grouped[set], rule)
	}
	names := make([]string, 0, len(grouped))
	for set := range grouped {
		names = append(names, set)
	}
	sort.Strings(names)
	m.rulesItems = nil
	for _, set := range names {
		m.rulesItems = append(m.rulesItems, ruleKey{set, ""})
		impact := impactFor(setImpacts, set)
		marker := "[-]"
		if m.collapsedRuleSets[set] {
			marker = "[+]"
		}
		m.feed.AddChild(m.label(display(fmt.Sprintf("%s %s / %d rules / %d files / %d violations / %d errors",
			marker, set, len(grouped[set]), len(impact.files), impact.violations, impact.errors))))
		if m.collapsedRuleSets[set] {
			continue
		}
		for _, rule := range grouped[set] {
			m.rulesItems = append(m.rulesItems, ruleKey{set, rule.ID})
			impact := impactFor(ruleImpacts, ruleKey{set, rule.ID})
			m.feed.AddChild(m.label(display(fmt.Sprintf("    %s  [%s] / %d files / %d violations / %d errors",
				rule.ID, rule.Type, len(impact.files), impact.violations, impact.errors))))
		}

	}
	if len(grouped) == 0 {
		m.feed.AddChild(m.label("Configure regex or JavaScript functions in .guard/rules.yml."))
	}
	m.rulesSelection = min(m.rulesSelection, max(0, len(m.rulesItems)-1))
	if len(m.rulesItems) > 0 {
		for i, row := range m.feed.Children() {
			color := m.theme.Background
			if i == m.rulesSelection {
				color = m.theme.Accent
			}
			row.Apply(tui.WithPaddingTRBL(0, 1, 0, 1),
				tui.WithBorderTRBL(tui.BorderNone, tui.BorderNone, tui.BorderNone, tui.BorderSingle),
				tui.WithBorderStyle(tui.Style{Fg: color, Bg: m.theme.Background}))
		}
		m.selectedRuleSet = m.rulesItems[m.rulesSelection].set
	}
}

func ruleDescription(rule config.Rule) string {
	if rule.Description != "" {
		return rule.Description
	}
	if rule.Message != "" {
		return rule.Message
	}
	return "No description provided for this rule."
}

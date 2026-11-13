package screens

import (
	"fmt"
	"sort"

	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
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
	var rows [][]string
	for _, set := range names {
		m.rulesItems = append(m.rulesItems, ruleKey{set, ""})
		impact := impactFor(setImpacts, set)
		marker := "[-]"
		if m.collapsedRuleSets[set] {
			marker = "[+]"
		}
		rows = append(rows, []string{marker + " " + set, "-", fmt.Sprint(len(grouped[set])),
			fmt.Sprint(len(impact.files)), fmt.Sprint(impact.violations), fmt.Sprint(impact.errors)})
		if m.collapsedRuleSets[set] {
			continue
		}
		for _, rule := range grouped[set] {
			m.rulesItems = append(m.rulesItems, ruleKey{set, rule.ID})
			impact := impactFor(ruleImpacts, ruleKey{set, rule.ID})
			rows = append(rows, []string{"    " + rule.ID, rule.Type, "-",
				fmt.Sprint(len(impact.files)), fmt.Sprint(impact.violations), fmt.Sprint(impact.errors)})
		}
	}
	m.ruleWidths = []int{15, 10, 5, 5, 10, 6}
	for _, row := range rows {
		for i, value := range row {
			m.ruleWidths[i] = max(m.ruleWidths[i], tui.StringWidth(display(value)))
		}
	}
	if len(grouped) == 0 {
		m.feed.AddChild(m.label("Configure regex or JavaScript functions in .guard/rules.yml."))
	}
	m.rulesSelection = min(m.rulesSelection, max(0, len(m.rulesItems)-1))
	if len(m.rulesItems) > 0 {
		for i, values := range rows {
			m.feed.AddChild(m.ruleTableRow(values, i == m.rulesSelection, false))
		}
		m.selectedRuleSet = m.rulesItems[m.rulesSelection].set
	}
}

func (m *Main) ruleTableRow(values []string, selected, heading bool) *tui.Element {
	theme := m.theme
	theme.Surface = theme.Background
	color := theme.Background
	if selected {
		color = theme.Accent
	}
	width := 3 + len(values) - 1
	for _, column := range m.ruleWidths {
		width += column
	}
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithGap(1),
		tui.WithHeight(1), tui.WithWidth(width), tui.WithFlexShrink(0),
		tui.WithPaddingTRBL(0, 1, 0, 1),
		tui.WithBorderTRBL(tui.BorderNone, tui.BorderNone, tui.BorderNone, tui.BorderSingle),
		tui.WithBorderStyle(tui.Style{Fg: color, Bg: theme.Background}))
	for i, value := range values {
		color := theme.Foreground
		if heading {
			color = theme.Muted
		} else if i == 0 && values[2] != "-" {
			color = theme.Accent
		} else if i == 1 || (i >= 3 && value == "0") {
			color = theme.Muted
		} else if i == 4 && value != "0" {
			color = theme.Modified
		} else if i == 5 && value != "0" {
			color = theme.Danger
		}
		style := tui.NewStyle().Foreground(color)
		if heading || (i == 0 && values[2] != "-") {
			style = style.Bold()
		}
		cell := m.label(display(value))
		cell.Apply(tui.WithWidth(m.ruleWidths[i]), tui.WithFlexShrink(0), tui.WithTextStyle(style))
		row.AddChild(cell)
	}
	return row
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

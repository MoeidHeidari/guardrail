package screens

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func (m *Main) overviewCard(title string, height int, rows ...*tui.Element) *tui.Element {
	theme := m.theme
	theme.Surface = theme.Background
	heading := m.label(title)
	heading.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(theme.Muted).Bold()))
	card := components.NewSection(theme, "", heading)
	card.AddChild(rows...)
	card.Apply(tui.WithHeight(height), tui.WithFlexShrink(1),
		tui.WithPaddingTRBL(0, 1, 0, 1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.Style{Fg: theme.Separator, Bg: theme.Background}))
	return card
}

func (m *Main) renderOverview() {
	m.feed.Apply(tui.WithGap(1))
	m.feed.AddChild(m.metricPanels())
	summary := m.label(fmt.Sprintf("%d files & folders   /   %d changed files", len(m.entries), m.changedFiles))
	if m.renderHeight < 30 {
		m.feed.AddChild(m.label(display(m.root) + " / " + m.analysisStatus()))
	} else {
		m.feed.AddChild(m.overviewCard("Workspace / Engineering status", 6, m.label(display(m.root)), summary, m.label(m.analysisStatus())))
	}
	m.renderOverviewWorkspace()
}

func (m *Main) metricPanels() *tui.Element {
	theme := m.theme
	theme.Surface = theme.Background
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithGap(1),
		tui.WithHeight(6), tui.WithFlexShrink(0))
	violations, failures, affected := 0, 0, make(map[string]bool)
	flagged := make(map[ruleKey]bool)
	for _, finding := range m.findings {
		affected[finding.Path] = true
		flagged[ruleKey{set: finding.Set, id: finding.Rule}] = true
		if finding.Kind == "error" {
			failures++
		} else {
			violations++
		}
	}
	flaggedRules := 0
	for _, rule := range m.analysisResult.Rules {
		if flagged[ruleKey{set: rule.Set, id: rule.ID}] || flagged[ruleKey{id: rule.ID}] {
			flaggedRules++
		}
	}
	width := max(1, (m.renderWidth-48)/4-4)
	for _, metric := range []struct {
		name                 string
		count, filled, total int
		detail               string
		color                tui.Color
	}{
		{"Violations", violations, violations, len(m.findings), "of all findings", m.theme.Modified},
		{"Errors", failures, failures, len(m.findings), "of all findings", m.theme.Danger},
		{"Affected files", len(affected), len(affected), m.analysisResult.Report.Checked, "of checked files", m.theme.Modified},
		{"Rules", len(m.analysisResult.Rules), flaggedRules, len(m.analysisResult.Rules), "with findings", m.theme.Accent},
	} {
		value := fmt.Sprintf("%d", metric.count)
		detail := fmt.Sprintf("%d/%d %s", metric.filled, metric.total, metric.detail)
		if m.analysisGeneration == 0 || m.analysisResult.Error != "" {
			value, detail = "--", "Awaiting analysis"
			if m.analysisResult.Error != "" {
				detail = "Analysis unavailable"
			}
		} else if m.analysisPending {
			detail = "Updating..."
		}
		number := m.label(value)
		number.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(metric.color).Bold()))
		bar := m.metricBar(metric.filled, metric.total, width, metric.color)
		if value == "--" {
			bar = m.metricBar(0, 0, width, metric.color)
		}
		caption := m.label(detail)
		caption.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
		card := m.overviewCard(metric.name, 6, number, bar, caption)
		card.Apply(tui.WithWidth(0), tui.WithFlexGrow(1), tui.WithMinWidth(0))
		row.AddChild(card)
	}
	return row
}

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
		tui.WithHeight(4), tui.WithFlexShrink(0))
	violations, failures, affected := 0, 0, make(map[string]bool)
	for _, finding := range m.findings {
		affected[finding.Path] = true
		if finding.Kind == "error" {
			failures++
		} else {
			violations++
		}
	}
	for _, metric := range []struct {
		name  string
		count int
	}{
		{"Violations", violations}, {"Errors", failures}, {"Affected files", len(affected)},
		{"Rules", len(m.analysisResult.Rules)},
	} {
		value := fmt.Sprint(metric.count)
		if m.analysisGeneration == 0 {
			value = "--"
		}
		card := m.overviewCard(metric.name, 4, m.label(value))
		color := m.theme.Accent
		if metric.name == "Violations" && metric.count > 0 {
			color = m.theme.Modified
		} else if metric.name == "Errors" && metric.count > 0 {
			color = m.theme.Danger
		}
		card.Children()[1].Apply(tui.WithTextStyle(tui.NewStyle().Foreground(color).Bold()))
		card.Apply(tui.WithWidth(0), tui.WithFlexGrow(1), tui.WithMinWidth(0))
		row.AddChild(card)
	}
	return row
}

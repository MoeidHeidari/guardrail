package screens

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func findingLocation(finding analysis.Finding) string {
	if finding.Line > 0 {
		return fmt.Sprintf("%s:%d:%d", finding.Path, finding.Line, finding.Column)
	}

	return finding.Path + " / file-level"
}

func (m *Main) measureFindingColumns() {
	m.tableWidths = []int{9, 24, 12, 18, 24}
	for _, finding := range m.findings {
		set, _ := m.findingSource(finding)
		for i, value := range []string{finding.Kind, findingLocation(finding), set, finding.Rule, finding.Message} {
			m.tableWidths[i] = max(m.tableWidths[i], tui.StringWidth(display(value)))
		}
	}
}

func (m *Main) findingTableRow(values []string, selected, heading bool) *tui.Element {
	theme := m.theme
	theme.Surface = theme.Background
	background := theme.Background
	if selected {
		background = theme.Selection
	}
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithGap(1),
		tui.WithHeight(1), tui.WithFlexShrink(0), tui.WithPaddingTRBL(0, 1, 0, 1),
		tui.WithBackground(tui.Style{Bg: background}))
	width := 6
	for _, column := range m.tableWidths {
		width += column
	}
	row.Apply(tui.WithWidth(width), tui.WithFlexShrink(0))
	for i, value := range values {
		cell := m.label(display(value))
		color := theme.Foreground
		if heading {
			color = theme.Muted
		} else if i == 0 {
			color = theme.Modified
			if value == "error" {
				color = theme.Danger
			}
		} else if selected {
			color = theme.Accent
		}
		style := tui.NewStyle().Foreground(color).Background(background)
		if heading {
			style = style.Bold()
		}
		cell.Apply(tui.WithMinWidth(0), tui.WithTextStyle(style),
			tui.WithBackground(tui.Style{Bg: background}))
		column := 10
		if i < len(m.tableWidths) {
			column = m.tableWidths[i]
		}
		cell.Apply(tui.WithWidth(column), tui.WithFlexShrink(0), tui.WithWrap(false))
		row.AddChild(cell)
	}
	return row
}

func (m *Main) findingDetails() *tui.Element {
	if len(m.visibleFindings()) == 0 {
		return m.label("")
	}
	finding := m.findings[m.findingIndex]
	set, source := m.findingSource(finding)
	if m.renderHeight < 30 {
		theme := m.theme
		theme.Surface = theme.Background
		panel := components.NewSection(theme, "", m.label(display(findingLocation(finding)+" ["+finding.Kind+"]")),
			m.label(display("Rule set: "+set+" / Rule: "+finding.Rule)),
			m.label(display(finding.Message)), m.label(display("Source: "+source)))
		panel.Apply(tui.WithHeight(4), tui.WithFlexShrink(0))
		return panel
	}
	card := m.overviewCard(display(findingLocation(finding)+" ["+finding.Kind+"]"), 6,
		m.label(display("Rule set: "+set+" / Rule: "+finding.Rule)),
		m.label(display(finding.Message)),
		m.label(display("Source: "+source)))
	card.Apply(tui.WithFlexShrink(0))
	return card
}

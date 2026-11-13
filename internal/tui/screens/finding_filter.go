package screens

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

var findingFilters = []string{"", "violation", "error"}

func (m *Main) handleFilterMouse(event tui.MouseEvent) bool {
	if (!m.findingsVisible() && !m.overviewVisible()) || event.Button != tui.MouseLeft {
		return false
	}
	for i, button := range m.filterButtons {
		if button.Rect().Contains(event.X, event.Y) {
			m.setFindingFilter(findingFilters[i])
			return true
		}
	}
	return false
}

func (m *Main) visibleFindings() []int {
	var indices []int
	for i, finding := range m.findings {
		if m.findingFilter == "" || finding.Kind == m.findingFilter {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m *Main) setFindingFilter(filter string) {
	m.findingFilter, m.findingOffset = filter, 0
	indices := m.visibleFindings()
	if len(indices) > 0 {
		m.findingIndex = indices[0]
	}
	if m.overviewVisible() {
		m.selectOverviewFinding()
	}
}

func (m *Main) moveFinding(delta int) {
	indices := m.visibleFindings()
	if len(indices) == 0 {
		return
	}
	position := 0
	for i, index := range indices {
		if index == m.findingIndex {
			position = i
			break
		}
	}
	position = max(0, min(position+delta, len(indices)-1))
	m.findingIndex = indices[position]
	pane := m.feed
	if m.overviewVisible() {
		pane = m.overviewList
	}
	if position < m.findingOffset {
		m.findingOffset = position
	} else if position >= m.findingOffset+viewport(pane) {
		m.findingOffset = position - viewport(pane) + 1
	}
	m.findingOffset = max(0, min(m.findingOffset, max(0, len(indices)-viewport(pane))))
}

func (m *Main) renderFindingFilters() *tui.Element {
	theme := m.theme
	theme.Surface = theme.Background
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithGap(2),
		tui.WithHeight(1), tui.WithFlexShrink(0))
	m.filterButtons = nil
	for i, label := range []string{"All [a]", "Violations [b]", "Errors [e]"} {
		button := m.label(label)
		button.Apply(tui.WithWidth(len(label)), tui.WithFlexShrink(0))
		if m.findingFilter == findingFilters[i] {
			button.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(theme.Accent).Bold()),
				tui.WithBackground(tui.Style{Bg: theme.Selection}))
		}
		row.AddChild(button)
		m.filterButtons = append(m.filterButtons, button)
	}
	return row
}

func (m *Main) scrollFindingX(delta int) bool {
	if !m.findingsVisible() || !m.changesFocus || m.feed == nil {
		return false
	}
	scrollHorizontal(m.feed, &m.findingX, delta)
	return true
}

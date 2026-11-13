package screens

import (
	"fmt"
	"path/filepath"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func (m *Main) rulesVisible() bool {
	return m.dashboardVisible() && m.workspacePage == rulesPage
}

func (m *Main) activateRulesPane() bool {
	if !m.rulesVisible() || !m.changesFocus {
		return false
	}
	if m.rulesFocus == 1 {
		if len(m.rulesMatches) == 0 {
			return true
		}
		m.rulesPreviewOpened = true
		m.rulesPath = ""
		m.updateRuleFiles()
	}
	m.rulesFocus = min(2, m.rulesFocus+1)
	return true
}

func (m *Main) updateRuleFiles() {
	m.rulesMatches = nil
	if len(m.rulesItems) == 0 {
		m.rulesPath = ""
		return
	}
	key := m.rulesItems[m.rulesSelection]
	for i, finding := range m.findings {
		set, _ := m.findingSource(finding)
		if set == key.set && (key.id == "" || key.id == finding.Rule) {
			m.rulesMatches = append(m.rulesMatches, i)
		}
	}
	m.rulesFileSelection = min(m.rulesFileSelection, max(0, len(m.rulesMatches)-1))
	if len(m.rulesMatches) == 0 {
		m.rulesPath = ""
		return
	}
	finding := m.findings[m.rulesMatches[m.rulesFileSelection]]
	path := filepath.FromSlash(finding.Path)
	if path != m.rulesPath {
		m.rulesPath, m.previewDirty = path, true
		m.jumpLine, m.jumpColumn = max(1, finding.Line), max(1, finding.Column)
	}
}

func (m *Main) renderRules() {
	theme := m.theme
	theme.Surface = theme.Background
	parent := m.feed
	parent.Apply(tui.WithScrollable(tui.ScrollNone), tui.WithGap(1))
	m.rulesList = components.NewContainer(theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(m.rulesListX, m.findingOffset))
	m.styleScrollbar(m.rulesList)
	m.feed = m.rulesList
	m.renderRuleList()
	m.feed = parent
	m.updateRuleFiles()
	description := "Select a rule to see its description."
	if len(m.rulesItems) == 0 {
		description = "Configure regex or JavaScript functions in .guard/rules.yml."
	}
	if len(m.rulesItems) > 0 {
		key := m.rulesItems[m.rulesSelection]
		description = "Rule set [" + key.set + "]: select a rule below to see its validation description."
		for _, rule := range m.analysisResult.Rules {
			if normalizedSet(rule.Set) == key.set && rule.ID == key.id {
				description = rule.ID + ": " + ruleDescription(rule)
				break
			}
		}
	}
	descriptionText := m.label(display(description))
	descriptionText.Apply(tui.WithHeight(2), tui.WithWrap(true),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Muted)))
	if m.renderHeight < 30 {
		descriptionText.Apply(tui.WithHeight(1))
	}
	parent.AddChild(descriptionText)
	m.rulesFiles = components.NewContainer(theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth))
	m.styleScrollbar(m.rulesFiles)
	for i, index := range m.rulesMatches {
		finding := m.findings[index]
		text := display(fmt.Sprintf("%s:%d [%s] %s", finding.Path, finding.Line, finding.Kind, finding.Rule))
		row := m.label(text)
		row.Apply(tui.WithWidth(tui.StringWidth(text)), tui.WithFlexShrink(0))
		if i == m.rulesFileSelection {
			row.Apply(tui.WithBackground(tui.Style{Bg: theme.Selection}),
				tui.WithTextStyle(tui.Style{Fg: theme.Accent, Bg: theme.Selection}))
		}
		m.rulesFiles.AddChild(row)
	}
	if len(m.rulesMatches) == 0 {
		m.rulesFiles.AddChild(m.label("No findings for this rule or set."))
	}
	m.rulesFiles.Apply(tui.WithScrollOffset(m.rulesFilesX, max(0, m.rulesFileSelection-max(1, m.renderHeight/4))))
	sets := make(map[string]bool)
	for _, rule := range m.analysisResult.Rules {
		sets[normalizedSet(rule.Set)] = true
	}
	header := components.NewContainer(theme, tui.WithHeight(1), tui.WithFlexShrink(0),
		tui.WithScrollable(tui.ScrollHorizontal), tui.WithScrollOffset(m.rulesListX, 0))
	header.AddChild(m.ruleTableRow([]string{"Rule set / Rule", "Type", "Rules", "Files", "Violations", "Errors"}, false, true))
	left := m.overviewCard(fmt.Sprintf("%d rule sets / %d rules", len(sets), len(m.analysisResult.Rules)), 0, header, m.rulesList)
	right := m.overviewCard("Affected files / Findings", 0, m.rulesFiles)
	for _, card := range []*tui.Element{left, right} {
		card.Apply(tui.WithHeightPercent(100), tui.WithWidth(0), tui.WithFlexGrow(1), tui.WithMinWidth(0))
	}
	direction := tui.Row
	if m.renderWidth < 100 {
		direction = tui.Column
		left.Apply(tui.WithHeight(0), tui.WithMinHeight(4), tui.WithFlexGrow(1))
		right.Apply(tui.WithHeight(0))
	}
	lists := components.NewContainer(theme, tui.WithDirection(direction), tui.WithGap(1),
		tui.WithHeight(0), tui.WithFlexGrow(1), tui.WithMinHeight(0))
	lists.AddChild(left, right)
	if !m.rulesPreviewOpened {
		m.rulesPreview = nil
		m.rulesClose = nil
		parent.AddChild(lists)
		return
	}
	m.rulesPreview = components.NewContainer(theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(m.contentX, m.contentOffset))
	m.styleScrollbar(m.rulesPreview)
	m.feed = m.rulesPreview
	if m.rulesPath != "" {
		if m.previewRequests != nil {
			m.requestPreview()
		} else {
			m.loadPreview()
		}
		m.renderPreview()
	} else {
		m.feed.AddChild(m.label("Select a rule, then a file to preview its findings."))
	}
	m.feed = parent
	title := "Live content"
	if m.rulesPath != "" {
		title += " / " + display(m.rulesPath)
	}
	preview := m.previewCard(title, max(5, (m.renderHeight-10)/2), m.rulesPreview, &m.rulesClose)
	preview.Apply(tui.WithFlexShrink(0))
	parent.AddChild(lists, preview)
}

func (m *Main) moveRules(delta int) bool {
	if !m.rulesVisible() || !m.changesFocus {
		return false
	}
	switch m.rulesFocus {
	case 0:
		m.rulesSelection = max(0, min(m.rulesSelection+delta, len(m.rulesItems)-1))
		m.rulesFileSelection = 0
		if m.rulesSelection < m.findingOffset {
			m.findingOffset = m.rulesSelection
		} else if m.rulesSelection >= m.findingOffset+viewport(m.rulesList) {
			m.findingOffset = m.rulesSelection - viewport(m.rulesList) + 1
		}
		m.updateRuleFiles()
	case 1:
		if len(m.rulesMatches) > 0 {
			m.rulesPreviewOpened = true
		}
		m.rulesFileSelection = max(0, min(m.rulesFileSelection+delta, len(m.rulesMatches)-1))
		m.rulesPath = ""
		m.updateRuleFiles()
	case 2:
		m.contentOffset = max(0, min(m.contentOffset+delta, len(m.preview.lines)-viewport(m.rulesPreview)))
	}
	return true
}

func (m *Main) handleRulesMouse(event tui.MouseEvent) bool {
	if !m.rulesVisible() {
		return false
	}
	for i, pane := range []*tui.Element{m.rulesList, m.rulesFiles, m.rulesPreview} {
		if pane == nil || !pane.Rect().Contains(event.X, event.Y) {
			continue
		}
		m.rulesFocus, m.changesFocus, m.watchFocus, m.navigationFocus = i, true, false, false
		switch event.Button {
		case tui.MouseWheelDown:
			m.moveRules(3)
		case tui.MouseWheelUp:
			m.moveRules(-3)
		case tui.MouseLeft:
			_, offset := pane.ScrollOffset()
			y := event.Y - pane.ContentRect().Y + offset
			if i == 1 && y >= 0 && y < len(m.rulesMatches) {
				m.rulesPreviewOpened = true
				m.rulesFileSelection, m.rulesPath = y, ""
				m.updateRuleFiles()
			} else if i == 0 {
				if y >= 0 && y < len(m.rulesItems) {
					m.rulesSelection, m.rulesFileSelection = y, 0
					m.updateRuleFiles()
				}
			}
		}
		return true
	}
	return false
}

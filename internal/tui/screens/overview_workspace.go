package screens

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func (m *Main) overviewVisible() bool {
	return m.dashboardVisible() && m.workspacePage == overviewPage
}

func (m *Main) previewEntry() (watcher.Entry, bool) {
	if m.rulesVisible() && m.rulesPath != "" {
		i, exists := m.entryPositions[m.rulesPath]
		if !exists {
			return watcher.Entry{}, false
		}
		return m.entries[i], true
	}
	if m.overviewVisible() && m.overviewPath != "" {
		i, exists := m.entryPositions[m.overviewPath]
		if !exists {
			return watcher.Entry{}, false
		}
		return m.entries[i], true
	}
	if len(m.rows) == 0 {
		return watcher.Entry{}, false
	}
	return m.rows[m.selected].entry, true
}

func (m *Main) selectOverviewFinding() {
	if !m.overviewPreviewOpened {
		return
	}
	indices := m.visibleFindings()
	if len(indices) == 0 {
		if m.overviewPath != "" {
			m.overviewPath, m.previewDirty = "", true
		}
		return
	}
	valid := false
	for _, index := range indices {
		valid = valid || index == m.findingIndex
	}
	if !valid {
		m.findingIndex = indices[0]
	}
	finding := m.findings[m.findingIndex]
	path := filepath.FromSlash(finding.Path)
	line, column := max(1, finding.Line), max(1, finding.Column)
	if m.overviewPath != path || m.overviewLine != line || m.overviewColumn != column {
		m.overviewPath, m.previewDirty = path, true
		m.overviewLine, m.overviewColumn = line, column
		m.jumpLine, m.jumpColumn = line, column
	}
}

func (m *Main) renderOverviewWorkspace() {
	theme := m.theme
	theme.Surface = theme.Background
	direction := tui.Row
	if m.renderWidth > 0 && m.renderWidth < 130 {
		direction = tui.Column
	}
	body := components.NewContainer(theme, tui.WithDirection(direction), tui.WithGap(1),
		tui.WithHeight(0), tui.WithFlexGrow(1), tui.WithMinHeight(0))
	summaries := components.NewContainer(theme, tui.WithWidthPercent(40), tui.WithFlexGrow(1),
		tui.WithHeightPercent(100), tui.WithMinWidth(0), tui.WithGap(1))
	if direction == tui.Column && m.overviewPreviewOpened {
		summaries.Apply(tui.WithWidthPercent(100), tui.WithHeight(7), tui.WithFlexGrow(0), tui.WithFlexShrink(0))
	}
	m.overviewList = components.NewContainer(theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(m.overviewListX, m.findingOffset))
	m.styleScrollbar(m.overviewList)
	indices := m.visibleFindings()
	m.selectOverviewFinding()
	start := min(m.findingOffset, len(indices))
	end := min(len(indices), start+max(30, m.renderHeight))
	spacer := func(height int) {
		if height > 0 {
			m.overviewList.AddChild(tui.New(tui.WithHeight(height), tui.WithFlexShrink(0)))
		}
	}
	spacer(start)
	for i := start; i < end; i++ {
		index := indices[i]
		finding := m.findings[index]
		set, _ := m.findingSource(finding)
		row := m.label("")
		background, foreground := theme.Background, theme.Foreground
		if index == m.findingIndex {
			background, foreground = theme.Selection, theme.Accent
		}
		row.Apply(tui.WithBackground(tui.Style{Bg: background}),
			tui.WithRichText(
				tui.TextSpan{Text: display(fmt.Sprintf("%s:%d  ", finding.Path, finding.Line)),
					Style: tui.Style{Fg: foreground, Bg: background}},
				tui.TextSpan{Text: "[" + display(set) + "]",
					Style: tui.NewStyle().Foreground(theme.Accent).Background(theme.Selection).Bold()},
				tui.TextSpan{Text: " " + display(finding.Rule),
					Style: tui.Style{Fg: theme.Muted, Bg: background}},
			))
		width := 0
		for _, span := range row.RichText() {
			width += tui.StringWidth(span.Text)
		}
		row.Apply(tui.WithWidth(width), tui.WithWrap(false))
		m.overviewList.AddChild(row)
	}
	spacer(len(indices) - end)
	if len(indices) == 0 {
		text := "No findings to preview."
		if m.findingFilter != "" {
			text = "No findings match this filter."
		}
		m.overviewList.AddChild(m.label(text))
	}
	list := m.overviewCard("Affected files / Findings", 0, m.renderFindingFilters(), m.overviewList)
	list.Apply(tui.WithHeight(0), tui.WithFlexGrow(1), tui.WithMinHeight(0))
	summaries.AddChild(list)
	sets := make(map[string]int)
	for _, rule := range m.analysisResult.Rules {
		sets[rule.Set]++
	}
	names := make([]string, 0, len(sets))
	for name := range sets {
		names = append(names, name)
	}
	sort.Strings(names)
	var rules []*tui.Element
	for _, name := range names[:min(len(names), 3)] {
		count := sets[name]
		if name == "" {
			name = "General"
		}
		rules = append(rules, m.label(display(fmt.Sprintf("%s / %d rules", name, count))))
	}
	if len(rules) == 0 {
		rules = append(rules, m.label("No configured rules."))
	}
	if direction == tui.Row {
		summaries.AddChild(m.overviewCard("Rule sets", 3+len(rules), rules...))
	}
	if !m.overviewPreviewOpened {
		m.overviewPreview = nil
		m.overviewClose = nil
		summaries.Apply(tui.WithWidthPercent(100))
		body.AddChild(summaries)
		m.feed.AddChild(body)
		return
	}
	m.overviewPreview = components.NewContainer(theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(m.contentX, m.contentOffset))
	m.styleScrollbar(m.overviewPreview)
	parent := m.feed
	m.feed = m.overviewPreview
	if m.previewRequests != nil {
		if m.overviewPath != "" {
			m.requestPreview()
		}
	} else {
		if m.overviewPath != "" {
			m.loadPreview()
		}
	}
	if m.overviewPath == "" {
		m.feed.AddChild(m.label("Select a finding to preview its saved content."))
	} else {
		m.renderPreview()
	}
	m.feed = parent
	title := "Live content"
	if m.overviewPath != "" {
		title += " / " + display(m.overviewPath)
	}
	preview := m.previewCard(title, 0, m.overviewPreview, &m.overviewClose)
	preview.Apply(tui.WithHeight(0), tui.WithFlexGrow(1), tui.WithMinHeight(0), tui.WithMinWidth(0),
		tui.WithWidthPercent(60))
	if direction == tui.Column {
		preview.Apply(tui.WithWidthPercent(100))
	} else {
		preview.Apply(tui.WithHeightPercent(100))
	}
	body.AddChild(summaries, preview)
	m.feed.AddChild(body)
}

func (m *Main) handleOverviewMouse(event tui.MouseEvent) bool {
	if m.overviewVisible() && m.overviewPreview != nil && m.overviewPreview.Rect().Contains(event.X, event.Y) {
		m.overviewPreviewFocus, m.changesFocus, m.watchFocus, m.navigationFocus = true, true, false, false
		switch event.Button {
		case tui.MouseWheelDown:
			m.moveOverview(3)
		case tui.MouseWheelUp:
			m.moveOverview(-3)
		}
		return true
	}
	if !m.overviewVisible() || m.overviewList == nil || !m.overviewList.Rect().Contains(event.X, event.Y) {
		return false
	}
	m.changesFocus, m.navigationFocus, m.watchFocus = true, false, false
	m.overviewPreviewFocus = false
	switch event.Button {
	case tui.MouseLeft:
		_, offset := m.overviewList.ScrollOffset()
		index := event.Y - m.overviewList.ContentRect().Y + offset
		indices := m.visibleFindings()
		if index >= 0 && index < len(indices) {
			m.findingIndex = indices[index]
			m.overviewPreviewOpened = true
			m.selectOverviewFinding()
		}
	case tui.MouseWheelUp:
		m.move(-3)
	case tui.MouseWheelDown:
		m.move(3)
	}
	return true
}

func (m *Main) moveOverview(delta int) bool {
	if !m.overviewVisible() {
		return false
	}

	if m.overviewPreviewFocus {
		m.contentOffset = max(0, min(m.contentOffset+delta, len(m.preview.lines)-viewport(m.overviewPreview)))
	} else {
		m.moveFinding(delta)
		if len(m.visibleFindings()) > 0 {
			m.overviewPreviewOpened = true
		}
		m.selectOverviewFinding()
	}
	return true
}

func (m *Main) scrollOverviewX(delta int) bool {
	if m.rulesVisible() && m.changesFocus && m.rulesFocus == 0 {
		m.expandRuleTree(delta > 0)
		return true
	}
	if m.rulesVisible() && m.changesFocus && m.rulesFocus == 2 {
		feed := m.feed
		m.feed = m.rulesPreview
		m.scrollContentX(delta)
		m.feed = feed
		return true
	}
	if !m.overviewVisible() || !m.changesFocus {
		return false
	}
	if m.overviewPreviewFocus {
		feed := m.feed
		m.feed = m.overviewPreview
		m.scrollContentX(delta)
		m.feed = feed
	} else if m.overviewList != nil {
		scrollHorizontal(m.overviewList, &m.overviewListX, delta)
	}
	return true
}

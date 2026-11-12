package screens

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func (m *Main) dashboardVisible() bool {
	return !m.contentTab
}

func (m *Main) renderPanel() *tui.Element {
	if m.contentTab {
		if m.previewRequests != nil {
			m.requestPreview()
		} else {
			m.loadPreview()
		}
	}
	offset, offsetX, scrollMode := m.changeOffset, 0, tui.ScrollVertical
	if m.contentTab {
		offset, offsetX, scrollMode = m.contentOffset, m.contentX, tui.ScrollBoth
	} else if m.dashboardVisible() {
		offset, scrollMode = 0, tui.ScrollNone
		if m.workspacePage != overviewPage {
			offset, scrollMode = m.findingOffset, tui.ScrollVertical
			if m.findingsVisible() {
				offsetX, scrollMode = m.findingX, tui.ScrollBoth
			}
		}
	}
	editorTheme := m.theme
	editorTheme.Surface = m.theme.Background
	m.feed = components.NewContainer(editorTheme, tui.WithFlexGrow(1),
		tui.WithHeight(0), tui.WithScrollable(scrollMode),
		tui.WithScrollOffset(offsetX, offset))
	m.styleScrollbar(m.feed)
	if m.findingsVisible() {
		m.measureFindingColumns()
	}
	title := fmt.Sprintf("%s   /   %d events", display(m.root), len(m.changes))
	pageLabel := workspaceLabels[m.workspacePage]
	if m.contentOpened || m.contentTab {
		pageLabel = "0 " + pageLabel
	}
	m.dashboardButton = m.label(pageLabel)
	m.contentTabButton = nil
	active := m.dashboardButton
	switch {
	case m.contentTab:
		m.contentOpened = true
		m.contentTabButton = m.label("2 Content")
		active = m.contentTabButton
		title = display(m.preview.path)
		if m.preview.language != "" {
			title += "   [" + m.preview.language + "]"
		}
		if len(m.rows) > 0 && !m.rows[m.selected].entry.IsDir() {
			title += fmt.Sprintf("   %d bytes", m.rows[m.selected].entry.Size)
		}
		m.renderPreview()
	case m.dashboardVisible():
		active, title = m.dashboardButton, workspaceLabels[m.workspacePage]
		m.renderWorkspacePage()
	}
	if m.contentOpened && m.contentTabButton == nil {
		m.contentTabButton = m.label("2 Content")
	}
	for _, tab := range []*tui.Element{m.dashboardButton, m.contentTabButton} {
		if tab == nil {
			continue
		}
		tab.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
	}
	if m.contentOpened {
		active.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Underline()))
	} else {
		active.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
	}
	tabs := components.NewContainer(editorTheme, tui.WithDirection(tui.Row), tui.WithGap(2),
		tui.WithHeight(1), tui.WithFlexShrink(0))
	tabs.AddChild(m.dashboardButton)
	if m.contentTabButton != nil {
		tabs.AddChild(m.contentTabButton)
	}
	content := components.NewSection(editorTheme, "", tabs)
	if m.contentTab {
		subtitle := m.label(title)
		subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
		if m.changesFocus {
			subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
		}
		content.AddChild(subtitle)
	}
	if m.findingsVisible() {
		content.AddChild(m.metricPanels())
		content.AddChild(m.label(m.analysisStatus()), m.label("Enter or click a finding to open its file."))
		content.AddChild(m.renderFindingFilters())
		header := components.NewContainer(editorTheme, tui.WithHeight(1), tui.WithFlexShrink(0),
			tui.WithScrollable(tui.ScrollHorizontal), tui.WithScrollOffset(m.findingX, 0))
		header.AddChild(m.findingTableRow([]string{"Kind", "File / location", "Rule set", "Rule", "Message"}, false, true))
		content.AddChild(header)
	}
	content.AddChild(m.feed)
	if m.findingsVisible() && len(m.visibleFindings()) > 0 {
		content.AddChild(m.findingDetails())
	}
	if !m.dashboardVisible() {
		m.feed.Apply(tui.WithBorder(tui.BorderRounded),
			tui.WithBorderStyle(tui.Style{Fg: m.theme.Separator, Bg: m.theme.Background}))
	}
	content.Apply(tui.WithPaddingTRBL(1, 2, 1, 2))
	return content
}

func (m *Main) renderDashboard() {
	m.renderOverview()
}

func (m *Main) renderWorkspacePage() {
	switch m.workspacePage {
	case overviewPage:
		m.renderDashboard()
	case violationsPage:
		m.renderFindings()
	case rulesPage:
		m.renderRules()
	}
}

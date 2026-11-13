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
	if m.contentTab && m.contentActive == "" && len(m.rows) > 0 {
		m.openTab(m.rows[m.selected].entry.Path)
	}
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
	active := m.dashboardButton
	switch {
	case m.contentTab:
		m.contentOpened = true
		title = display(m.preview.path)
		if m.preview.language != "" {
			title += "   [" + m.preview.language + "]"
		}
		if entry, exists := m.previewEntry(); exists && !entry.IsDir() {
			title += fmt.Sprintf("   %d bytes", entry.Size)
		}
		m.renderPreview()
	case m.dashboardVisible():
		active, title = m.dashboardButton, workspaceLabels[m.workspacePage]
		m.renderWorkspacePage()
	}
	m.dashboardButton.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
	if m.contentOpened && !m.contentTab {
		active.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Underline()))
	} else if !m.contentOpened {
		active.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
	}
	m.highlightFocus(m.dashboardButton, m.changesFocus && !m.contentTab)
	tabs := components.NewContainer(editorTheme, tui.WithDirection(tui.Row), tui.WithGap(2),
		tui.WithHeight(1), tui.WithFlexShrink(0))
	tabs.AddChild(m.dashboardButton)
	m.renderContentTabs(tabs)
	content := components.NewSection(editorTheme, "", tabs)
	if m.contentTab {
		subtitle := m.label(title)
		subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
		if m.changesFocus {
			subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
		}
		m.highlightFocus(subtitle, m.changesFocus)
		content.AddChild(subtitle)
	}
	if m.findingsVisible() {
		content.AddChild(m.metricPanels())
		if m.analysisPending || m.analysisResult.Error != "" || m.analysisGeneration == 0 {
			content.AddChild(m.label(m.analysisStatus()))
		}
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

func (m *Main) previewCard(title string, height int, content *tui.Element, closeButton **tui.Element) *tui.Element {
	heading := m.label(title)
	heading.Apply(tui.WithWidth(0), tui.WithFlexGrow(1), tui.WithMinWidth(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted).Bold()))
	m.highlightFocus(heading, m.changesFocus && ((m.overviewVisible() && m.overviewPreviewFocus) ||
		(m.rulesVisible() && m.rulesFocus == 2)))
	button := m.label("[x]")
	button.Apply(tui.WithWidth(3), tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
	theme := m.theme
	theme.Surface = theme.Background
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithHeight(1), tui.WithFlexShrink(0))
	row.AddChild(heading, button)
	card := components.NewSection(theme, "", row, content)
	card.Apply(tui.WithHeight(height), tui.WithFlexShrink(1),
		tui.WithPaddingTRBL(0, 1, 0, 1), tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.Style{Fg: theme.Separator, Bg: theme.Background}))
	*closeButton = button
	return card
}

func (m *Main) closeOverviewPreview() {
	m.overviewPreviewOpened, m.overviewPreviewFocus = false, false
	m.overviewPreview, m.overviewClose = nil, nil
	m.overviewPath = ""
	m.invalidatePreview()
}

func (m *Main) closeRulesPreview() {
	m.rulesPreviewOpened = false
	m.rulesPreview, m.rulesClose = nil, nil
	m.rulesPath = ""
	if m.rulesFocus == 2 {
		m.rulesFocus = 1
	}
	m.invalidatePreview()
}

func (m *Main) invalidatePreview() {
	m.previewID++
	m.requestedPath = ""
	m.previewDirty = true
	m.jumpLine, m.jumpColumn = 0, 0
	m.dragPane = nil
}

func (m *Main) closeActivePreview() {
	switch {
	case m.contentTab:
		m.closeContent()
	case m.overviewVisible() && m.overviewPreviewOpened:
		m.closeOverviewPreview()
	case m.rulesVisible() && m.rulesPreviewOpened:
		m.closeRulesPreview()
	}
}

func (m *Main) handlePreviewClose(event tui.MouseEvent) bool {
	if event.Button != tui.MouseLeft {
		return false
	}
	switch {
	case m.contentOpened && m.handleContentTabMouse(event):
	case m.overviewVisible() && m.overviewPreviewOpened && m.overviewClose != nil && m.overviewClose.Rect().Contains(event.X, event.Y):
		m.closeOverviewPreview()
	case m.rulesVisible() && m.rulesPreviewOpened && m.rulesClose != nil && m.rulesClose.Rect().Contains(event.X, event.Y):
		m.closeRulesPreview()
	default:
		return false
	}
	return true
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

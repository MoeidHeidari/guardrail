package screens

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func (m *Main) dashboardVisible() bool {
	return m.dashboardTab && !m.contentTab
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
	}
	editorTheme := m.theme
	editorTheme.Surface = m.theme.Background
	m.feed = components.NewContainer(editorTheme, tui.WithFlexGrow(1),
		tui.WithHeight(0), tui.WithScrollable(scrollMode),
		tui.WithScrollOffset(offsetX, offset))
	m.styleScrollbar(m.feed)
	title := fmt.Sprintf("%s   /   %d events", display(m.root), len(m.changes))
	m.dashboardButton = m.label("0 Dashboard")
	m.watchTabButton = m.label("1 Watch")
	m.contentTabButton = m.label("2 Content")
	active := m.watchTabButton
	switch {
	case m.contentTab:
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
		active, title = m.dashboardButton, "Dashboard"
		m.renderDashboard()
	default:
		m.renderChanges()
	}
	for _, tab := range []*tui.Element{m.dashboardButton, m.watchTabButton, m.contentTabButton} {
		tab.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
	}
	active.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Underline()))
	tabs := components.NewContainer(editorTheme, tui.WithDirection(tui.Row), tui.WithGap(2),
		tui.WithHeight(1), tui.WithFlexShrink(0))
	tabs.AddChild(m.dashboardButton, m.watchTabButton, m.contentTabButton)
	subtitle := m.label(title)
	subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)))
	if m.changesFocus {
		subtitle.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
	}
	content := components.NewSection(editorTheme, "", tabs, subtitle, m.feed)
	content.Apply(tui.WithPaddingTRBL(1, 2, 1, 2))
	return content
}

func (m *Main) renderDashboard() {
	theme := m.theme
	theme.Surface = theme.Background
	overview := components.NewSection(theme, "Workspace",
		m.label(display(m.root)),
		m.label(fmt.Sprintf("%d files & folders", len(m.entries))),
		m.label(fmt.Sprintf("%d changed files", m.changedFiles)),
	)
	m.feed.AddChild(overview)
}

func (m *Main) renderChanges() {
	if len(m.changes) == 0 {
		m.feed.Apply(tui.WithScrollable(tui.ScrollNone), tui.WithJustify(tui.JustifyCenter))
		title := m.label("Watching for filesystem changes...")
		title.Apply(tui.WithTextAlign(tui.TextAlignCenter))
		hint := m.label("Select a file to preview its content.")
		hint.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted)),
			tui.WithTextAlign(tui.TextAlignCenter))
		m.feed.AddChild(title, m.label(""), hint)
	}
	for i := len(m.changes) - 1; i >= 0; i-- {
		change := m.changes[i]
		color := m.theme.Modified
		if change.Kind == "created" {
			color = m.theme.Added
		} else if change.Kind == "deleted" {
			color = m.theme.Danger
		}
		el := m.label("")
		el.Apply(tui.WithRichText(
			tui.TextSpan{Text: change.Time.Format("15:04:05") + "  ", Style: tui.NewStyle().Foreground(m.theme.Muted)},
			tui.TextSpan{Text: fmt.Sprintf("%-8s  ", change.Kind), Style: tui.NewStyle().Foreground(color)},
			tui.TextSpan{Text: display(change.Path)},
		))
		m.feed.AddChild(el)
	}
}

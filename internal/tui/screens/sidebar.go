package screens

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

type workspacePage int

const (
	overviewPage workspacePage = iota
	violationsPage
	rulesPage
)

var workspaceLabels = [...]string{"Overview", "Violations", "Rules"}
var workspaceShortcuts = [...]string{"o", "v", "u"}

func (m *Main) styleSidebarRow(item *tui.Element, selected bool, background tui.Color) {
	color := m.theme.Surface
	if selected {
		color = m.theme.Accent
	} else {
		background = m.theme.Surface
	}
	item.Apply(tui.WithHeight(1), tui.WithPaddingTRBL(0, 1, 0, 1),
		tui.WithBorderTRBL(tui.BorderNone, tui.BorderNone, tui.BorderNone, tui.BorderSingle),
		tui.WithBorderStyle(tui.Style{Fg: color, Bg: background}))
}

func (m *Main) openWorkspace(page workspacePage) {
	m.watchFocus = false
	m.workspacePage, m.navigationIndex = page, int(page)
	m.dashboardTab, m.contentTab = true, false
	m.animationOnly = false
}

func (m *Main) openNavigation(index int) {
	if index == len(workspaceLabels) {
		m.watchFocus, m.navigationFocus, m.changesFocus = true, false, false
		m.navigationIndex = index
		return
	}
	m.openWorkspace(workspacePage(index))
}

func (m *Main) cycleFocus() {
	lastRulesPane := 1
	if m.rulesPreviewOpened {
		lastRulesPane = 2
	}
	if m.rulesVisible() && m.changesFocus && m.rulesFocus < lastRulesPane {
		m.rulesFocus++
		return
	}
	m.rulesFocus = 0
	if m.watchFocus {
		m.watchFocus = false
		m.navigationIndex = int(m.workspacePage)
		return
	}
	switch {
	case m.navigationFocus:
		m.navigationFocus, m.changesFocus = false, false
	case m.changesFocus:
		m.navigationFocus, m.changesFocus = true, false
		m.navigationIndex = int(m.workspacePage)
	default:
		m.changesFocus = true
	}
}

func (m *Main) renderSidebar() *tui.Element {
	m.navigationItems = nil
	heading := m.label("Workspace")
	heading.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted).Bold()))
	menu := components.NewSection(m.theme, "", heading)
	menu.Apply(tui.WithHeight(4), tui.WithFlexShrink(1))
	for i, name := range workspaceLabels {
		label := name + "  [" + workspaceShortcuts[i] + "]"
		item := m.label(label)
		if m.dashboardVisible() && m.workspacePage == workspacePage(i) {
			item.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Selection}),
				tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
		}
		m.styleSidebarRow(item, (m.dashboardVisible() && m.workspacePage == workspacePage(i)) ||
			(m.navigationFocus && m.navigationIndex == i), m.theme.Selection)
		if m.navigationFocus && m.navigationIndex == i {
			item.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Selection}),
				tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Bold()))
		}
		m.navigationItems = append(m.navigationItems, item)
		menu.AddChild(item)
	}
	treeHeading := m.label("Context")
	color := m.theme.Muted
	if !m.changesFocus && !m.navigationFocus {
		color = m.theme.Accent
	}
	treeHeading.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(color).Bold()))
	context := components.NewSection(m.theme, "", treeHeading, m.tree)
	context.Apply(tui.WithHeight(0), tui.WithFlexGrow(1))
	m.watchTabButton = m.label("1 Watch")
	if m.watchFocus {
		m.watchTabButton.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Selection}),
			tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent)))
	}
	if m.navigationFocus && m.navigationIndex == len(workspaceLabels) {
		m.watchTabButton.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Selection}),
			tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Bold()))
	}
	m.styleSidebarRow(m.watchTabButton, m.watchFocus ||
		(m.navigationFocus && m.navigationIndex == len(workspaceLabels)), m.theme.Selection)
	m.watchPane = components.NewContainer(m.theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(m.watchX, m.changeOffset))
	m.styleScrollbar(m.watchPane)
	m.renderActivity(m.watchPane)
	utilities := components.NewSection(m.theme, "", m.watchTabButton, m.watchPane)
	utilities.Apply(tui.WithHeight(min(8, max(3, m.renderHeight/4))), tui.WithFlexShrink(0))
	divider := func() *tui.Element {
		return tui.New(tui.WithHR(), tui.WithHeight(1), tui.WithFlexShrink(0),
			tui.WithBackground(tui.Style{Bg: m.theme.Surface}),
			tui.WithTextStyle(tui.Style{Fg: m.theme.Separator, Bg: m.theme.Surface}))
	}
	sidebar := components.NewSection(m.theme, "", menu, divider(), context, divider(), utilities)
	sidebar.Apply(tui.WithPaddingTRBL(1, 2, 1, 2))
	return sidebar
}

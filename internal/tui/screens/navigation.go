package screens

import (
	"path/filepath"

	"github.com/grindlemire/go-tui"
)

func (m *Main) move(delta int) {
	if m.watchFocus {
		m.moveWatch(delta)
		return
	}
	if m.navigationFocus {
		if delta == 1 && m.navigationIndex == len(workspaceLabels)-1 && len(m.rows) > 0 {
			m.navigationFocus = false
			return
		}
		m.navigationIndex = max(0, min(m.navigationIndex+delta, len(workspaceLabels)-1))
		return
	}
	if m.changesFocus {
		if m.moveRules(delta) || m.moveOverview(delta) {
			return
		}
		if m.dashboardVisible() {
			if m.workspacePage != overviewPage {
				if m.findingsVisible() {
					m.moveFinding(delta)
				} else if m.feed != nil {
					_, maximum := m.feed.MaxScroll()
					m.findingOffset = max(0, min(m.findingOffset+delta, maximum))
				}
			}
			return
		}
		if m.contentTab {
			m.contentOffset = max(0, min(m.contentOffset+delta, len(m.preview.lines)-viewport(m.feed)))
			return
		}
		m.changeOffset = max(0, min(m.changeOffset+delta, len(m.changes)-viewport(m.feed)))
		return
	}
	if m.moveContextBoundary(delta) {
		return
	}
	m.selected = max(0, min(m.selected+delta, len(m.rows)-1))
	m.reveal()
}

func (m *Main) expand() {
	if m.watchFocus {
		m.openWatchFile()
		return
	}
	if m.stepFilter(1) {
		return
	}
	if m.rulesVisible() && m.changesFocus && m.rulesFocus < 2 {
		m.activateRulesPane()
		return
	}
	if m.changesFocus && m.findingsVisible() {
		return
	}
	if m.navigationFocus {
		m.activate()
		return
	}
	if m.changesFocus || len(m.rows) == 0 {
		return
	}
	entry := m.rows[m.selected].entry
	if entry.IsDir() && m.collapsed[entry.Path] {
		m.toggle()
	} else if entry.IsDir() && m.selected+1 < len(m.rows) && m.rows[m.selected+1].depth > m.rows[m.selected].depth {
		m.move(1)
	} else if !entry.IsDir() {
		m.activate()
	}
}

func (m *Main) collapse() {
	if m.stepFilter(-1) {
		return
	}
	if m.changesFocus && (m.findingsVisible() || (m.overviewVisible() && !m.overviewPreviewFocus)) {
		m.focusSidebar()
		return
	}
	if m.overviewVisible() && m.changesFocus && m.overviewPreviewFocus {
		m.overviewPreviewFocus = false
		return
	}
	if m.rulesVisible() && m.changesFocus && m.rulesFocus > 0 {
		m.rulesFocus--
		return
	}
	if m.watchFocus {
		m.focusContext()
		return
	}
	if m.navigationFocus {
		return
	}
	if m.changesFocus {
		m.focusSidebar()
		return
	}
	if len(m.rows) == 0 {
		return
	}
	entry := m.rows[m.selected].entry
	if entry.IsDir() && !m.collapsed[entry.Path] {
		m.toggle()
		return
	}
	parent := filepath.Dir(entry.Path)
	for i, row := range m.rows {
		if row.entry.Path == parent {
			m.selected = i
			m.reveal()
			return
		}
	}
}

func (m *Main) KeyMap() tui.KeyMap {
	if m.searchOpen {
		return m.searchKeyMap()
	}
	quit := func(event tui.KeyEvent) { event.App().Stop() }
	return tui.KeyMap{
		tui.On(tui.Rune('q'), quit),
		tui.On(tui.KeyCtrlC, quit),
		tui.On(tui.Rune('/'), func(tui.KeyEvent) { m.openSearch() }),
		tui.On(tui.KeyTab, func(tui.KeyEvent) { m.cycleFocus() }),
		tui.On(tui.KeyEscape, func(tui.KeyEvent) { m.focusSidebar() }),
		tui.On(tui.Rune('o'), func(tui.KeyEvent) { m.openWorkspaceFocused(overviewPage) }),
		tui.On(tui.Rune('v'), func(tui.KeyEvent) { m.openWorkspaceFocused(violationsPage) }),
		tui.On(tui.Rune('u'), func(tui.KeyEvent) { m.openWorkspaceFocused(rulesPage) }),
		tui.On(tui.Rune('c'), func(tui.KeyEvent) { m.focusContext() }),
		tui.On(tui.Rune('w'), func(tui.KeyEvent) { m.openNavigation(len(workspaceLabels)) }),
		tui.On(tui.Rune('<'), func(tui.KeyEvent) { m.cycleTab(-1) }),
		tui.On(tui.Rune('>'), func(tui.KeyEvent) { m.cycleTab(1) }),
		tui.On(tui.Rune('0'), func(tui.KeyEvent) { m.jumpTab(0) }),
		tui.On(tui.Rune('1'), func(tui.KeyEvent) { m.jumpTab(1) }),
		tui.On(tui.Rune('2'), func(tui.KeyEvent) { m.jumpTab(2) }),
		tui.On(tui.Rune('3'), func(tui.KeyEvent) { m.jumpTab(3) }),
		tui.On(tui.Rune('4'), func(tui.KeyEvent) { m.jumpTab(4) }),
		tui.On(tui.Rune('5'), func(tui.KeyEvent) { m.jumpTab(5) }),
		tui.On(tui.Rune('6'), func(tui.KeyEvent) { m.jumpTab(6) }),
		tui.On(tui.Rune('7'), func(tui.KeyEvent) { m.jumpTab(7) }),
		tui.On(tui.Rune('8'), func(tui.KeyEvent) { m.jumpTab(8) }),
		tui.On(tui.Rune('9'), func(tui.KeyEvent) { m.jumpTab(9) }),
		tui.On(tui.Rune('r'), func(tui.KeyEvent) { m.previewDirty = true }),
		tui.On(tui.Rune('x'), func(tui.KeyEvent) { m.closeActivePreview() }),
		tui.On(tui.Rune('['), func(tui.KeyEvent) { m.scrollFocusedX(-4) }),
		tui.On(tui.Rune(']'), func(tui.KeyEvent) { m.scrollFocusedX(4) }),
		tui.On(tui.Rune('a'), func(tui.KeyEvent) {
			if m.findingsVisible() || m.overviewVisible() {
				m.setFindingFilter("")
			}
		}),
		tui.On(tui.Rune('b'), func(tui.KeyEvent) {
			if m.findingsVisible() || m.overviewVisible() {
				m.setFindingFilter("violation")
			}
		}),
		tui.On(tui.Rune('e'), func(tui.KeyEvent) {
			if m.findingsVisible() || m.overviewVisible() {
				m.setFindingFilter("error")
			}
		}),
		tui.On(tui.KeyUp, func(tui.KeyEvent) { m.move(-1) }),
		tui.On(tui.Rune('k'), func(tui.KeyEvent) { m.move(-1) }),
		tui.On(tui.KeyDown, func(tui.KeyEvent) { m.move(1) }),
		tui.On(tui.Rune('j'), func(tui.KeyEvent) { m.move(1) }),
		tui.On(tui.KeyEnter, func(tui.KeyEvent) { m.activate() }),
		tui.On(tui.Rune(' '), func(tui.KeyEvent) { m.activateSpace() }),
		tui.On(tui.KeyRight, func(tui.KeyEvent) { m.expand() }),
		tui.On(tui.Rune('l'), func(tui.KeyEvent) { m.expand() }),
		tui.On(tui.KeyLeft, func(tui.KeyEvent) { m.collapse() }),
		tui.On(tui.Rune('h'), func(tui.KeyEvent) { m.collapse() }),
		tui.On(tui.KeyPageUp, func(tui.KeyEvent) { m.move(-m.pageSize()) }),
		tui.On(tui.KeyPageDown, func(tui.KeyEvent) { m.move(m.pageSize()) }),
		tui.On(tui.KeyHome, func(tui.KeyEvent) { m.move(-max(len(m.rows), len(m.changes), len(m.preview.lines), len(m.findings))) }),
		tui.On(tui.KeyEnd, func(tui.KeyEvent) { m.move(max(len(m.rows), len(m.changes), len(m.preview.lines), len(m.findings))) }),
	}
}

func (m *Main) HandleMouse(event tui.MouseEvent) bool {
	if m.handleSearchMouse(event) {
		return true
	}
	if m.handleScrollbar(event) {
		return true
	}
	if event.Action != tui.MousePress {
		return false
	}
	if m.handleSidebarHorizontalWheel(event) {
		return true
	}
	if m.handlePreviewClose(event) {
		return true
	}
	if m.handleFilterMouse(event) {
		return true
	}
	if m.handleOverviewMouse(event) || m.handleRulesMouse(event) {
		return true
	}
	if event.Button == tui.MouseLeft {
		for i, item := range m.navigationItems {
			if item.Rect().Contains(event.X, event.Y) {
				m.navigationFocus, m.changesFocus = true, false
				m.openNavigation(i)
				return true
			}
		}
		if m.dashboardButton != nil && m.dashboardButton.Rect().Contains(event.X, event.Y) {
			m.openWorkspace(m.workspacePage)
			m.changesFocus = true
			return true
		}
		if m.watchTabButton != nil && m.watchTabButton.Rect().Contains(event.X, event.Y) {
			m.openNavigation(len(workspaceLabels))
			return true
		}
		if m.handleContentTabClick(event) {
			return true
		}
	}
	if m.handleWatchMouse(event) {
		return true
	}
	inTree := m.tree != nil && m.tree.Rect().Contains(event.X, event.Y)
	inFeed := m.feed != nil && m.feed.Rect().Contains(event.X, event.Y)
	if !inTree && !inFeed {
		return false
	}
	m.changesFocus = inFeed
	m.watchFocus = false
	m.navigationFocus = false
	switch event.Button {
	case tui.MouseWheelUp, tui.MouseWheelDown:
		delta := 3
		if event.Button == tui.MouseWheelUp {
			delta = -delta
		}
		if inFeed {
			m.move(delta)
		} else {
			m.treeOffset = max(0, min(m.treeOffset+delta, len(m.rows)*treeRowHeight-viewport(m.tree)))
		}
		return true
	case tui.MouseLeft:
		if inFeed && m.findingsVisible() {
			_, offset := m.feed.ScrollOffset()
			y := event.Y - m.feed.ContentRect().Y + offset
			index := y / findingHeight
			indices := m.visibleFindings()
			if y >= 0 && index < len(indices) {
				m.findingIndex = indices[index]
				m.openFinding()
			}
		}
		if inTree {
			_, offset := m.tree.ScrollOffset()
			index := (event.Y - m.tree.ContentRect().Y + offset) / treeRowHeight
			if index >= 0 && index < len(m.rows) {
				m.selected = index
				m.toggle()
			}
		}
		return true
	}
	return false
}

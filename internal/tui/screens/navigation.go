package screens

import (
	"path/filepath"

	"github.com/grindlemire/go-tui"
)

func (m *Main) move(delta int) {
	if m.watchFocus {
		m.changeOffset = max(0, min(m.changeOffset+delta, max(0, len(m.changes)-viewport(m.watchPane))))
		return
	}
	if m.navigationFocus {
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
	m.selected = max(0, min(m.selected+delta, len(m.rows)-1))
	m.reveal()
}

func (m *Main) reveal() {
	top := m.selected * treeRowHeight
	if top < m.treeOffset {
		m.treeOffset = top
	} else if top+treeRowHeight > m.treeOffset+viewport(m.tree) {
		m.treeOffset = max(top, top+treeRowHeight-viewport(m.tree))
	}
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)*treeRowHeight-viewport(m.tree)))
}

func (m *Main) expand() {
	if m.scrollFindingX(4) {
		return
	}
	if m.scrollOverviewX(4) {
		return
	}
	if m.watchFocus {
		m.scrollSidebarX(4)
		return
	}
	if m.navigationFocus {
		m.toggle()
		return
	}
	if m.changesFocus && m.contentTab {
		m.scrollContentX(4)
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
	} else {
		m.scrollSidebarX(4)
	}
}

func (m *Main) collapse() {
	if m.scrollFindingX(-4) {
		return
	}
	if m.scrollOverviewX(-4) {
		return
	}
	if m.watchFocus {
		m.scrollSidebarX(-4)
		return
	}
	if m.navigationFocus {
		return
	}
	if m.changesFocus && m.contentTab {
		m.scrollContentX(-4)
		return
	}
	if m.changesFocus || len(m.rows) == 0 {
		return
	}
	if m.treeX > 0 {
		m.scrollSidebarX(-4)
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
	quit := func(event tui.KeyEvent) { event.App().Stop() }
	page := func() int {
		if m.watchFocus {
			return viewport(m.watchPane)
		}
		if m.changesFocus {
			if m.findingsVisible() {
				return max(1, viewport(m.feed)/findingHeight)
			}
			return viewport(m.feed)
		}
		return max(1, viewport(m.tree)/treeRowHeight)
	}
	return tui.KeyMap{
		tui.On(tui.Rune('q'), quit),
		tui.On(tui.KeyCtrlC, quit),
		tui.On(tui.KeyTab, func(tui.KeyEvent) { m.cycleFocus() }),
		tui.On(tui.Rune('0'), func(tui.KeyEvent) { m.openWorkspace(m.workspacePage) }),
		tui.On(tui.Rune('1'), func(tui.KeyEvent) { m.openNavigation(len(workspaceLabels)) }),
		tui.On(tui.Rune('2'), func(tui.KeyEvent) {
			if m.contentOpened {
				m.watchFocus = false
				m.dashboardTab, m.contentTab = false, true
			}
		}),
		tui.On(tui.Rune('o'), func(tui.KeyEvent) { m.openWorkspace(overviewPage) }),
		tui.On(tui.Rune('v'), func(tui.KeyEvent) { m.openWorkspace(violationsPage) }),
		tui.On(tui.Rune('u'), func(tui.KeyEvent) { m.openWorkspace(rulesPage) }),
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
		tui.On(tui.KeyEnter, func(tui.KeyEvent) { m.toggle() }),
		tui.On(tui.Rune(' '), func(tui.KeyEvent) { m.toggle() }),
		tui.On(tui.KeyRight, func(tui.KeyEvent) { m.expand() }),
		tui.On(tui.Rune('l'), func(tui.KeyEvent) { m.expand() }),
		tui.On(tui.KeyLeft, func(tui.KeyEvent) { m.collapse() }),
		tui.On(tui.Rune('h'), func(tui.KeyEvent) { m.collapse() }),
		tui.On(tui.KeyPageUp, func(tui.KeyEvent) { m.move(-page()) }),
		tui.On(tui.KeyPageDown, func(tui.KeyEvent) { m.move(page()) }),
		tui.On(tui.KeyHome, func(tui.KeyEvent) { m.move(-max(len(m.rows), len(m.changes), len(m.preview.lines), len(m.findings))) }),
		tui.On(tui.KeyEnd, func(tui.KeyEvent) { m.move(max(len(m.rows), len(m.changes), len(m.preview.lines), len(m.findings))) }),
	}
}

func (m *Main) HandleMouse(event tui.MouseEvent) bool {
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
		if m.contentTabButton != nil && m.contentTabButton.Rect().Contains(event.X, event.Y) {
			m.watchFocus = false
			m.dashboardTab, m.contentTab = false, true
			return true
		}
	}
	if m.watchPane != nil && m.watchPane.Rect().Contains(event.X, event.Y) {
		m.watchFocus, m.changesFocus, m.navigationFocus = true, false, false
		switch event.Button {
		case tui.MouseWheelUp:
			m.move(-3)
		case tui.MouseWheelDown:
			m.move(3)
		}
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

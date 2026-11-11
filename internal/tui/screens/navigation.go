package screens

import (
	"path/filepath"

	"github.com/grindlemire/go-tui"
)

func viewport(el *tui.Element) int {
	if el == nil {
		return 10
	}
	_, height := el.ViewportSize()
	return max(1, height)
}

func (m *Main) move(delta int) {
	if m.changesFocus {
		if m.dashboardVisible() {
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
	if m.selected < m.treeOffset {
		m.treeOffset = m.selected
	} else if m.selected >= m.treeOffset+viewport(m.tree) {
		m.treeOffset = m.selected - viewport(m.tree) + 1
	}
}

func (m *Main) toggle() {
	if m.changesFocus || len(m.rows) == 0 {
		return
	}
	if !m.rows[m.selected].entry.IsDir() {
		m.contentTab = true
		return
	}
	path := m.rows[m.selected].entry.Path
	m.collapsed[path] = !m.collapsed[path]
	m.flattenRows()
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)-viewport(m.tree)))
	m.reveal()
}

func (m *Main) expand() {
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
	}
}

func (m *Main) collapse() {
	if m.changesFocus && m.contentTab {
		m.scrollContentX(-4)
		return
	}
	if m.changesFocus || len(m.rows) == 0 {
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
		if m.changesFocus {
			return viewport(m.feed)
		}
		return viewport(m.tree)
	}
	return tui.KeyMap{
		tui.On(tui.Rune('q'), quit),
		tui.On(tui.KeyCtrlC, quit),
		tui.On(tui.KeyTab, func(tui.KeyEvent) { m.changesFocus = !m.changesFocus }),
		tui.On(tui.Rune('0'), func(tui.KeyEvent) { m.dashboardTab, m.contentTab = true, false }),
		tui.On(tui.Rune('1'), func(tui.KeyEvent) { m.dashboardTab, m.contentTab = false, false }),
		tui.On(tui.Rune('2'), func(tui.KeyEvent) { m.dashboardTab, m.contentTab = false, true }),
		tui.On(tui.Rune('r'), func(tui.KeyEvent) { m.previewDirty = true }),
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
		tui.On(tui.KeyHome, func(tui.KeyEvent) { m.move(-max(len(m.rows), len(m.changes), len(m.preview.lines))) }),
		tui.On(tui.KeyEnd, func(tui.KeyEvent) { m.move(max(len(m.rows), len(m.changes), len(m.preview.lines))) }),
	}
}

func (m *Main) HandleMouse(event tui.MouseEvent) bool {
	if m.handleScrollbar(event) {
		return true
	}
	if event.Action != tui.MousePress {
		return false
	}
	if event.Button == tui.MouseLeft {
		if m.dashboardButton != nil && m.dashboardButton.Rect().Contains(event.X, event.Y) {
			m.dashboardTab, m.contentTab = true, false
			return true
		}
		if m.watchTabButton != nil && m.watchTabButton.Rect().Contains(event.X, event.Y) {
			m.dashboardTab, m.contentTab = false, false
			return true
		}
		if m.contentTabButton != nil && m.contentTabButton.Rect().Contains(event.X, event.Y) {
			m.dashboardTab, m.contentTab = false, true
			return true
		}
	}
	inTree := m.tree != nil && m.tree.Rect().Contains(event.X, event.Y)
	inFeed := m.feed != nil && m.feed.Rect().Contains(event.X, event.Y)
	if !inTree && !inFeed {
		return false
	}
	m.changesFocus = inFeed
	switch event.Button {
	case tui.MouseWheelUp, tui.MouseWheelDown:
		delta := 3
		if event.Button == tui.MouseWheelUp {
			delta = -delta
		}
		if inFeed {
			m.move(delta)
		} else {
			m.treeOffset = max(0, min(m.treeOffset+delta, len(m.rows)-viewport(m.tree)))
		}
		return true
	case tui.MouseLeft:
		if inTree {
			_, offset := m.tree.ScrollOffset()
			index := event.Y - m.tree.ContentRect().Y + offset
			if index >= 0 && index < len(m.rows) {
				m.selected = index
				m.toggle()
			}
		}
		return true
	}
	return false
}

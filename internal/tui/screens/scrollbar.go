package screens

import "github.com/grindlemire/go-tui"

func scrollHorizontal(pane *tui.Element, offset *int, delta int) {
	if pane == nil {
		return
	}
	maximum, _ := pane.MaxScroll()
	*offset = max(0, min(*offset+delta, maximum))
}

func (m *Main) scrollSidebarX(delta int) {
	if m.watchFocus {
		scrollHorizontal(m.watchPane, &m.watchX, delta)
	} else if !m.navigationFocus && !m.changesFocus {
		scrollHorizontal(m.tree, &m.treeX, delta)
	}
}

func (m *Main) handleSidebarHorizontalWheel(event tui.MouseEvent) bool {
	if event.Mod&tui.ModShift == 0 || (event.Button != tui.MouseWheelUp && event.Button != tui.MouseWheelDown) {
		return false
	}
	delta := 4
	if event.Button == tui.MouseWheelUp {
		delta = -delta
	}
	switch {
	case m.tree != nil && m.tree.Rect().Contains(event.X, event.Y):
		m.watchFocus, m.navigationFocus, m.changesFocus = false, false, false
	case m.watchPane != nil && m.watchPane.Rect().Contains(event.X, event.Y):
		m.watchFocus, m.navigationFocus, m.changesFocus = true, false, false
	default:
		return false
	}
	m.scrollSidebarX(delta)
	return true
}

func (m *Main) scrollFocusedX(delta int) {
	if m.watchFocus || (!m.navigationFocus && !m.changesFocus) {
		m.scrollSidebarX(delta)
		return
	}
	if !m.changesFocus {
		return
	}
	if m.rulesVisible() {
		switch m.rulesFocus {
		case 0:
			scrollHorizontal(m.rulesList, &m.rulesListX, delta)
		case 1:
			scrollHorizontal(m.rulesFiles, &m.rulesFilesX, delta)
		case 2:
			m.scrollOverviewX(delta)
		}
		return
	}
	if m.scrollFindingX(delta) || m.scrollOverviewX(delta) {
		return
	}
	if m.contentTab {
		m.scrollContentX(delta)
	}
}

func viewport(el *tui.Element) int {
	if el == nil {
		return 10
	}
	_, height := el.ViewportSize()
	return max(1, height)
}

func (m *Main) styleScrollbar(el *tui.Element) {
	el.Apply(
		tui.WithScrollbarStyle(tui.Style{Fg: m.theme.Separator, Bg: m.theme.Background}),
		tui.WithScrollbarThumbStyle(tui.Style{Fg: m.theme.Accent, Bg: m.theme.Background}),
	)
}

func scrollThumb(el *tui.Element) (top, size, travel, maximum int) {
	height := viewport(el)
	_, content := el.ContentSize()
	_, maximum = el.MaxScroll()
	size = min(height, max(1, height*height/max(1, content)))
	travel = height - size
	_, offset := el.ScrollOffset()
	if maximum > 0 {
		top = offset * travel / maximum
	}
	return
}

func (m *Main) handleScrollbar(event tui.MouseEvent) bool {
	if m.dragPane != nil {
		if event.Action == tui.MouseRelease {
			m.dragPane = nil
			return true
		}
		if event.Action == tui.MouseDrag {
			el := m.dragPane
			_, _, travel, maximum := scrollThumb(el)
			position := max(0, min(travel, event.Y-el.ContentRect().Y-m.dragGrab))
			offset := 0
			if travel > 0 {
				offset = (position*maximum + travel/2) / travel
			}
			m.setScroll(el, offset)
			return true
		}
	}
	if event.Action != tui.MousePress || event.Button != tui.MouseLeft {
		return false
	}
	panes := []*tui.Element{m.tree, m.feed, m.watchPane}
	if m.overviewVisible() {
		panes = append(panes, m.overviewList, m.overviewPreview)
	}
	if m.rulesVisible() {
		panes = append(panes, m.rulesList, m.rulesFiles, m.rulesPreview)
	}
	for _, el := range panes {
		if el == nil {
			continue
		}
		rect := el.ContentRect()
		top, size, _, maximum := scrollThumb(el)
		if maximum == 0 || event.X != rect.X+rect.Width-1 || !rect.Contains(event.X, event.Y) {
			continue
		}
		m.changesFocus = el == m.feed || el == m.overviewList || el == m.overviewPreview
		for i, pane := range []*tui.Element{m.rulesList, m.rulesFiles, m.rulesPreview} {
			if m.rulesVisible() && el == pane {
				m.rulesFocus, m.changesFocus = i, true
			}
		}
		m.overviewPreviewFocus = el == m.overviewPreview
		m.watchFocus = el == m.watchPane
		m.navigationFocus = false
		y := event.Y - rect.Y
		if y >= top && y < top+size {
			m.dragPane, m.dragGrab = el, y-top
		} else {
			_, offset := el.ScrollOffset()
			delta := viewport(el)
			if y < top {
				delta = -delta
			}
			m.setScroll(el, max(0, min(maximum, offset+delta)))
		}
		return true
	}
	return false
}

func (m *Main) setScroll(el *tui.Element, offset int) {
	if el == m.rulesPreview {
		m.contentOffset = offset
	} else if el == m.rulesList {
		m.findingOffset = offset
	} else if el == m.rulesFiles {
		m.rulesFileSelection = min(offset, max(0, len(m.rulesMatches)-1))
		m.rulesPath = ""
		m.updateRuleFiles()
	} else if el == m.overviewPreview {
		m.contentOffset = offset
	} else if el == m.overviewList {
		m.findingOffset = offset
	} else if el == m.watchPane {
		m.changeOffset = offset
	} else if el == m.tree {
		m.treeOffset = offset
	} else if m.contentTab {
		m.contentOffset = offset
	} else if m.dashboardVisible() {
		m.findingOffset = offset
	} else {
		m.changeOffset = offset
	}
	if m.dragPane != nil {
		m.dragPane = el
	}
}

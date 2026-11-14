package screens

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/grindlemire/go-tui"
)

func (m *Main) saveTabView() {
	if !m.contentTab || m.contentActive == "" {
		return
	}
	if m.tabViews == nil {
		m.tabViews = make(map[string][2]int)
	}
	m.tabViews[m.contentActive] = [2]int{m.contentX, m.contentOffset}
}

func (m *Main) tabView(path string) (int, int) {
	view := m.tabViews[path]
	return view[0], view[1]
}

func (m *Main) openTab(path string) {
	switching := !m.contentTab || path != m.contentActive
	if switching {
		m.saveTabView()
	}
	if !slices.Contains(m.contentTabs, path) {
		m.contentTabs = append(m.contentTabs, path)
	}
	m.contentActive = path
	m.contentTab, m.contentOpened = true, true
	m.watchFocus, m.dashboardTab = false, false
	if switching {
		m.contentX, m.contentOffset = m.tabView(path)
		m.invalidatePreview()
	}
	m.animationOnly = false
}

func (m *Main) activateTabAt(index int) {
	if index < 0 || index >= len(m.contentTabs) {
		return
	}
	path := m.contentTabs[index]
	m.openTab(path)
	if position, exists := m.rowPositions[path]; exists {
		m.selected = position
		m.reveal()
	}
}

func (m *Main) cycleTab(delta int) {
	if len(m.contentTabs) == 0 {
		return
	}
	position := 0
	if m.contentTab {
		position = 1 + max(slices.Index(m.contentTabs, m.contentActive), 0)
	}
	m.jumpTab((position + delta + len(m.contentTabs) + 1) % (len(m.contentTabs) + 1))
}

func (m *Main) jumpTab(n int) {
	switch {
	case n == 0:
		m.openWorkspace(m.workspacePage)
	case n <= len(m.contentTabs):
		m.activateTabAt(n - 1)
	}
}

func (m *Main) closeTab(index int) {
	if index < 0 || index >= len(m.contentTabs) {
		return
	}
	closed := m.contentTabs[index]
	wasContent := m.contentTab
	m.contentTabs = slices.Delete(m.contentTabs, index, index+1)
	delete(m.tabViews, closed)
	if closed == m.contentActive {
		m.contentActive = ""
		if wasContent {
			m.contentOffset, m.contentX = 0, 0
		}
		if len(m.contentTabs) > 0 {
			m.contentActive = m.contentTabs[min(index, len(m.contentTabs)-1)]
			if wasContent {
				m.contentX, m.contentOffset = m.tabView(m.contentActive)
			}
		}
	}
	if len(m.contentTabs) == 0 {
		m.contentOpened, m.contentTab = false, false
		m.dashboardTab = true
	}
	if wasContent {
		m.invalidatePreview()
	}
}

func (m *Main) closeContent() {
	m.closeTab(slices.Index(m.contentTabs, m.contentActive))
}

func (m *Main) tabLabel(index int) string {
	path := m.contentTabs[index]
	name := filepath.Base(path)
	for _, other := range m.contentTabs {
		if other != path && filepath.Base(other) == name {
			name = path
			break
		}
	}
	if index < 9 {
		return strconv.Itoa(index+1) + " " + display(name)
	}
	return display(name)
}

func (m *Main) visibleTabs(budget int) (int, int) {
	count := len(m.contentTabs)
	active := slices.Index(m.contentTabs, m.contentActive)
	start := max(0, min(m.tabStart, count-1))
	if active >= 0 && active < start {
		start = active
	}
	fit := func(start int) int {
		used, end := 0, start
		for end < count {
			width := tui.StringWidth(m.tabLabel(end)) + 6
			if used+width > budget && end > start {
				break
			}
			used += width
			end++
		}
		return end
	}
	end := fit(start)
	for active >= end && start < active {
		start++
		end = fit(start)
	}
	m.tabStart, m.tabEnd = start, end
	return start, end
}

func (m *Main) renderContentTabs(tabs *tui.Element) {
	m.contentTabButtons = make([]*tui.Element, len(m.contentTabs))
	m.contentCloses = make([]*tui.Element, len(m.contentTabs))
	m.tabMore = [2]*tui.Element{}
	m.contentTabButton, m.contentClose = nil, nil
	if len(m.contentTabs) == 0 {
		return
	}
	budget := max(12, m.renderWidth-52-tui.StringWidth(m.dashboardButton.Text()))
	start, end := m.visibleTabs(budget)
	muted := tui.NewStyle().Foreground(m.theme.Muted)
	if start > 0 {
		m.tabMore[0] = m.label(fmt.Sprintf("‹%d", start))
		m.tabMore[0].Apply(tui.WithWidth(tui.StringWidth(m.tabMore[0].Text())), tui.WithTextStyle(muted))
		tabs.AddChild(m.tabMore[0])
	}
	for i := start; i < end; i++ {
		path := m.contentTabs[i]
		button := m.label(m.tabLabel(i))
		closeButton := m.label("[x]")
		closeButton.Apply(tui.WithWidth(3), tui.WithTextStyle(muted))
		style := muted
		if path == m.contentActive {
			m.contentTabButton, m.contentClose = button, closeButton
			if m.contentTab {
				style = tui.NewStyle().Foreground(m.theme.Accent).Underline()
			}
		}
		button.Apply(tui.WithTextStyle(style))
		m.highlightFocus(button, m.changesFocus && m.contentTab && path == m.contentActive)
		m.contentTabButtons[i], m.contentCloses[i] = button, closeButton
		tabs.AddChild(button, closeButton)
	}
	if end < len(m.contentTabs) {
		m.tabMore[1] = m.label(fmt.Sprintf("%d›", len(m.contentTabs)-end))
		m.tabMore[1].Apply(tui.WithWidth(tui.StringWidth(m.tabMore[1].Text())), tui.WithTextStyle(muted))
		tabs.AddChild(m.tabMore[1])
	}
	if m.contentTabButton == nil {
		m.contentTabButton, m.contentClose = m.contentTabButtons[start], m.contentCloses[start]
	}
}

func (m *Main) handleContentTabMouse(event tui.MouseEvent) bool {
	for i := range m.contentTabs {
		if i < len(m.contentCloses) && m.contentCloses[i] != nil && m.contentCloses[i].Rect().Contains(event.X, event.Y) {
			m.closeTab(i)
			return true
		}
	}
	return false
}

func (m *Main) handleContentTabClick(event tui.MouseEvent) bool {
	for i, button := range m.contentTabButtons {
		if button != nil && button.Rect().Contains(event.X, event.Y) {
			m.activateTabAt(i)
			return true
		}
	}
	if more := m.tabMore[0]; more != nil && more.Rect().Contains(event.X, event.Y) {
		m.activateTabAt(max(0, m.tabStart-1))
		return true
	}
	if more := m.tabMore[1]; more != nil && more.Rect().Contains(event.X, event.Y) {
		m.activateTabAt(min(len(m.contentTabs)-1, m.tabEnd))
		return true
	}
	return false
}

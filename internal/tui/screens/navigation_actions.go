package screens

func (m *Main) toggle() {
	if m.watchFocus {
		return
	}
	if m.navigationFocus {
		m.openNavigation(m.navigationIndex)
		return
	}
	if m.changesFocus && m.findingsVisible() {
		m.openFinding()
		return
	}
	if m.changesFocus && m.overviewVisible() {
		if len(m.visibleFindings()) > 0 {
			m.overviewPreviewOpened = true
		}
		m.selectOverviewFinding()
		return
	}
	if m.activateRulesPane() {
		return
	}
	if m.changesFocus || len(m.rows) == 0 {
		return
	}
	if !m.rows[m.selected].entry.IsDir() {
		m.contentTab, m.contentOpened = true, true
		return
	}
	path := m.rows[m.selected].entry.Path
	m.collapsed[path] = !m.collapsed[path]
	m.flattenRows()
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)*treeRowHeight-viewport(m.tree)))
	m.reveal()
}

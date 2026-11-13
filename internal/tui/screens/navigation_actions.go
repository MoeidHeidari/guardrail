package screens

func (m *Main) activateSpace() {
	if m.rulesVisible() && m.changesFocus && m.rulesFocus == 0 {
		m.toggleRuleSet()
		return
	}
	m.activate()
}

func (m *Main) moveContextBoundary(delta int) bool {
	if delta == -1 && m.selected == 0 {
		m.navigationFocus, m.navigationIndex = true, len(workspaceLabels)-1
		return true
	}
	if delta == 1 && (len(m.rows) == 0 || m.selected == len(m.rows)-1) {
		m.openNavigation(len(workspaceLabels))
		return true
	}
	return false
}

func (m *Main) pageSize() int {
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

func (m *Main) reveal() {
	top := m.selected * treeRowHeight
	if top < m.treeOffset {
		m.treeOffset = top
	} else if top+treeRowHeight > m.treeOffset+viewport(m.tree) {
		m.treeOffset = max(top, top+treeRowHeight-viewport(m.tree))
	}
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)*treeRowHeight-viewport(m.tree)))
}

func (m *Main) toggle() {
	if m.watchFocus {
		m.openWatchFile()
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
		m.openTab(m.rows[m.selected].entry.Path)
		return
	}
	path := m.rows[m.selected].entry.Path
	m.collapsed[path] = !m.collapsed[path]
	m.flattenRows()
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)*treeRowHeight-viewport(m.tree)))
	m.reveal()
}

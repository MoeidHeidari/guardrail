package screens

func (m *Main) toggleRuleSet() {
	if len(m.rulesItems) == 0 {
		return
	}
	if m.collapsedRuleSets == nil {
		m.collapsedRuleSets = make(map[string]bool)
	}
	key := m.rulesItems[m.rulesSelection]
	for i := m.rulesSelection; i >= 0 && key.id != ""; i-- {
		if m.rulesItems[i].set == key.set && m.rulesItems[i].id == "" {
			m.rulesSelection = i
			break
		}
	}
	m.collapsedRuleSets[key.set] = !m.collapsedRuleSets[key.set]
}

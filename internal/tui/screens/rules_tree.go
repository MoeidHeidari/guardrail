package screens

func (m *Main) expandRuleTree(expand bool) {
	if len(m.rulesItems) == 0 {
		return
	}
	if m.collapsedRuleSets == nil {
		m.collapsedRuleSets = make(map[string]bool)
	}
	key := m.rulesItems[m.rulesSelection]
	if key.id != "" {
		if !expand {
			for i := m.rulesSelection - 1; i >= 0; i-- {
				if m.rulesItems[i].set == key.set && m.rulesItems[i].id == "" {
					m.rulesSelection = i
					break
				}
			}
		}
	} else {
		m.collapsedRuleSets[key.set] = !expand
	}
}

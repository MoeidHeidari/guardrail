package screens

import (
	"strings"

	"github.com/grindlemire/go-tui"
)

func (m *Main) footerActions() []string {
	switch {
	case m.searchOpen:
		if m.searchPreviewFocus {
			return []string{"Left results", "[ ] horizontal", "Up/Down scroll", "Esc back"}
		}
		return []string{"Type to search", "Up/Down select", "Enter/Right preview", "Esc cancel"}
	case m.watchFocus:
		return []string{"Enter/Right open", "Up/Down select", "Esc sidebar", "[ ] horizontal"}
	case m.navigationFocus:
		return []string{"Enter/Right open", "Up/Down select"}
	case !m.changesFocus:
		return []string{"Enter/Right open", "Left parent", "Up/Down select", "[ ] horizontal"}
	case m.contentTab:
		return []string{"< > tabs", "x close", "Esc sidebar", "Up/Down scroll", "[ ] horizontal", "r refresh"}
	case m.rulesVisible():
		if m.rulesFocus == 0 {
			return []string{"Space fold", "Left/Right pane", "Up/Down select", "Esc sidebar", "[ ] horizontal"}
		}
		if m.rulesFocus == 1 {
			return []string{"Left/Right pane", "Enter preview", "Up/Down select", "Esc sidebar", "[ ] horizontal"}
		}
		return []string{"Left back", "x close", "Esc sidebar", "Up/Down scroll", "[ ] horizontal"}
	case m.overviewVisible() && m.overviewPreviewFocus:
		return []string{"Esc sidebar", "x close", "Up/Down scroll", "[ ] horizontal"}
	default:
		return []string{"Left/Right filter", "Enter open", "Esc sidebar", "Up/Down select", "[ ] horizontal"}
	}
}

func (m *Main) footerHint(width int) string {
	budget := max(0, width-6-min(62, width/2))
	parts := []string{"q quit"}
	if width >= 55 {
		parts = []string{"Tab focus", "q quit"}
	}
	if width >= 100 {
		parts = []string{"Tab focus", "0-9 tabs", "q quit"}
	}
	actions := append(m.footerActions(), "/ search", "o Overview", "v Violations", "u Rules", "c Context", "w Watch")
	if m.searchOpen {
		parts = nil
		actions = m.footerActions()
	}
	for _, action := range actions {
		candidate := append([]string{action}, parts...)
		if tui.StringWidth(strings.Join(candidate, "   ")) <= budget {
			parts = candidate
		}
	}
	return strings.Join(parts, "   ")
}

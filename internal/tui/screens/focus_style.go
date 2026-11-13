package screens

import "github.com/grindlemire/go-tui"

func (m *Main) highlightFocus(heading *tui.Element, focused bool) {
	if !focused {
		return
	}
	heading.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Selection}),
		tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.SelectionText).Background(m.theme.Selection).Bold()))
}

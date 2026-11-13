package screens

import (
	"fmt"

	"github.com/grindlemire/go-tui"
)

func (m *Main) renderActivity(pane *tui.Element) {
	if len(m.changes) == 0 {
		pane.AddChild(m.label("Waiting for saved changes..."))
		return
	}
	for i := len(m.changes) - 1; i >= 0; i-- {
		change := m.changes[i]
		color, marker := m.theme.Modified, "M"
		switch change.Kind {
		case "created":
			color, marker = m.theme.Added, "A"
		case "deleted":
			color, marker = m.theme.Danger, "D"
		}
		row := m.label("")
		row.Apply(tui.WithRichText(
			tui.TextSpan{Text: change.Time.Format("15:04") + " ", Style: tui.Style{Fg: m.theme.Muted}},
			tui.TextSpan{Text: fmt.Sprintf("%s ", marker), Style: tui.Style{Fg: color}},
			tui.TextSpan{Text: display(change.Path)}))
		row.Apply(tui.WithWidth(tui.StringWidth(change.Time.Format("15:04")+" "+marker+" "+display(change.Path))),
			tui.WithFlexShrink(0))
		pane.AddChild(row)
		if m.watchFocus && len(m.changes)-1-i == m.watchSelection {
			row.Apply(tui.WithRichText(tui.TextSpan{
				Text:  change.Time.Format("15:04") + " " + marker + " " + display(change.Path),
				Style: tui.NewStyle().Foreground(m.theme.SelectionText).Background(m.theme.Selection).Bold(),
			}))
			m.highlightFocus(row, true)
		}
	}
}

func (m *Main) moveWatch(delta int) {
	if delta == -1 && m.watchSelection == 0 {
		m.focusContext()
		return
	}
	m.watchSelection = max(0, min(m.watchSelection+delta, len(m.changes)-1))
	height := max(1, viewport(m.watchPane))
	if m.watchSelection < m.changeOffset {
		m.changeOffset = m.watchSelection
	} else if m.watchSelection >= m.changeOffset+height {
		m.changeOffset = m.watchSelection - height + 1
	}
	m.changeOffset = min(m.changeOffset, max(0, len(m.changes)-height))
}

func (m *Main) openWatchFile() {
	if len(m.changes) == 0 {
		return
	}
	m.watchSelection = min(m.watchSelection, len(m.changes)-1)
	change := m.changes[len(m.changes)-1-m.watchSelection]
	if i, exists := m.entryPositions[change.Path]; exists && m.entries[i].IsDir() {
		return
	}
	m.openTab(change.Path)
	m.returnToTree = true
	m.focusMain()
	m.invalidatePreview()
}

func (m *Main) handleWatchMouse(event tui.MouseEvent) bool {
	if m.watchPane == nil || !m.watchPane.Rect().Contains(event.X, event.Y) {
		return false
	}
	m.watchFocus, m.changesFocus, m.navigationFocus = true, false, false
	switch event.Button {
	case tui.MouseWheelUp, tui.MouseWheelDown:
		delta := 3
		if event.Button == tui.MouseWheelUp {
			delta = -delta
		}
		m.changeOffset = max(0, min(m.changeOffset+delta, max(0, len(m.changes)-viewport(m.watchPane))))
		m.watchSelection = max(m.changeOffset, min(m.watchSelection, m.changeOffset+max(1, viewport(m.watchPane))-1))
	case tui.MouseLeft:
		_, offset := m.watchPane.ScrollOffset()
		index := event.Y - m.watchPane.ContentRect().Y + offset
		if index >= 0 && index < len(m.changes) {
			m.watchSelection = index
			m.openWatchFile()
		}
	}
	return true
}

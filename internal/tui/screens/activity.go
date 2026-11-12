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
		pane.AddChild(row)
	}
}

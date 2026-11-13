package screens

import (
	"strings"

	"github.com/grindlemire/go-tui"
)

func (m *Main) metricBar(count, total, width int, color tui.Color) *tui.Element {
	width = max(1, width)
	filled := 0
	if total > 0 {
		filled = int(float64(width) * float64(max(0, min(count, total))) / float64(total))
	}
	bar := m.label("")
	bar.Apply(tui.WithRichText(
		tui.TextSpan{Text: strings.Repeat("━", filled), Style: tui.Style{Fg: color}},
		tui.TextSpan{Text: strings.Repeat("━", width-filled), Style: tui.Style{Fg: m.theme.Separator}},
	))
	return bar
}

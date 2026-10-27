package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewSection(theme styles.Theme, title string, children ...*tui.Element) *tui.Element {
	section := NewContainer(theme,
		tui.WithPadding(1),
		tui.WithGap(0),
	)
	if title != "" {
		section.AddChild(tui.New(
			tui.WithText(title),
			tui.WithTextStyle(tui.Style{Fg: theme.Muted}),
			tui.WithWrap(false),
			tui.WithTruncate(true),
			tui.WithFlexShrink(0),
		))
	}
	section.AddChild(children...)
	return section
}

package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewLayout(theme styles.Theme, header, sidebar, content *tui.Element, footer ...*tui.Element) *tui.Element {
	root := NewContainer(theme,
		tui.WithWidthPercent(100),
		tui.WithHeightPercent(100),
		tui.WithGap(0),
		tui.WithBackground(tui.Style{Bg: theme.Background}),
	)
	body := NewContainer(theme,
		tui.WithDirection(tui.Row),
		tui.WithFlexGrow(1),
		tui.WithHeight(0),
		tui.WithGap(1),
		tui.WithBackground(tui.Style{Bg: theme.Separator}),
	)
	sidebar.Apply(tui.WithWidth(38), tui.WithFlexShrink(1))
	content.Apply(tui.WithWidthPercent(70), tui.WithFlexGrow(1), tui.WithFlexShrink(0))
	body.AddChild(sidebar, content)
	root.AddChild(header, body)
	root.AddChild(footer...)
	return root
}

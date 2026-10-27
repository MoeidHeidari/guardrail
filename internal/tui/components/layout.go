package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewLayout(theme styles.Theme, header, sidebar, content *tui.Element) *tui.Element {
	root := NewContainer(theme,
		tui.WithWidthPercent(100),
		tui.WithHeightPercent(100),
		tui.WithPadding(1),
		tui.WithGap(1),
		tui.WithBackground(tui.Style{Bg: theme.Background}),
	)
	body := NewContainer(theme,
		tui.WithDirection(tui.Row),
		tui.WithFlexGrow(1),
		tui.WithGap(1),
		tui.WithBackground(tui.Style{Bg: theme.Background}),
	)
	sidebar.Apply(tui.WithWidthPercent(25), tui.WithFlexShrink(0))
	content.Apply(tui.WithWidth(0), tui.WithFlexGrow(1))
	body.AddChild(sidebar, content)
	root.AddChild(header, body)
	return root
}

package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewHeader(theme styles.Theme, title, hint string) *tui.Element {
	header := NewContainer(theme,
		tui.WithHeight(3),
		tui.WithFlexShrink(0),
		tui.WithDirection(tui.Row),
		tui.WithJustify(tui.JustifySpaceBetween),
		tui.WithAlign(tui.AlignCenter),
		tui.WithPaddingTRBL(0, 2, 0, 2),
		tui.WithGap(1),
	)
	header.AddChild(
		tui.New(
			tui.WithText(title),
			tui.WithTextStyle(tui.Style{Fg: theme.Accent}),
			tui.WithWrap(false),
			tui.WithTruncate(true),
			tui.WithFlexShrink(0),
		),
		tui.New(
			tui.WithText(hint),
			tui.WithTextStyle(tui.Style{Fg: theme.Muted}),
			tui.WithWrap(false),
			tui.WithTruncate(true),
			tui.WithMinWidth(0),
		),
	)
	return header
}

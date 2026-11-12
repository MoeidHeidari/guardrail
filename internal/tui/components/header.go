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
		tui.WithPaddingTRBL(1, 2, 1, 2),
		tui.WithGap(1),
		tui.WithBackground(tui.Style{Bg: theme.Navbar}),
	)
	brand := tui.New(tui.WithText(title), tui.WithHeight(1),
		tui.WithWrap(false), tui.WithTruncate(true), tui.WithFlexShrink(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Accent).Bold()))
	context := tui.New(tui.WithText(hint), tui.WithHeight(1),
		tui.WithWrap(false), tui.WithTruncate(true), tui.WithMinWidth(0), tui.WithFlexGrow(1),
		tui.WithTextStyle(tui.Style{Fg: theme.Muted}))
	status := tui.New(tui.WithText("Watching"), tui.WithHeight(1),
		tui.WithWrap(false), tui.WithTruncate(true), tui.WithFlexShrink(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Added)))
	header.AddChild(brand, context, status)
	return header
}

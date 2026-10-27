package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewContainer(theme styles.Theme, options ...tui.Option) *tui.Element {
	defaults := []tui.Option{
		tui.WithDisplay(tui.DisplayFlex),
		tui.WithDirection(tui.Column),
		tui.WithMinWidth(0),
		tui.WithMinHeight(0),
		tui.WithBackground(tui.Style{Bg: theme.Surface}),
		tui.WithTextStyle(tui.Style{Fg: theme.Foreground}),
	}
	return tui.New(append(defaults, options...)...)
}

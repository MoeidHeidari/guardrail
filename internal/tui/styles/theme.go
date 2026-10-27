package styles

import "github.com/grindlemire/go-tui"

type Theme struct {
	Background tui.Color
	Surface    tui.Color
	Foreground tui.Color
	Muted      tui.Color
	Accent     tui.Color
}

func Default() Theme {
	return Theme{
		Background: tui.RGBColor(16, 19, 26),
		Surface:    tui.RGBColor(23, 28, 38),
		Foreground: tui.RGBColor(230, 237, 243),
		Muted:      tui.RGBColor(139, 152, 173),
		Accent:     tui.RGBColor(116, 192, 252),
	}
}

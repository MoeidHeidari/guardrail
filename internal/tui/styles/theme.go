package styles

import (
	"github.com/grindlemire/go-tui"
	"github.com/muesli/termenv"
)

type Theme struct {
	Syntax          string
	Background      tui.Color
	Surface         tui.Color
	Foreground      tui.Color
	Muted           tui.Color
	Accent          tui.Color
	Added           tui.Color
	Modified        tui.Color
	Activity        tui.Color
	Chrome          tui.Color
	Navbar          tui.Color
	Danger          tui.Color
	Selection       tui.Color
	SelectionText   tui.Color
	ActiveSelection tui.Color
	Separator       tui.Color
}

func Default() Theme {
	return Theme{
		Syntax:          "onedark",
		Background:      tui.DefaultColor(),
		Surface:         tui.DefaultColor(),
		Foreground:      tui.DefaultColor(),
		Muted:           tui.White,
		Accent:          tui.Blue,
		Added:           tui.Green,
		Modified:        tui.Yellow,
		Activity:        tui.Magenta,
		Chrome:          tui.DefaultColor(),
		Navbar:          tui.DefaultColor(),
		Danger:          tui.Red,
		Selection:       tui.RGBColor(45, 45, 45),
		SelectionText:   tui.RGBColor(235, 235, 235),
		ActiveSelection: tui.RGBColor(50, 44, 54),
		Separator:       tui.BrightBlack,
	}
}

func Detect() Theme {
	if termenv.NewOutput(nil).HasDarkBackground() {
		return Default()
	}
	return Light()
}

func Light() Theme {
	theme := Default()
	theme.Syntax = "github"
	theme.Muted = tui.BrightBlack
	theme.Separator = tui.White
	theme.Selection = tui.RGBColor(235, 235, 235)
	theme.SelectionText = tui.RGBColor(30, 30, 30)
	theme.ActiveSelection = tui.RGBColor(237, 228, 241)
	return theme
}

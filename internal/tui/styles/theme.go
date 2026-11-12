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
	ActiveSelection tui.Color
	Separator       tui.Color
}

func Default() Theme {
	return Theme{
		Syntax:          "onedark",
		Background:      tui.RGBColor(8, 20, 33),
		Surface:         tui.RGBColor(10, 24, 39),
		Foreground:      tui.RGBColor(215, 220, 229),
		Muted:           tui.RGBColor(137, 145, 161),
		Accent:          tui.RGBColor(110, 190, 255),
		Added:           tui.RGBColor(152, 195, 121),
		Modified:        tui.RGBColor(229, 192, 123),
		Activity:        tui.RGBColor(198, 120, 221),
		Chrome:          tui.RGBColor(8, 20, 33),
		Navbar:          tui.RGBColor(11, 27, 43),
		Danger:          tui.RGBColor(224, 108, 117),
		Selection:       tui.RGBColor(14, 47, 77),
		ActiveSelection: tui.RGBColor(42, 30, 60),
		Separator:       tui.RGBColor(31, 57, 81),
	}
}

func Detect() Theme {
	if termenv.NewOutput(nil).HasDarkBackground() {
		return Default()
	}
	return Light()
}

func Light() Theme {
	return Theme{
		Syntax:          "github",
		Background:      tui.RGBColor(250, 251, 253),
		Surface:         tui.RGBColor(241, 243, 247),
		Foreground:      tui.RGBColor(36, 41, 51),
		Muted:           tui.RGBColor(91, 101, 119),
		Accent:          tui.RGBColor(38, 86, 185),
		Added:           tui.RGBColor(36, 110, 53),
		Modified:        tui.RGBColor(137, 87, 12),
		Activity:        tui.RGBColor(128, 52, 164),
		Chrome:          tui.RGBColor(233, 237, 243),
		Navbar:          tui.RGBColor(228, 233, 242),
		Danger:          tui.RGBColor(187, 45, 59),
		Selection:       tui.RGBColor(220, 230, 249),
		ActiveSelection: tui.RGBColor(237, 222, 248),
		Separator:       tui.RGBColor(211, 217, 227),
	}
}

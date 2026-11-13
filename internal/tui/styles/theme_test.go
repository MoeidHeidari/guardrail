package styles

import (
	"math"
	"testing"

	"github.com/grindlemire/go-tui"
)

func TestThemesUseTerminalDefaults(t *testing.T) {
	for name, theme := range map[string]Theme{"dark": Default(), "light": Light()} {
		t.Run(name, func(t *testing.T) {
			for label, color := range map[string]tui.Color{
				"background": theme.Background, "surface": theme.Surface,
				"foreground": theme.Foreground, "chrome": theme.Chrome, "navbar": theme.Navbar,
			} {
				if !color.IsDefault() {
					t.Errorf("%s must inherit the terminal default", label)
				}
			}
			for label, color := range map[string]tui.Color{
				"muted": theme.Muted, "accent": theme.Accent, "added": theme.Added,
				"modified": theme.Modified, "activity": theme.Activity,
				"danger": theme.Danger, "separator": theme.Separator,
			} {
				if color.Type() != tui.ColorANSI || color.ANSI() > 15 {
					t.Errorf("%s must use the terminal's configurable ANSI palette", label)
				}
			}
			if theme.Syntax == "" {
				t.Fatal("theme must specify syntax palette")
			}
		})
	}
}

func TestLightAndDarkSelectionTints(t *testing.T) {
	dark, light := Default(), Light()
	if dark.Selection.Equal(light.Selection) || dark.ActiveSelection.Equal(light.ActiveSelection) {
		t.Fatal("selection tints must adapt to the detected background")
	}

	if dark.Syntax == light.Syntax {
		t.Fatal("syntax highlighting must adapt to the detected background")
	}
}

func TestSelectionContrast(t *testing.T) {
	luminance := func(color tui.Color) float64 {
		r, g, b := color.RGB()
		linear := func(value uint8) float64 {
			c := float64(value) / 255
			if c <= 0.04045 {
				return c / 12.92
			}
			return math.Pow((c+0.055)/1.055, 2.4)
		}
		return 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
	}
	for _, theme := range []Theme{Default(), Light()} {
		for _, background := range []tui.Color{theme.Selection, theme.ActiveSelection} {
			a, b := luminance(theme.SelectionText), luminance(background)
			if contrast := (max(a, b) + 0.05) / (min(a, b) + 0.05); contrast < 4.5 {
				t.Errorf("selection contrast %.2f must be at least 4.5", contrast)
			}
		}
	}
}

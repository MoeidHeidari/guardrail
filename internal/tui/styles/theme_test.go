package styles

import (
	"math"
	"testing"

	"github.com/grindlemire/go-tui"
)

func luminance(color tui.Color) float64 {
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

func TestThemeContrast(t *testing.T) {
	for name, theme := range map[string]Theme{"dark": Default(), "light": Light()} {
		t.Run(name, func(t *testing.T) {
			for label, color := range map[string]tui.Color{
				"foreground": theme.Foreground, "muted": theme.Muted, "accent": theme.Accent,
				"added": theme.Added, "modified": theme.Modified, "activity": theme.Activity,
				"danger": theme.Danger,
			} {
				for _, bg := range []tui.Color{theme.Background, theme.Surface, theme.Chrome} {
					a, b := luminance(color), luminance(bg)
					contrast := (max(a, b) + 0.05) / (min(a, b) + 0.05)
					if contrast < 4.5 {
						t.Errorf("%s contrast %.2f is below 4.5", label, contrast)
					}
				}
			}
			if theme.Syntax == "" {
				t.Fatal("theme must specify syntax palette")
			}
		})
	}
}

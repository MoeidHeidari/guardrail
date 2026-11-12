package components_test

import (
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func TestHeaderUsesSingleTextRow(t *testing.T) {
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		for _, width := range []int{40, 80, 120} {
			header := components.NewHeader(theme, "GUARDRAIL", "crossview-enterprise / workspace")
			header.RenderTo(tui.NewBuffer(width, 3), width, 3)
			for _, item := range header.Children() {
				if item.Rect().Y != 1 || item.Rect().Height != 1 ||
					!header.ContentRect().ContainsRect(item.Rect()) {
					t.Fatalf("navbar text must share one row inside its padding: %+v", item.Rect())
				}
			}
		}
	}
}

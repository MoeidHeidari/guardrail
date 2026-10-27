package components

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func NewSidebar(theme styles.Theme, children ...*tui.Element) *tui.Element {
	return NewSection(theme, "Workspace", children...)
}

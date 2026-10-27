package screens

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

type Main struct {
	theme styles.Theme
}

func NewMain(theme styles.Theme) *Main {
	return &Main{theme: theme}
}

func (m *Main) Render(app *tui.App) *tui.Element {
	header := components.NewHeader(m.theme, "GUARDRAIL", "q / ctrl+c quit")
	sidebar := components.NewSidebar(m.theme,
		tui.New(

			tui.WithTextStyle(tui.Style{Fg: m.theme.Accent}),
			tui.WithWrap(false),
			tui.WithTruncate(true),
		),
	)
	content := components.NewSection(m.theme, "Overview")
	return components.NewLayout(m.theme, header, sidebar, content)
}

func (m *Main) KeyMap() tui.KeyMap {
	quit := func(event tui.KeyEvent) {
		event.App().Stop()
	}
	return tui.KeyMap{
		tui.On(tui.Rune('q'), quit),
		tui.On(tui.KeyCtrlC, quit),
	}
}

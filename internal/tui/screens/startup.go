package screens

import (
	"fmt"
	"strings"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

type PathPicker struct {
	theme    styles.Theme
	input    *tui.Input
	validate func(string) (string, error)
	selected string
	err      string
}

func NewPathPicker(theme styles.Theme, initial string, validate func(string) (string, error)) *PathPicker {
	picker := &PathPicker{theme: theme, validate: validate}
	picker.input = tui.NewInput(
		tui.WithInputWidth(72),
		tui.WithInputBorder(tui.BorderRounded),
		tui.WithInputFocusColor(theme.Accent),
		tui.WithInputAutoFocus(true),
		tui.WithInputValue(tui.NewState(initial)),
		tui.WithInputOnSubmit(picker.submit),
	)
	return picker
}

func (picker *PathPicker) Selected() string { return picker.selected }

func (picker *PathPicker) submit(value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		picker.err = "Enter a context directory."
		return
	}
	selected, err := picker.validate(value)
	if err != nil {
		picker.err = err.Error()
		return
	}
	picker.selected, picker.err = selected, ""
}

func (picker *PathPicker) Render(app *tui.App) *tui.Element {
	panel := components.NewContainer(picker.theme,
		tui.WithWidth(78), tui.WithGap(1), tui.WithPaddingTRBL(1, 2, 1, 2),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.Style{Fg: picker.theme.Separator}),
		tui.WithFlexShrink(0))
	panel.AddChild(startupBrand(picker.theme, "WATCH SETUP"))
	panel.AddChild(tui.New(tui.WithText("Choose a context"), tui.WithHeight(1),
		tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Accent).Bold())))
	panel.AddChild(tui.New(tui.WithText("Select the project directory Guardrail should watch and analyze."),
		tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Foreground))))
	panel.AddChild(tui.New(tui.WithText("PROJECT DIRECTORY"),
		tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Muted).Bold())))
	var input *tui.Element
	if app != nil {
		input = app.Mount(picker, "context-input", func() tui.Component { return picker.input })
	} else {
		input = picker.input.Render(nil)
	}
	panel.AddChild(input)
	if picker.err != "" {
		panel.AddChild(tui.New(tui.WithText(picker.err), tui.WithWrap(true),
			tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Danger))))
	}
	panel.AddChild(tui.New(tui.WithHR(), tui.WithHeight(1), tui.WithFlexShrink(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Separator))))
	panel.AddChild(startupFooter(picker.theme, "[Enter] Start watching    [Esc] Cancel"))
	panel.AddChild(tui.New(tui.WithText("Or run: guardrail watch <path>"),
		tui.WithTextStyle(tui.NewStyle().Foreground(picker.theme.Muted))))
	root := components.NewContainer(picker.theme, tui.WithFlexGrow(1),
		tui.WithAlign(tui.AlignCenter), tui.WithJustify(tui.JustifyCenter))
	root.AddChild(panel)
	return root
}

func (picker *PathPicker) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnPreemptStop(tui.KeyEscape, func(event tui.KeyEvent) { event.App().Stop() }),
		tui.OnPreemptStop(tui.KeyCtrlC, func(event tui.KeyEvent) { event.App().Stop() }),
	}
}

type SetupDialog struct {
	theme   styles.Theme
	root    string
	message string
}

func NewSetupDialog(theme styles.Theme, root string, err error) *SetupDialog {
	return &SetupDialog{theme: theme, root: root, message: err.Error()}
}

func (dialog *SetupDialog) Render(*tui.App) *tui.Element {
	panel := components.NewContainer(dialog.theme,
		tui.WithWidth(78), tui.WithGap(0), tui.WithPaddingTRBL(1, 2, 1, 2),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.Style{Fg: dialog.theme.Separator}),
		tui.WithFlexShrink(0))
	panel.AddChild(startupBrand(dialog.theme, "FIRST-RUN SETUP"))
	panel.AddChild(tui.New(tui.WithText("Rules are not ready"), tui.WithHeight(1),
		tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Danger).Bold())))
	panel.AddChild(tui.New(tui.WithText("Guardrail needs at least one valid rule set before opening this workspace."),
		tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Foreground))))
	panel.AddChild(tui.New(tui.WithText(fmt.Sprintf("CONTEXT  %s", dialog.root)),
		tui.WithWrap(true), tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Muted))))
	panel.AddChild(startupDivider(dialog.theme))
	problem := components.NewContainer(dialog.theme,
		tui.WithGap(1), tui.WithPaddingTRBL(0, 1, 0, 1),
		tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.Style{Fg: dialog.theme.Danger}))
	problem.AddChild(tui.New(tui.WithText("CONFIGURATION ERROR"),
		tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Danger).Bold())))
	problem.AddChild(tui.New(tui.WithText(dialog.message), tui.WithWrap(true),
		tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Foreground))))
	panel.AddChild(problem)
	panel.AddChild(tui.New(tui.WithText("NEXT STEPS"),
		tui.WithTextStyle(tui.NewStyle().Foreground(dialog.theme.Muted).Bold())))
	panel.AddChild(tui.New(tui.WithText("1. Fix the rule configuration or referenced files under .guard/."),
		tui.WithWrap(true), tui.WithTextStyle(tui.Style{Fg: dialog.theme.Foreground})))
	panel.AddChild(tui.New(tui.WithText("2. Read docs/GETTING_STARTED.md and docs/RULES.md."),
		tui.WithWrap(true), tui.WithTextStyle(tui.Style{Fg: dialog.theme.Accent})))
	panel.AddChild(startupDivider(dialog.theme))
	panel.AddChild(startupFooter(dialog.theme, "[q] Quit    [Esc] Quit"))
	root := components.NewContainer(dialog.theme, tui.WithFlexGrow(1),
		tui.WithAlign(tui.AlignCenter), tui.WithJustify(tui.JustifyCenter))
	root.AddChild(panel)
	return root
}

func startupDivider(theme styles.Theme) *tui.Element {
	return tui.New(tui.WithHR(), tui.WithHeight(1), tui.WithFlexShrink(0),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Separator)))
}

func startupBrand(theme styles.Theme, section string) *tui.Element {
	row := components.NewContainer(theme, tui.WithDirection(tui.Row), tui.WithGap(1), tui.WithHeight(1))
	row.AddChild(tui.New(tui.WithText("GUARDRAIL"),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Accent).Bold())))
	row.AddChild(tui.New(tui.WithText("/ "+section),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Muted))))
	return row
}

func startupFooter(theme styles.Theme, text string) *tui.Element {
	return tui.New(tui.WithText(text), tui.WithHeight(1),
		tui.WithTextStyle(tui.NewStyle().Foreground(theme.Accent).Bold()))
}

func (dialog *SetupDialog) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyEscape, func(event tui.KeyEvent) { event.App().Stop() }),
		tui.OnStop(tui.Rune('q'), func(event tui.KeyEvent) { event.App().Stop() }),
		tui.OnStop(tui.KeyCtrlC, func(event tui.KeyEvent) { event.App().Stop() }),
	}
}

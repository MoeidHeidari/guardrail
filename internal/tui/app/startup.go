package app

import (
	"context"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/tui/screens"
	"github.com/grindlemire/go-tui"
)

func validateRules(root string) error {
	rules, err := config.Load(root)
	if err != nil {
		return err
	}
	_, err = analysis.New(root, rules)
	return err
}

func selectContext(ctx context.Context, terminal *tui.App, picker *screens.PathPicker) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-terminal.StopCh():
			return "", nil
		case event := <-terminal.Events():
			terminal.Dispatch(event)
			if root := picker.Selected(); root != "" {
				return root, nil
			}
			terminal.MarkDirty()
			terminal.Render()
		}
	}
}

func waitForSetup(ctx context.Context, terminal *tui.App) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-terminal.StopCh():
			return nil
		case event := <-terminal.Events():
			terminal.Dispatch(event)
			terminal.MarkDirty()
			terminal.Render()
		}
	}
}

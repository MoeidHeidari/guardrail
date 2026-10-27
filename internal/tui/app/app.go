package app

import (
	"errors"
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/tui/screens"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func Run() (err error) {
	terminal, err := tui.NewApp(
		tui.WithRootComponent(screens.NewMain(styles.Default())),
		tui.WithoutMouse(),
	)
	if err != nil {
		return fmt.Errorf("create terminal workspace: %w", err)
	}
	defer func() {
		if closeErr := terminal.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close terminal workspace: %w", closeErr))
		}
	}()

	if err := terminal.Open(); err != nil {
		return fmt.Errorf("open terminal workspace: %w", err)
	}
	for {
		select {
		case <-terminal.StopCh():
			return nil
		case event := <-terminal.Events():
			terminal.Dispatch(event)
			terminal.Render()
		}
	}
}

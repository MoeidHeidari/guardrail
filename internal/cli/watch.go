package cli

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/app"
	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Open the terminal workspace (no file watching yet)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Run()
		},
	}
}

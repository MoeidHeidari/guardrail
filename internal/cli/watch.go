package cli

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/tui/app"
	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	var contextPath string
	command := &cobra.Command{
		Use:   "watch [context-path]",
		Short: "Watch a context directory in an interactive terminal workspace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if cmd.Flags().Changed("context") {
					return fmt.Errorf("provide either a context argument or --context, not both")
				}
				contextPath = args[0]
			}
			return app.Run(cmd.Context(), contextPath)
		},
	}
	command.Flags().StringVarP(&contextPath, "context", "c", ".", "Context directory (defaults to the current working directory)")
	return command
}

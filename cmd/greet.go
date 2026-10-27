package cmd

import (
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/logger"
	"github.com/MoeidHeidari/guardrail/pkg/greeting"
	"github.com/spf13/cobra"
)

func newGreetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Greet someone by name",
		Long: `Greet prints a customizable greeting message.

Example:
  guardrail greet Alice
  guardrail greet Alice --shout
  guardrail greet --name Bob`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				name = "World"
			}
			shout, _ := cmd.Flags().GetBool("shout")
			msg := greeting.Build(name, shout)
			logger.L().Sugar().Debugf("Generated greeting: %s", msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Name of the person to greet")
	cmd.Flags().Bool("shout", false, "Print the greeting in uppercase")
	return cmd
}

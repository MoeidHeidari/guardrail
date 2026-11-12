package cli

import (
	"encoding/json"
	"fmt"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/spf13/cobra"
)

type ExitError struct {
	Code int
}

func (err *ExitError) Error() string {
	if err.Code == 1 {
		return "check found rule violations"
	}
	return "check could not complete without analysis errors"
}

func newCheckCmd() *cobra.Command {
	var contextPath, format string
	command := &cobra.Command{
		Use: "check [context-path]", Short: "Check text files against .guard/rules.yml rules",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if cmd.Flags().Changed("context") {
					return fmt.Errorf("provide either a context argument or --context, not both")
				}
				contextPath = args[0]
			}
			if format != "text" && format != "json" {
				return fmt.Errorf("format must be text or json")
			}
			root, err := watcher.ResolveRoot(contextPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(root)
			var engine *analysis.Engine
			if err == nil {
				engine, err = analysis.New(root, cfg)
			}
			report := analysis.Report{}
			if err == nil {
				report, err = engine.Check(cmd.Context(), root)
			}
			if err != nil {
				report.Files = append(report.Files, analysis.Result{Path: config.FileName,
					Findings: []analysis.Finding{analysis.ErrorFinding(config.FileName, "configuration", err)}})
			}
			if format == "json" {
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(report); err != nil {
					return err
				}
			} else {
				for _, finding := range report.Findings() {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), finding.String()); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Checked %d files; skipped %d; %d findings\n",
					report.Checked, report.Skipped, len(report.Findings())); err != nil {
					return err
				}
			}
			if code := report.ExitCode(); code != 0 {
				return &ExitError{Code: code}
			}
			return nil
		},
	}
	command.Flags().StringVarP(&contextPath, "context", "c", ".", "Context directory")
	command.Flags().StringVar(&format, "format", "text", "Report format: text or json")
	return command
}

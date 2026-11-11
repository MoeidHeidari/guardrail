package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/cli"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"watch", "--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := executeCommand(args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(output, "watch") {
				t.Fatalf("help does not describe watch: %s", output)
			}
		})
	}
}

func TestRejectUnsupportedCommandsAndFlags(t *testing.T) {
	for _, args := range [][]string{
		{"check"},
		{"greet"},
		{"version"},
		{"rules"},
		{"watch", ".", ".."},
		{"watch", ".", "--context", "."},
		{"watch", "--path", "."},
		{"--config", "config.yaml"},
		{"--verbose"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := executeCommand(args...); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestWatchContextValidation(t *testing.T) {
	for _, args := range [][]string{
		{"watch", "/guardrail/nonexistent-context"},
		{"watch", "--context", "/guardrail/nonexistent-context"},
		{"watch", "-c", "/guardrail/nonexistent-context"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, err := executeCommand(args...)
			if err == nil || !strings.Contains(err.Error(), "resolve context") {
				t.Fatalf("expected context validation before terminal startup, got %v", err)
			}
		})
	}
}

func TestRootCommandsAreIndependent(t *testing.T) {
	first := cli.NewRootCmd()
	second := cli.NewRootCmd()
	first.SetArgs([]string{"watch", "--help"})
	first.SetOut(new(bytes.Buffer))
	if err := first.Execute(); err != nil {
		t.Fatal(err)
	}
	if second.PersistentFlags().Lookup("config") != nil || second.PersistentFlags().Lookup("verbose") != nil {
		t.Fatal("root command has unused configuration flags")
	}
	watch, _, err := second.Find([]string{"watch"})
	if err != nil || watch.Name() != "watch" {
		t.Fatalf("watch command missing: %v", err)
	}
}

package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/cmd"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestGreetDefault(t *testing.T) {
	out, err := executeCommand("greet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Hello, World") {
		t.Errorf("expected 'Hello, World' in output, got: %q", out)
	}
}

func TestGreetWithName(t *testing.T) {
	out, err := executeCommand("greet", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected name 'Alice' in output, got: %q", out)
	}
}

func TestGreetShout(t *testing.T) {
	out, err := executeCommand("greet", "Alice", "--shout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != strings.ToUpper(out) {
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if strings.Contains(line, "ALICE") {
				return
			}
		}
		t.Errorf("expected uppercase greeting, got: %q", out)
	}
}

func TestVersionCmd(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "guardrail") {
		t.Errorf("expected 'guardrail' in version output, got: %q", out)
	}
}

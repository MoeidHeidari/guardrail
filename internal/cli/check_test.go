package cli_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/cli"
)

func TestCheckJSONAndExitStatus(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".guard"), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(".guard/rules.yml", "version: 1\nrules:\n  - id: todo\n    type: regex\n    pattern: TODO\n")
	for _, test := range []struct {
		content string
		code    int
	}{{"package main\n// TODO fix", 1}, {"package main", 0}, {"package {", 2}} {
		write("main.go", test.content)
		output, err := executeCommand("check", "-c", root, "--format", "json")
		var report analysis.Report
		if decodeErr := json.Unmarshal([]byte(output), &report); decodeErr != nil {
			t.Fatalf("invalid JSON: %v / %s", decodeErr, output)
		}
		var exit *cli.ExitError
		if test.code == 0 {
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.As(err, &exit) || exit.Code != test.code {
			t.Fatalf("wrong exit status: %v", err)
		}
		if report.Checked != 1 || report.ExitCode() != test.code {
			t.Fatalf("wrong report: %+v", report)
		}
	}
	output, err := executeCommand("check", t.TempDir(), "--format", "json")
	var report analysis.Report
	if json.Unmarshal([]byte(output), &report) != nil || err == nil || report.ExitCode() != 2 {
		t.Fatalf("missing config must produce error JSON: %s / %v", output, err)
	}
}

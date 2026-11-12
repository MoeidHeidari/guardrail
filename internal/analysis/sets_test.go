package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestSetScopesAttributionAndIgnore(t *testing.T) {
	root := t.TempDir()
	for path, text := range map[string]string{
		".guard/go/a.yml":      "version: 1\ninclude: ['**/*.go']\nrules:\n  - id: todo\n    type: regex\n    pattern: TODO\n",
		".guard/python/a.yaml": "version: 1\ninclude: ['**/*.py']\nrules:\n  - id: todo\n    type: regex\n    pattern: TODO\n",
		"a.go":                 "package main\n// TODO\n", "a.py": "# TODO\n", "ignored.go": "package main\n// TODO\n",
		".guardignore": "ignored.go\n",
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	report, err := engine.Check(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	findings := report.Findings()
	if report.Checked != 2 || len(findings) != 2 || findings[0].Set != "go" ||
		findings[1].Set != "python" || findings[0].Source != ".guard/go/a.yml" {
		t.Fatalf("scope, ignore or attribution mismatch: %+v", findings)
	}
}

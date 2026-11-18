package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestValidateRulesRequiresReadyRuleConfiguration(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".guard"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte("version: 1\nrules: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateRules(root); err == nil {
		t.Fatal("empty rule configuration must block workspace startup")
	}
	rules := "version: 1\nrules:\n  - id: todo\n    type: regex\n    pattern: TODO\n"
	if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(rules), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateRules(root); err != nil {
		t.Fatalf("valid rule configuration must allow workspace startup: %v", err)
	}
}

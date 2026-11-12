package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFolderRuleSets(t *testing.T) {
	root := t.TempDir()
	write := func(path, data string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(".guard/general.yaml", "version: 1\ninclude: ['**/*.go']\nrules:\n  - id: same\n    type: regex\n    pattern: TODO\n")
	write(".guard/python/rules.yml", "version: 1\ninclude: ['**/*.py']\ntimeout_ms: 50\nrules:\n  - id: same\n    type: regex\n    pattern: FIXME\n")
	write("examples/not-loaded.yml", "invalid: true")
	cfg, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Rules) != 2 || cfg.Rules[0].Set != "General" || cfg.Rules[1].Set != "python" ||
		cfg.Rules[1].Source != ".guard/python/rules.yml" || cfg.Rules[1].TimeoutMS != 50 {
		t.Fatalf("incorrect attribution: %+v", cfg.Rules)
	}
	if !cfg.Applies("a.py") || !cfg.Applies("a.go") || cfg.Applies("a.rs") {
		t.Fatal("set scopes not combined")
	}
	write(".guard/python/duplicate.yaml", "version: 1\nrules:\n  - id: same\n    type: regex\n    pattern: x\n")
	if _, err := Load(root); err == nil {
		t.Fatal("duplicate IDs within a set accepted")
	}
}

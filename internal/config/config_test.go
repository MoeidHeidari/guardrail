package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndScope(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".guard"), 0700); err != nil {
		t.Fatal(err)
	}
	data := "version: 1\nrules:\n  - id: todo\n    type: regex\n    pattern: TODO\n"
	if err := os.WriteFile(filepath.Join(root, FileName), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".guard", "examples"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".guardrail.yml", ".guard/examples/inline.txt"} {
		if err := os.WriteFile(filepath.Join(root, path), []byte("invalid: configuration"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TimeoutMS != 100 || !cfg.Applies("main.go") || !cfg.Applies("src/main.ts") {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	for _, path := range []string{".git/config", FileName, ".guard/rules/check.js", ".guardrail/rule.js", "node_modules/a.js", "src/vendor/a.go", "target/a.rs", "image.png"} {
		if cfg.Applies(path) {
			t.Fatalf("default scope includes %s", path)
		}
	}
}

func TestInvalidConfiguration(t *testing.T) {
	valid := "version: 1\nrules:\n  - id: test\n    type: regex\n    pattern: TODO\n"
	cases := []string{
		"", "version: 2\nrules: []\n", valid + "unknown: true\n",
		valid + "---\nversion: 1\n", strings.Replace(valid, "TODO", "''", 1),
		valid + "  - id: test\n    type: regex\n    pattern: FIXME\n",
		strings.Replace(valid, "regex", "python", 1),
		valid + "include: ['[']\n", valid + "timeout_ms: -1\n",
		"version: 1\nrules:\n  - id: js\n    type: javascript\n",
		"version: 1\nrules:\n  - id: js\n    type: javascript\n    code: x\n    file: .guardrail/a.js\n",
	}
	for _, data := range cases {
		t.Run(data, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, ".guard"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, FileName), []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(root); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("missing config must not silently pass")
	}
}

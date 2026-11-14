package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func testEngine(t *testing.T, rule config.Rule) *Engine {
	t.Helper()
	engine, err := New(t.TempDir(), config.Config{Include: []string{"**"}, TimeoutMS: 30, Rules: []config.Rule{rule}})
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func TestRegexLocations(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "todo", Type: "regex", Pattern: "TODO", Message: "remove TODO"})
	result := engine.Evaluate(context.Background(), "a.txt", []byte("é TODO\nTODO"))
	if len(result.Findings) != 2 {
		t.Fatalf("unexpected findings: %+v", result)
	}
	first, second := result.Findings[0], result.Findings[1]
	if first.Line != 1 || first.Column != 3 || second.Line != 2 || second.Column != 1 || first.Message != "remove TODO" {
		t.Fatalf("invalid Unicode positions: %+v", result.Findings)
	}
}

func TestJavaScriptContract(t *testing.T) {
	cases := []struct {
		code string
		want int
		kind string
	}{
		{`(path, change) => path === "a.txt" && change.content.includes("bad")`, 1, "violation"},
		{`() => false`, 0, ""},
		{`() => []`, 0, ""},
		{`() => [{line: 2, column: 2, message: "bad"}]`, 1, "violation"},
		{`() => { throw Error("broken"); }`, 1, "error"},
		{`() => undefined`, 1, "error"},
		{`() => "false"`, 1, "error"},
		{`() => ({line: 1, column: 1})`, 1, "error"},
		{`async () => false`, 1, "error"},
		{`() => [{line: 20, column: 1}]`, 1, "error"},
		{`() => [{line: 1, column: 50}]`, 1, "error"},
		{`() => [{line: 1.1, column: 1}]`, 1, "error"},
		{`() => [{line: 1, column: 1, extra: true}]`, 1, "error"},
		{`() => { while (true) {} }`, 1, "error"},
		{`(() => {while (true) {} })()`, 1, "error"},
		{`() => [{ get line() {while (true) {}} }]`, 1, "error"},
		{`() => [null]`, 1, "error"},
		{`() => typeof require !== "undefined" || typeof process !== "undefined" || typeof fetch !== "undefined"`, 0, ""},
	}
	for _, test := range cases {
		t.Run(test.code, func(t *testing.T) {
			engine := testEngine(t, config.Rule{ID: "js", Type: "javascript", Code: test.code})
			start := time.Now()
			result := engine.Evaluate(context.Background(), "a.txt", []byte("bad\nok"))
			if time.Since(start) > time.Second {
				t.Fatal("JavaScript execution was not bounded")
			}
			if len(result.Findings) != test.want {
				t.Fatalf("unexpected findings: %+v", result.Findings)
			}
			if test.want > 0 && result.Findings[0].Kind != test.kind {
				t.Fatalf("unexpected finding kind: %+v", result.Findings)
			}
		})
	}
}

func TestScriptsAndContainment(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".guardrail"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".guardrail/rule.js"), []byte(`(path, change) => true`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{TimeoutMS: 30, Include: []string{"**"}, Rules: []config.Rule{{ID: "js", Type: "javascript", File: ".guardrail/rule.js"}}}
	engine, err := New(root, cfg)
	if err != nil || len(engine.Evaluate(context.Background(), "a.txt", []byte("x")).Findings) != 1 {
		t.Fatalf("file function did not run: %v", err)
	}
	for _, path := range []string{"../rule.js", "rule.js", ".guardrail/../rule.js", "/tmp/rule.js"} {
		cfg.Rules[0].File = path
		if _, err := New(root, cfg); err == nil {
			t.Fatalf("script escape accepted: %s", path)
		}
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "rule.js"), []byte(`()=>true`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "rule.js"), filepath.Join(root, ".guardrail/link.js")); err != nil {
		t.Fatal(err)
	}
	cfg.Rules[0].File = ".guardrail/link.js"
	if _, err := New(root, cfg); err == nil {
		t.Fatal("symlink script accepted")
	}
}

func TestReadLimitsAndBadPattern(t *testing.T) {
	root := t.TempDir()
	for _, content := range []string{strings.Repeat("a", ContentLimit+1), "a\x00b", "\xff"} {
		if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(root, "a.txt"); err == nil {
			t.Fatal("unsupported file accepted")
		}
	}
	if _, err := New(root, config.Config{Rules: []config.Rule{{ID: "bad", Type: "regex", Pattern: "("}}}); err == nil {
		t.Fatal("bad regex accepted")
	}
}

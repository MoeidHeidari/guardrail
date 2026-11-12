package analysis

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestProjectExampleRules(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	foundLength := false
	for _, rule := range cfg.Rules {
		if rule.ID == "max-source-lines" {
			foundLength = rule.Set == "maintainability" && rule.File == ".guard/rules/max-source-lines.js"
		}
	}
	if !foundLength {
		t.Fatal("maintainability set must reference the project length script")
	}
	result := engine.Evaluate(context.Background(), "example.go", []byte("package main\n<<<<<<< ours\n=======\n>>>>>>> theirs\n"))
	found := 0
	for _, finding := range result.Findings {
		if finding.Rule == "merge-conflict-markers" {
			found++
		}

	}
	if found != 3 {
		t.Fatalf("merge-marker regex must report all three lines: %+v", result.Findings)
	}
}

func TestFileLengthThreshold(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "length", Type: "javascript", Code: `(path, content) => {
		const count = content.split("\n").length - (content.endsWith("\n") ? 1 : 0);
		return count > 300 ? [{line: 301, column: 1, message: "too long"}] : false;
	}`})
	for _, count := range []int{299, 300, 301} {
		result := engine.Evaluate(context.Background(), "a.txt", []byte(strings.Repeat("line\n", count)))
		if count <= 300 && len(result.Findings) != 0 {
			t.Fatalf("%d lines must pass", count)
		}
		if count == 301 && (len(result.Findings) != 1 || result.Findings[0].Line != 301) {
			t.Fatalf("wrong threshold or position: %+v", result.Findings)
		}
	}
}

func TestProjectFileLengthBoundary(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{99, 100, 101, 109, 110} {
		for _, trailingNewline := range []bool{false, true} {
			content := "package main\n" + strings.Repeat("// source\n", count-1)
			if !trailingNewline {
				content = strings.TrimSuffix(content, "\n")
			}
			result := engine.Evaluate(context.Background(), "example.go", []byte(content))
			var findings []Finding
			for _, finding := range result.Findings {
				if finding.Kind == "error" {
					t.Fatalf("%d lines produced an execution error: %+v", count, finding)
				}
				if finding.Rule == "max-source-lines" {
					findings = append(findings, finding)
				}
			}
			if count <= 100 {
				if len(findings) != 0 {
					t.Fatalf("%d lines must pass: %+v", count, findings)
				}
			} else if len(findings) != 1 || findings[0].Line != 101 ||
				findings[0].Column != 1 || !strings.Contains(findings[0].Message, "limit is 100") {
				t.Fatalf("%d lines must have a valid, consistent finding: %+v", count, findings)
			}
		}
	}
}

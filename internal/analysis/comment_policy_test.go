package analysis

import (
	"context"
	"testing"
)

func TestCommunityCommentPolicy(t *testing.T) {
	engine := communityPackEngine(t, "no-comments")
	for _, test := range []struct {
		path, content string
		want          int
	}{
		{"a.go", "package main\n// ordinary\nfunc private() {}\n", 1},
		{"a.go", "package main\n// Public explains the function.\n// More documentation.\nfunc Public() {}\n", 0},
		{"a.go", "//go:build linux\n\npackage main\n", 0},
		{"a.go", "package main\n//nolint:unused\nfunc private() {}\n", 0},
		{"a.go", "package main\n/* C preamble */\nimport \"C\"\n", 0},
		{"a.go", "package main\n// Public type.\ntype Public struct{}\n", 0},
		{"a.go", "package main\nvar text = \"// ordinary\"\n", 0},
		{"a.js", "// ordinary\nconst x = 1;", 1},
		{"a.js", "/** Documentation. */\nexport function f() {}", 0},
		{"a.js", "// eslint-disable-next-line\nconst x = 1;", 0},
		{"a.ts", "// @ts-expect-error\nconst x: number = 'a';", 0},
		{"a.tsx", "const x = <div>{/* ordinary */}</div>;", 1},
		{"a.py", "# ordinary\nx = 1\n", 1},
		{"a.py", "#!/usr/bin/python\n# coding: utf-8\ndef f():\n    \"\"\"Docs.\"\"\"\n    pass\n", 0},
		{"a.rs", "// ordinary\nfn f() {}\n", 1},
		{"a.rs", "/// Documentation.\npub fn f() {}\n", 0},
		{"a.rs", "/* ordinary */\nfn f() {}\n", 1},
	} {
		t.Run(test.path, func(t *testing.T) {
			result := engine.Evaluate(context.Background(), test.path, []byte(test.content))
			if len(result.Findings) != test.want {
				t.Fatalf("expected %d findings: %+v", test.want, result.Findings)
			}
			for _, finding := range result.Findings {
				if finding.Kind != "violation" || finding.Rule != "no-comments" {
					t.Fatalf("unexpected error: %+v", finding)
				}
			}
		})
	}
}

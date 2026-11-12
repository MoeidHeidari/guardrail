package analysis

import (
	"context"
	"testing"
)

func TestTreeSitterLanguages(t *testing.T) {
	for path, content := range map[string]string{
		"a.go": "package main\nfunc main() {}", "a.js": "const a = 1;",
		"a.ts": "const a: number = 1;", "a.tsx": "const a = <div/>;",
		"a.py": "def f():\n    return 1\n", "a.rs": "fn main() {}",
	} {
		t.Run(path, func(t *testing.T) {
			syntax, err := Parse(context.Background(), path, []byte(content))
			if err != nil || !syntax.Parsed || syntax.Language == "text" {
				t.Fatalf("grammar did not parse %s: %+v %v", path, syntax, err)
			}
		})
	}
	syntax, err := Parse(context.Background(), "a.go", []byte("func {"))
	if err == nil || !syntax.Parsed {
		t.Fatal("malformed syntax was not reported")
	}
	syntax, err = Parse(context.Background(), "a.yaml", []byte("a: b"))
	if err != nil || syntax.Parsed || syntax.Language != "text" {
		t.Fatal("unsupported grammar must be explicitly text-only")
	}
}

package analysis

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineeringPacks(t *testing.T) {
	for _, test := range []struct {
		pack, path, content string
		line, column        int
	}{
		{"architecture", "domain/a.go", "package domain\nimport \"app/infrastructure/db\"\n", 2, 8},
		{"architecture", "domain/a.js", "import db from '../infrastructure/db';", 1, 16},
		{"architecture", "core/a.ts", "import { db } from '../adapters/db';", 1, 20},
		{"architecture", "domain/a.tsx", "import db from '../transport/db';", 1, 16},
		{"architecture", "core/a.py", "from infrastructure.db import load\n", 1, 6},
		{"architecture", "domain/a.rs", "use crate::infrastructure::db;\n", 1, 5},
		{"architecture", "domain/alias.py", "import infrastructure.db as database\n", 1, 8},
		{"architecture", "domain/group.rs", "use crate::{infrastructure::db, domain::model};\n", 1, 5},
		{"architecture", "domain/require.js", "const db = require('../infrastructure/db');", 1, 20},
		{"architecture", "domain/reexport.ts", "export { db } from '../infrastructure/db';", 1, 20},
		{"architecture", "domain/dynamic.js", "import('../infrastructure/db');", 1, 8},
		{"source-policy", "a.go", "package main\nfunc f() { fmt.Println(\"debug\") }\n", 2, 12},
		{"source-policy", "a.js", "console.log('debug');", 1, 1},
		{"source-policy", "a.ts", "console.debug('debug');", 1, 1},
		{"source-policy", "a.tsx", "const a = <div/>;\nconsole.trace(a);", 2, 1},
		{"source-policy", "a.py", "print('debug')\n", 1, 1},
		{"source-policy", "a.rs", "fn main() { dbg!(1); }\n", 1, 13},
		{"source-policy", "comment.go", "package main\nfunc f() { fmt /* note */ . Println(1) }\n", 2, 12},
		{"source-policy", "comment.js", "console /* note */ . log(1);", 1, 1},
		{"source-policy", "debugger.js", "debugger;", 1, 1},
		{"api-policy", "a.go", "package main\nfunc Public() {}\n", 2, 6},
		{"api-policy", "a.js", "export function publicFn() {}", 1, 17},
		{"api-policy", "a.ts", "export function publicFn(): void {}", 1, 17},
		{"api-policy", "a.tsx", "export function Public() { return <div/>; }", 1, 17},
		{"api-policy", "a.py", "def public_fn():\n    pass\n", 1, 5},
		{"api-policy", "a.rs", "pub fn public_fn() {}\n", 1, 8},
		{"complexity", "a.go", "package main\nfunc f(a, b, c, d, e, f int) {}\n", 2, 1},
		{"complexity", "a.js", "function f(a,b,c,d,e,f) {}", 1, 1},
		{"complexity", "a.ts", "function f(a:number,b:number,c:number,d:number,e:number,f:number) {}", 1, 1},
		{"complexity", "a.tsx", "const f = (a,b,c,d,e,f) => <div/>;", 1, 11},
		{"complexity", "a.py", "def f(a,b,c,d,e,f):\n    pass\n", 1, 1},
		{"complexity", "a.rs", "fn f(a:i32,b:i32,c:i32,d:i32,e:i32,f:i32) {}\n", 1, 1},
		{"complexity", "closure.go", "package main\nvar f = func(a,b,c,d,e,f int) {}\n", 2, 9},
		{"complexity", "lambda.py", "f = lambda a,b,c,d,e,f: a\n", 1, 5},
		{"complexity", "closure.rs", "fn main() { let f = |a,b,c,d,e,f| a; }\n", 1, 21},
		{"dependency-policy", "package.json", `{"dependencies":{"left-pad":"1.0.0"}}`, 1, 1},
		{"dependency-policy", "package.json", `{"devDependencies":{"alias":"npm:request@2"}}`, 1, 1},
	} {
		t.Run(test.pack+"/"+test.path, func(t *testing.T) {
			engine := communityPackEngine(t, test.pack)
			result := engine.Evaluate(context.Background(), test.path, []byte(test.content))
			if len(result.Findings) != 1 {
				_, nodes, _ := parseSource(context.Background(), test.path, []byte(test.content), true)
				t.Logf("syntax: %+v", nodes)
				t.Fatalf("expected one finding: %+v", result.Findings)
			}

			got := result.Findings[0]
			language := map[string]string{".go": "go", ".js": "javascript", ".ts": "typescript",
				".tsx": "typescript", ".py": "python", ".rs": "rust", ".json": "javascript"}[filepath.Ext(test.path)]
			if got.Kind != "violation" || got.Set != language || got.Line != test.line || got.Column != test.column {
				t.Fatalf("incorrect finding: %+v", got)
			}
			if test.pack == "complexity" && !strings.Contains(got.Message, "6 parameters") {
				t.Fatalf("incorrect parameter count: %s", got.Message)
			}
		})
	}
}

func TestEngineeringPackScopeAndPrivateDeclarations(t *testing.T) {
	engine := communityPackEngine(t, "architecture")
	result := engine.Evaluate(context.Background(), "infrastructure/a.go",
		[]byte("package infrastructure\nimport \"app/infrastructure/db\"\n"))
	if len(result.Findings) != 0 {
		t.Fatalf("non-domain code should not be restricted: %+v", result.Findings)
	}
	engine = communityPackEngine(t, "api-policy")
	for path, content := range map[string]string{
		"a.go": "package main\nfunc private() {}\n",
		"a.js": "function privateFn() {}",
		"a.py": "def _private():\n    pass\n",
		"a.rs": "pub(crate) fn internal() {}\n",
	} {
		result := engine.Evaluate(context.Background(), path, []byte(content))
		if len(result.Findings) != 0 {
			t.Fatalf("private declarations should pass: %+v", result.Findings)
		}
	}
}

func TestEngineeringPackCleanCode(t *testing.T) {
	sources := map[string]string{
		"a.go":  "package main\n// Public documents the function.\nfunc Public() { _ = \"fmt.Println(1)\" }\n",
		"a.js":  "/** Docs. */\nexport function f() { const text = 'console.log(1)'; }\n",
		"a.ts":  "/** Docs. */\nexport function f(): void { const text = 'console.log(1)'; }\n",
		"a.tsx": "/** Docs. */\nexport function F() { return <div>console.log(1)</div>; }\n",
		"a.py":  "def public_fn():\n    \"\"\"Documentation.\"\"\"\n    value = 'print(1)'\n",
		"a.rs":  "/// Documentation.\npub fn public_fn() { let value = \"dbg!(1)\"; }\n",
	}
	for _, pack := range []string{"architecture", "source-policy", "api-policy", "complexity"} {
		engine := communityPackEngine(t, pack)
		for path, content := range sources {
			t.Run(pack+"/"+path, func(t *testing.T) {
				result := engine.Evaluate(context.Background(), "domain/"+path, []byte(content))
				if len(result.Findings) != 0 {
					t.Fatalf("clean syntax must pass: %+v", result.Findings)
				}
			})
		}
	}
}

func TestEngineeringBudgetBoundaries(t *testing.T) {
	engine := communityPackEngine(t, "complexity")
	for path, content := range map[string]string{
		"a.go": "package main\nfunc f(a,b,c,d,e int) {}\n",
		"a.js": "function f(a,b,c,d,e) {}",
		"a.ts": "function f(a:number,b:number,c:number,d:number,e:number) {}",
		"a.py": "def f(a,b,c,d,e):\n    pass\n",
		"a.rs": "fn f(a:i32,b:i32,c:i32,d:i32,e:i32) {}\n",
	} {
		result := engine.Evaluate(context.Background(), path, []byte(content))
		if len(result.Findings) != 0 {
			t.Fatalf("five parameters must pass for %s: %+v", path, result.Findings)
		}
	}
	for _, depth := range []int{3, 4} {
		content := "package main\nfunc f() {\n" + strings.Repeat("if true {\n", depth) +
			strings.Repeat("}\n", depth) + "}\n"
		result := engine.Evaluate(context.Background(), "a.go", []byte(content))
		if len(result.Findings) != depth-3 {
			t.Fatalf("depth %d: %+v", depth, result.Findings)
		}
	}
	for _, lines := range []int{60, 61} {
		content := "package main\nfunc f() {\n" + strings.Repeat("// line\n", lines-2) + "}\n"
		result := engine.Evaluate(context.Background(), "a.go", []byte(content))
		if len(result.Findings) != lines-60 {
			t.Fatalf("%d function lines: %+v", lines, result.Findings)
		}
	}
}

func TestDependencyPolicyErrors(t *testing.T) {
	engine := communityPackEngine(t, "dependency-policy")
	for _, content := range []string{`{`, `[]`, `{"dependencies":[]}`, `{"dependencies":{"name":42}}`} {
		result := engine.Evaluate(context.Background(), "package.json", []byte(content))
		if len(result.Findings) != 1 || result.Findings[0].Kind != "error" {
			t.Fatalf("invalid manifest must fail explicitly: %+v", result.Findings)
		}
	}
	result := engine.Evaluate(context.Background(), "package.json", []byte(`{"dependencies":{"approved":"1.0.0"}}`))
	if len(result.Findings) != 0 {
		t.Fatalf("approved manifest must pass: %+v", result.Findings)
	}
}

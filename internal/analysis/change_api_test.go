package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestChangeSourceAndLocations(t *testing.T) {
	for _, test := range []struct {
		content string
		lines   int
	}{
		{"", 0}, {"hello", 1}, {"hello\n", 1}, {"hello\r\nworld\r\n", 2}, {"\n\n", 2},
	} {
		engine := testEngine(t, config.Rule{ID: "source", Type: "javascript", Code: `(path, change) => {
			if (path !== "a.txt" || change.version !== 1 || change.language !== "text") throw Error("metadata");
			if (!Object.isFrozen(change) || !Object.isFrozen(change.lines)) throw Error("mutable source");
			if (change.lines.some(line => line.endsWith("\r"))) throw Error("CRLF");
			return change.lineCount !== ` + strconv.Itoa(test.lines) + ` || change.lines.length !== change.lineCount;
		}`})
		result := engine.Evaluate(context.Background(), "a.txt", []byte(test.content))
		if len(result.Findings) != 0 {
			t.Fatalf("source %q: %+v", test.content, result.Findings)
		}
	}
	engine := testEngine(t, config.Rule{ID: "position", Type: "javascript", Code: `(path, change) => {
		const call = change.calls()[0];
		const location = call.location;
		if (location.start.offset !== 17 || location.start.column !== 17 ||
		    location.end.column !== 31 || location.end.offset !== 31) throw Error(JSON.stringify(location));
		if (change.content.slice(location.start.offset, location.end.offset) !== call.text) throw Error("slice");
		if (!Object.isFrozen(call) || !Object.isFrozen(location) || !Object.isFrozen(location.start)) throw Error("mutable model");
		if (change.ast.field(call.rawNode, "function") !== call.target.rawNode) throw Error("raw interop");
		return [call.finding("call")];
	}`})
	result := engine.Evaluate(context.Background(), "a.js", []byte(`const s = "é😀"; console.log(s);`))
	if len(result.Findings) != 1 || result.Findings[0].Kind != "violation" || result.Findings[0].Column != 17 {
		t.Fatalf("incorrect helper location: %+v", result.Findings)
	}
}

func TestChangeCapabilitiesAndErrors(t *testing.T) {
	for _, code := range []string{
		`(path, change) => change.functions()`,
		`(path, change) => change.require("types")`,
		`(path, change) => change.finding("bad", {line:0,column:1})`,
		`(path, change) => change.finding("", {line:1,column:1})`,
	} {
		engine := testEngine(t, config.Rule{ID: "error", Type: "javascript", Code: code})
		result := engine.Evaluate(context.Background(), "a.txt", []byte("text"))
		if len(result.Findings) != 1 || result.Findings[0].Kind != "error" {
			t.Fatalf("unsupported/invalid capability must error: %+v", result.Findings)
		}
	}
	engine := testEngine(t, config.Rule{ID: "syntax", Type: "javascript", Code: `(path, change) => change.calls()`})
	result := engine.Evaluate(context.Background(), "a.go", []byte("func {"))
	if len(result.Findings) != 2 || result.Findings[1].Kind != "error" {
		t.Fatalf("invalid syntax must not pass helpers: %+v", result.Findings)
	}
}

func TestChangeFunctionModels(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "functions", Type: "javascript", Code: `(path, change) => {
		change.require("functions", "imports", "comments", "calls");
		const fn = change.functions()[0];
		if (fn.metrics.parameterCount !== 2 || fn.metrics.nestingDepth !== 1 ||
		    !fn.body || fn.isPublic !== true || !fn.documentation) throw Error(JSON.stringify(fn));
		if (!Object.isFrozen(fn.parameters) || !Object.isFrozen(fn.metrics) ||
		    change.functions() !== change.functions()) throw Error("mutable/uncached");
		return [];
	}`})
	for path, content := range map[string]string{
		"a.go":  "package p\n// Public documents it.\nfunc Public(a,b int) { if true {} }\n",
		"a.js":  "/** Docs. */\nexport function f(a,b) { if (a) {} }",
		"a.ts":  "/** Docs. */\nexport function f(a:number,b:number) { if (a) {} }",
		"a.tsx": "/** Docs. */\nexport function F(a,b) { if (a) {} return <div/>; }",
		"a.py":  "def f(a,b):\n    \"\"\"Docs.\"\"\"\n    if a:\n        pass\n",
		"a.rs":  "/// Docs.\npub fn f(a:i32,b:i32) { if true {} }\n",
	} {
		result := engine.Evaluate(context.Background(), path, []byte(content))
		if len(result.Findings) != 0 {
			t.Fatalf("%s: %+v", path, result.Findings)
		}
	}
}

func TestChangeImportModels(t *testing.T) {
	for _, test := range []struct{ path, content, source string }{
		{"a.go", "package p\nimport db \"app/domain\"\n", "app/domain"},
		{"a.js", "import db from './domain';", "./domain"},
		{"a.ts", "export { db } from './domain';", "./domain"},
		{"a.py", "import domain.db as db\n", "domain.db"},
		{"a.rs", "use crate::domain::db;\n", "crate::domain::db"},
	} {
		engine := testEngine(t, config.Rule{ID: "import", Type: "javascript", Code: `(path, change) => {
			const imports = change.imports();
			return imports.length !== 1 || imports[0].source !== "` + test.source + `" ||
				imports[0].isDynamic || !imports[0].sourceNode;
		}`})
		result := engine.Evaluate(context.Background(), test.path, []byte(test.content))
		if len(result.Findings) != 0 {
			t.Fatalf("%s: %+v", test.path, result.Findings)
		}
	}
	engine := testEngine(t, config.Rule{ID: "dynamic", Type: "javascript", Code: `(path, change) => {
		const entry = change.imports()[0];
		return !entry.isDynamic || entry.source !== null || entry.form !== "import";
	}`})
	if result := engine.Evaluate(context.Background(), "a.js", []byte("import(variable);")); len(result.Findings) != 0 {
		t.Fatalf("unknown dynamic source must be explicit: %+v", result.Findings)
	}
}

func TestASTAndHelperRuleParity(t *testing.T) {
	root := filepath.Join("..", "..", "rulepacks", "functions", "source-policy")
	for path, content := range map[string]string{
		"a.go": "package p\nfunc f() { fmt /* note */ . Println(1) }\n",
		"a.js": "console.log(1); debugger; const s = 'console.log(1)';",
		"a.ts": "console.debug(1);",
		"a.py": "print(1)\n",
		"a.rs": "fn f() { dbg!(1); }\n",
	} {
		var previous []Finding
		for _, file := range []string{"no-debug-output.js", "no-debug-output-ast.js"} {
			code, err := os.ReadFile(filepath.Join(root, file))
			if err != nil {
				t.Fatal(err)
			}
			engine := testEngine(t, config.Rule{ID: "debug", Type: "javascript", Code: string(code)})
			result := engine.Evaluate(context.Background(), path, []byte(content))
			for _, finding := range result.Findings {
				if finding.Kind == "error" {
					t.Fatalf("rule failed: %+v", finding)
				}
			}
			if previous != nil {
				if len(previous) != len(result.Findings) {
					t.Fatalf("%s: variants differ", path)
				}
				for i := range previous {
					if previous[i] != result.Findings[i] {
						t.Fatalf("%s: variants differ: %+v vs %+v", path, previous, result.Findings)
					}
				}
			}
			previous = result.Findings
		}
	}
}

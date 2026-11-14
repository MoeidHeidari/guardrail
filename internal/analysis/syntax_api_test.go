package analysis

import (
	"context"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestSyntaxAPILocationsAndReadOnlyNodes(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "ast", Type: "javascript", Code: `(path, content, syntax) => {
		syntax.requireAST();
		if (syntax.version !== 1 || syntax.language !== "JavaScript") throw Error("wrong API");
		const call = syntax.find("call_expression")[0];
		const target = syntax.field(call, "function");
		if (syntax.text(target) !== "console.log") throw Error("wrong UTF-16 offsets");
		if (!Object.isFrozen(syntax) || !Object.isFrozen(syntax.nodes) ||
			!Object.isFrozen(call) || !Object.isFrozen(call.fields) ||
			!Object.isFrozen(call.children)) throw Error("mutable API");
		if (syntax.parent(target) !== call) throw Error("wrong parent");
		return [{line:call.line,column:call.column,message:"call"}];
	}`})
	result := engine.Evaluate(context.Background(), "a.js", []byte(`const s = "é😀"; console.log(s);`))
	if len(result.Findings) != 1 || result.Findings[0].Kind != "violation" || result.Findings[0].Column != 17 {
		t.Fatalf("invalid syntax API position: %+v", result.Findings)
	}
}

func TestSyntaxAPIErrors(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "ast", Type: "javascript",
		Code: `(path, content, syntax) => { syntax.requireAST(); return []; }`})
	for _, test := range []struct{ path, content string }{{"a.txt", "plain"}, {"a.go", "func {"}} {
		result := engine.Evaluate(context.Background(), test.path, []byte(test.content))
		last := result.Findings[len(result.Findings)-1]
		if last.Kind != "error" || last.Rule != "ast" {
			t.Fatalf("unavailable syntax must not appear successful: %+v", result.Findings)
		}
	}
	engine = testEngine(t, config.Rule{ID: "ast", Type: "javascript",
		Code: `(path, content, syntax) => syntax.text({start:0,end:1})`})
	result := engine.Evaluate(context.Background(), "a.js", []byte("const x = 1;"))
	if len(result.Findings) != 1 || result.Findings[0].Kind != "error" {
		t.Fatalf("foreign nodes must be rejected: %+v", result.Findings)
	}
}

func TestSyntaxCaptureLimitAndCancellation(t *testing.T) {
	content := []byte(strings.Repeat("const x = 1;\n", 15000))
	_, nodes, err := parseSource(context.Background(), "a.js", content, true)
	if err == nil || !strings.Contains(err.Error(), "named nodes") || nodes != nil {
		t.Fatalf("oversized tree must fail explicitly: nodes=%d err=%v", len(nodes), err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := parseSource(ctx, "a.go", []byte("package main"), true); err != context.Canceled {
		t.Fatalf("cancellation not preserved: %v", err)
	}
}

func TestSyntaxSnapshotIsolationAndRuleScope(t *testing.T) {
	engine, err := New(t.TempDir(), config.Config{Include: []string{"**"}, TimeoutMS: 100,
		Rules: []config.Rule{
			{ID: "mutate", Type: "javascript", Code: `(path, content, syntax) => {
				syntax.nodes[0].type = "modified"; syntax.nodes = [];
				return [];
			}`},
			{ID: "verify", Type: "javascript", Code: `(path, content, syntax) => {
				if (syntax.nodes[0].type !== "program") throw Error("shared mutation");
				return [];
			}`},
		}})
	if err != nil {
		t.Fatal(err)
	}
	result := engine.Evaluate(context.Background(), "a.js", []byte("const x = 1;"))
	if len(result.Findings) != 1 || result.Findings[0].Rule != "mutate" || result.Findings[0].Kind != "error" {
		t.Fatalf("mutation must fail without affecting the following rule: %+v", result.Findings)
	}
	engine = testEngine(t, config.Rule{ID: "excluded", Type: "javascript",
		Include: []string{"**/*.py"}, Code: `() => true`})
	engine.Config.TimeoutMS = 1000
	result = engine.Evaluate(context.Background(), "a.js", []byte(strings.Repeat("const x = 1;\n", 15000)))
	if len(result.Findings) != 0 {
		t.Fatalf("excluded JS rules must not trigger AST capture limits: %+v", result.Findings)
	}
}

func TestProjectDebugRuleUsesSyntax(t *testing.T) {
	root := "../.."
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := New(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		content string
		want    int
	}{
		{"package main\nfunc f() { fmt.Println(1) }\n", 1},
		{"package main\n// fmt.Println(1)\nfunc f() { _ = \"fmt.Println(1)\" }\n", 0},
	} {
		result := engine.Evaluate(context.Background(), "a.go", []byte(test.content))
		count := 0
		for _, finding := range result.Findings {
			if finding.Kind == "error" {
				t.Fatalf("project rule error: %+v", finding)
			}
			if finding.Rule == "no-debug-output" {
				count++
			}
		}
		if count != test.want {
			t.Fatalf("project debug rule: %+v", result.Findings)
		}
	}
}

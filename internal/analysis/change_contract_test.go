package analysis

import (
	"context"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestChangeReceivesExactlyTwoArguments(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "arity", Type: "javascript", Code: `function(path, change) {
		if (arguments.length !== 2 || typeof change !== "object" || change.content !== "saved") {
			throw Error("incorrect rule contract");
		}
		return false;
	}`})
	result := engine.Evaluate(context.Background(), "a.txt", []byte("saved"))
	if len(result.Findings) != 0 {
		t.Fatalf("two-argument contract failed: %+v", result.Findings)
	}
}

func TestChangeCommentsAndMultilineLocations(t *testing.T) {
	for path, content := range map[string]string{
		"a.go":  "package p\n// comment\nfunc f() {}\n",
		"a.js":  "// comment\nconst x = 1;",
		"a.ts":  "// comment\nconst x: number = 1;",
		"a.tsx": "// comment\nconst x = <div/>;",
		"a.py":  "# comment\nx = 1\n",
		"a.rs":  "// comment\nfn f() {}\n",
	} {
		engine := testEngine(t, config.Rule{ID: "comments", Type: "javascript", Code: `(path, change) => {
			const comments = change.comments();
			if (comments.length !== 1 || !comments[0].text.includes("comment")) throw Error("comments");
			return [];
		}`})
		result := engine.Evaluate(context.Background(), path, []byte(content))
		if len(result.Findings) != 0 {
			t.Fatalf("%s comment adapter: %+v", path, result.Findings)
		}
	}
	engine := testEngine(t, config.Rule{ID: "multiline", Type: "javascript", Code: `(path, change) => {
		const comment = change.comments()[0];
		if (comment.location.end.line !== 2 || comment.location.end.column !== 6) throw Error(JSON.stringify(comment.location));
		return [comment.finding("comment")];
	}`})
	result := engine.Evaluate(context.Background(), "a.js", []byte("/* é😀\r\nxx */\n"))
	if len(result.Findings) != 1 || result.Findings[0].Kind != "violation" {
		t.Fatalf("multiline source location: %+v", result.Findings)
	}
}

func TestChangeNamedImportsAndArguments(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "imports", Type: "javascript", Code: `(path, change) => {
		const entry = change.imports()[0];
		if (entry.names.length !== 2 || entry.names[0].name !== "read" ||
		    entry.names[0].alias !== "load" || entry.names[1].name !== "write") throw Error(JSON.stringify(entry));
		const call = change.calls()[0];
		if (call.target.spelling !== "load" || call.arguments.length !== 2) throw Error("arguments");
		return [];
	}`})
	result := engine.Evaluate(context.Background(), "a.ts",
		[]byte("import { read as load, write } from './io'; load(1, /* note */ 2);"))
	if len(result.Findings) != 0 {
		t.Fatalf("named imports/call arguments: %+v", result.Findings)
	}
}

func TestChangeNestedFunctionNesting(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "nested", Type: "javascript", Code: `(path, change) => {
		const functions = change.functions();
		if (functions.length !== 2 || functions[0].metrics.nestingDepth !== 1 ||
		    functions[1].metrics.nestingDepth !== 2) throw Error("nested metrics");
		return [];
	}`})
	result := engine.Evaluate(context.Background(), "a.js",
		[]byte("function outer() { if (true) { const inner = () => { if (true) { if (true) {} } }; } }"))
	if len(result.Findings) != 0 {
		t.Fatalf("nested functions must have independent nesting: %+v", result.Findings)
	}
}

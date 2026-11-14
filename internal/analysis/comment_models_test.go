package analysis

import (
	"context"
	"fmt"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestChangeCommentClassification(t *testing.T) {
	for _, test := range []struct {
		name, path, content      string
		documentation, directive bool
	}{
		{"go ordinary", "a.go", "package p\n// ordinary\nfunc private() {}\n", false, false},
		{"go function", "a.go", "package p\n// Public docs.\nfunc Public() {}\n", true, false},
		{"go method", "a.go", "package p\ntype T struct{}\n// Public docs.\nfunc (t T) Public() {}\n", true, false},
		{"go type", "a.go", "package p\n// Public docs.\ntype Public struct{}\n", true, false},
		{"go grouped constant", "a.go", "package p\n// Public docs.\nconst (\nPublic = 1\n)\n", true, false},
		{"go variable", "a.go", "package p\n// Public docs.\nvar Public = 1\n", true, false},
		{"go package", "a.go", "// Package p docs.\npackage p\n", true, false},
		{"go cgo", "a.go", "package p\n/* C preamble */\nimport \"C\"\n", true, false},
		{"go detached", "a.go", "package p\n// Public docs.\n\nfunc Public() {}\n", false, false},
		{"go build", "a.go", "//go:build linux\n\npackage p\n", false, true},
		{"go legacy build", "a.go", "// +build linux\n\npackage p\n", false, true},
		{"go line", "a.go", "package p\n//line source.go:1\nvar x = 1\n", false, true},
		{"go export", "a.go", "package p\n//export private\nfunc private() {}\n", false, true},
		{"go nolint", "a.go", "package p\n//nolint:unused\nfunc private() {}\n", false, true},
		{"go both", "a.go", "package p\n//go:noinline\nfunc Public() {}\n", true, true},
		{"js ordinary", "a.js", "// ordinary\nconst x = 1;", false, false},
		{"js jsdoc", "a.js", "/** Docs. */\nfunction f() {}", true, false},
		{"js eslint", "a.js", "// eslint-disable-next-line\nconst x = 1;", false, true},
		{"js prettier", "a.js", "// prettier-ignore\nconst x = 1;", false, true},
		{"js coverage", "a.js", "/* istanbul ignore next */\nfunction f() {}", false, true},
		{"js purity", "a.js", "/*#__PURE__*/ f();", false, true},
		{"ts typing", "a.ts", "// @ts-expect-error\nconst x = 1;", false, true},
		{"tsx jsx", "a.tsx", "// @jsx h\nconst x = <div/>;", false, true},
		{"tsx ordinary", "a.tsx", "const x = <div>{/* ordinary */}</div>;", false, false},
		{"python ordinary", "a.py", "# ordinary\nx = 1\n", false, false},
		{"python shebang", "a.py", "#!/usr/bin/python\nx = 1\n", false, true},
		{"python encoding", "a.py", "# coding: utf-8\nx = 1\n", false, true},
		{"python typing", "a.py", "x = 1 # type: int\n", false, true},
		{"python lint", "a.py", "x = 1 # noqa\n", false, true},
		{"python formatting", "a.py", "# fmt: off\nx = 1\n", false, true},
		{"python coverage", "a.py", "x = 1 # pragma: no cover\n", false, true},
		{"rust ordinary", "a.rs", "// ordinary\nfn f() {}\n", false, false},
		{"rust outer line", "a.rs", "/// Docs.\npub fn f() {}\n", true, false},
		{"rust inner line", "a.rs", "//! Docs.\nfn f() {}\n", true, false},
		{"rust outer block", "a.rs", "/** Docs. */\npub fn f() {}\n", true, false},
		{"rust inner block", "a.rs", "/*! Docs. */\nfn f() {}\n", true, false},
		{"unknown directive", "a.js", "// custom-tool: ignore\nconst x = 1;", false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			code := fmt.Sprintf(`(path, change) => {
				const comments = change.comments();
				if (comments.length !== 1) throw Error("comment count: " + comments.length);
				const comment = comments[0];
				if (comment.isDocumentation !== %t || comment.isDirective !== %t) {
					throw Error(JSON.stringify(comment));
				}
				if (!Object.isFrozen(comments) || !Object.isFrozen(comment) ||
				    comments !== change.comments()) throw Error("mutable or uncached comments");
				if (change.ast.text(comment.rawNode) !== comment.text) throw Error("raw interop");
				return [comment.finding("classified")];
			}`, test.documentation, test.directive)
			engine := testEngine(t, config.Rule{ID: "classification", Type: "javascript", Code: code})
			result := engine.Evaluate(context.Background(), test.path, []byte(test.content))
			if len(result.Findings) != 1 || result.Findings[0].Kind != "violation" ||
				result.Findings[0].Message != "classified" {
				t.Fatalf("unexpected classification result: %+v", result.Findings)
			}
		})
	}
}

func TestChangeGoCommentDocumentationGroups(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "groups", Type: "javascript", Code: `(path, change) => {
		const comments = change.comments();
		if (comments.length !== 3 || !comments[0].isDocumentation ||
		    !comments[1].isDocumentation || comments[2].isDocumentation) throw Error("documentation group");
		return [];
	}`})
	result := engine.Evaluate(context.Background(), "a.go", []byte(
		"package p\n// Public docs.\n// More docs.\nfunc Public() {}\n// private docs.\nfunc private() {}\n"))
	if len(result.Findings) != 0 {
		t.Fatalf("incorrect group classification: %+v", result.Findings)
	}
}

func TestChangePythonDocstringsAreNotComments(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "docstrings", Type: "javascript", Code: `(path, change) => {
		if (change.comments().length || !change.functions()[0].documentation) throw Error("docstring classification");
		return [];
	}`})
	result := engine.Evaluate(context.Background(), "a.py", []byte("def f():\n    \"\"\"Docs.\"\"\"\n    pass\n"))
	if len(result.Findings) != 0 {
		t.Fatalf("docstrings must remain function documentation: %+v", result.Findings)
	}
}

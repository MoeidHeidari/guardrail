package analysis

import (
	"context"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestSyntaxSnapshotEmptyAndUnicodeBoundaries(t *testing.T) {
	for _, source := range []string{"", "const value = 'é😀';\r\n"} {
		syntax, nodes, err := parseSource(context.Background(), "a.js", []byte(source), true)
		if err != nil || !syntax.Parsed || len(nodes) == 0 {
			t.Fatalf("capture failed: syntax=%+v nodes=%d err=%v", syntax, len(nodes), err)
		}
		for _, node := range nodes {
			if node.Line < 1 || node.Column < 1 || node.End < node.Start {
				t.Fatalf("invalid boundary: %+v", node)
			}
		}
		if nodes[0].Start != 0 || nodes[0].Column != 1 {
			t.Fatalf("root boundary must start at 1:1: %+v", nodes[0])
		}
	}
}

func TestTextRulesDoNotDecodeSyntaxNodes(t *testing.T) {
	engine := testEngine(t, config.Rule{ID: "text", Type: "javascript",
		Code: `(path, content, syntax) => {
			if (syntax.language !== "text" || syntax.parsed) throw Error("wrong metadata");
			return false;
		}`})
	_, err := engine.javascript(context.Background(), engine.rules[0], "a.txt", "plain",
		syntaxSnapshot{Language: "text"}, []byte("not JSON"))
	if err != nil {
		t.Fatalf("text-only rules must not decode unused syntax: %v", err)
	}
	engine = testEngine(t, config.Rule{ID: "ast", Type: "javascript", Code: `(path, content, syntax) => syntax.nodes`})
	_, err = engine.javascript(context.Background(), engine.rules[0], "a.js", "",
		syntaxSnapshot{Language: "JavaScript", Parsed: true}, []byte("not JSON"))
	if err == nil {
		t.Fatal("accessing syntax nodes must decode and surface invalid snapshot errors")
	}
}

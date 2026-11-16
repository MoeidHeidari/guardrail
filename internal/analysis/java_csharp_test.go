package analysis

import (
	"context"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/config"
)

func TestJavaAndCSharpHelpers(t *testing.T) {
	tests := []struct {
		path, source, code string
	}{
		{
			"a.java",
			`import java.util.*;
import static java.util.Collections.emptyList;
class Example {
  /** Docs. */
  public void run(int a, int b) {
    if (a > 0) { System.out.println(a, b); }
    java.util.function.Function<Integer, Integer> f = x -> x + 1;
  }
  private Example() {}
}`,
			`(path, change) => {
			  change.require("functions", "imports", "comments", "calls");
			  if (change.language !== "Java" || !change.ast.parsed) throw Error("language");
			  const imports = change.imports();
			  if (imports.length !== 2 || imports[0].source !== "java.util.*" ||
			      imports[1].source !== "java.util.Collections.emptyList" || imports[0].alias !== null) {
			    throw Error("imports " + JSON.stringify(imports));
			  }
			  const functions = change.functions();
			  if (functions.length !== 3 || functions[0].name !== "run" ||
			      !functions[0].isPublic || !functions[0].documentation ||
			      functions[0].metrics.parameterCount !== 2 || functions[0].metrics.nestingDepth !== 1 ||
			      functions[1].metrics.parameterCount !== 1 || functions[2].isPublic) {
			    throw Error("functions " + JSON.stringify(functions));
			  }
			  const comments = change.comments(), calls = change.calls();
			  if (comments.length !== 1 || !comments[0].isDocumentation ||
			      calls.length !== 1 || calls[0].target.spelling !== "System.out.println" ||
			      calls[0].arguments.length !== 2) throw Error("comments/calls");
			  return [calls[0].finding("call")];
			}`,
		},
		{
			"a.cs",
			`global using System.Collections.Generic;
using Text = System.Text;
using static System.Console;
class Example {
  /// <summary>Docs.</summary>
  public void Run(int a, int b) {
    if (a > 0) { Console.WriteLine(a, b); }
    int Local(int value) => value;
    System.Func<int, int> f = x => x + 1;
  }
  private Example() {}
}`,
			`(path, change) => {
			  change.require("functions", "imports", "comments", "calls");
			  if (change.language !== "C#" || !change.ast.parsed) throw Error("language");
			  const imports = change.imports();
			  if (imports.length !== 3 || imports[0].source !== "System.Collections.Generic" ||
			      imports[1].source !== "System.Text" || imports[1].alias !== "Text" ||
			      imports[2].source !== "System.Console") throw Error("imports " + JSON.stringify(imports));
			  const functions = change.functions();
			  if (functions.length !== 4 || functions[0].name !== "Run" ||
			      !functions[0].isPublic || !functions[0].documentation ||
			      functions[0].metrics.parameterCount !== 2 || functions[0].metrics.nestingDepth !== 1 ||
			      functions[1].name !== "Local" || functions[1].metrics.parameterCount !== 1 ||
			      functions[2].metrics.parameterCount !== 1 || functions[3].isPublic) {
			    throw Error("functions " + JSON.stringify(functions));
			  }
			  const comments = change.comments(), calls = change.calls();
			  if (comments.length !== 1 || !comments[0].isDocumentation ||
			      calls.length !== 1 || calls[0].target.spelling !== "Console.WriteLine" ||
			      calls[0].arguments.length !== 2) throw Error("comments/calls");
			  return [calls[0].finding("call")];
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			engine := testEngine(t, config.Rule{ID: "helpers", Type: "javascript", Code: test.code})
			result := engine.Evaluate(context.Background(), test.path, []byte(test.source))
			if len(result.Findings) != 1 || result.Findings[0].Kind != "violation" ||
				result.Findings[0].Message != "call" {
				t.Fatalf("unexpected helper result: %+v", result.Findings)
			}
		})
	}
}

func TestJavaAndCSharpSyntaxFailures(t *testing.T) {
	for _, test := range []struct{ path, source string }{
		{"a.java", "class Example { void run( }"},
		{"a.cs", "class Example { void Run( }"},
	} {
		t.Run(test.path, func(t *testing.T) {
			engine := testEngine(t, config.Rule{ID: "ast", Type: "javascript",
				Code: `(path, change) => { change.ast.requireAST(); return false; }`})
			result := engine.Evaluate(context.Background(), test.path, []byte(test.source))
			if !result.Syntax.Parsed || len(result.Findings) != 2 ||
				result.Findings[0].Rule != "parser" || result.Findings[1].Rule != "ast" {
				t.Fatalf("syntax errors must be surfaced: %+v", result)
			}
		})
	}
}

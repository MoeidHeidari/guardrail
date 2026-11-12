package analysis

import (
	"context"
	"fmt"
	"path/filepath"

	ts "github.com/tree-sitter/go-tree-sitter"
	golang "github.com/tree-sitter/tree-sitter-go/bindings/go"
	javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

type Syntax struct {
	Language string `json:"language"`
	Parsed   bool   `json:"parsed"`
}

type SyntaxError struct {
	Language string
	Line     int
	Column   int
}

func (err *SyntaxError) Error() string {
	return fmt.Sprintf("%s syntax error at %d:%d (error or missing node)", err.Language, err.Line, err.Column)
}

func Parse(ctx context.Context, path string, content []byte) (Syntax, error) {
	if err := ctx.Err(); err != nil {
		return Syntax{}, err
	}
	var language *ts.Language
	syntax := Syntax{}
	switch filepath.Ext(path) {
	case ".go":
		syntax.Language, language = "Go", ts.NewLanguage(golang.Language())
	case ".js", ".jsx", ".mjs", ".cjs":
		syntax.Language, language = "JavaScript", ts.NewLanguage(javascript.Language())
	case ".ts":
		syntax.Language, language = "TypeScript", ts.NewLanguage(typescript.LanguageTypescript())
	case ".tsx":
		syntax.Language, language = "TSX", ts.NewLanguage(typescript.LanguageTSX())
	case ".py":
		syntax.Language, language = "Python", ts.NewLanguage(python.Language())
	case ".rs":
		syntax.Language, language = "Rust", ts.NewLanguage(rust.Language())
	default:
		syntax.Language = "text"
		return syntax, nil
	}
	parser := ts.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(language); err != nil {
		return syntax, err
	}
	tree := parser.ParseWithOptions(func(offset int, _ ts.Point) []byte {
		if offset >= len(content) {
			return nil
		}
		return content[offset:]
	}, nil, &ts.ParseOptions{ProgressCallback: func(ts.ParseState) bool { return ctx.Err() != nil }})
	if tree == nil {
		if err := ctx.Err(); err != nil {
			return syntax, err
		}
		return syntax, fmt.Errorf("%s parser did not produce a syntax tree", syntax.Language)
	}
	defer tree.Close()
	syntax.Parsed = true
	node := tree.RootNode()
	if node.HasError() {
		for !node.IsError() && !node.IsMissing() {
			var next *ts.Node
			for i := uint(0); i < node.ChildCount(); i++ {
				child := node.Child(i)
				if child.HasError() || child.IsMissing() {
					next = child
					break
				}
			}
			if next == nil {
				break
			}
			node = next
		}
		line, column := position(content, int(node.StartByte()))
		return syntax, &SyntaxError{Language: syntax.Language, Line: line, Column: column}
	}
	return syntax, nil
}

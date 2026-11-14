package analysis

import (
	"context"
	"fmt"
	"unicode/utf8"

	ts "github.com/tree-sitter/go-tree-sitter"
)

const syntaxNodeLimit = 50000

type syntaxNode struct {
	Type     string         `json:"type"`
	Start    int            `json:"start"`
	End      int            `json:"end"`
	Line     int            `json:"line"`
	Column   int            `json:"column"`
	EndLine  int            `json:"endLine"`
	Parent   int            `json:"parent"`
	Children []int          `json:"children"`
	Fields   map[string]int `json:"fields"`
}

type syntaxSnapshot struct {
	Language string       `json:"language"`
	Parsed   bool         `json:"parsed"`
	Error    string       `json:"error"`
	Nodes    []syntaxNode `json:"nodes"`
}

func captureSyntax(ctx context.Context, root *ts.Node, content []byte) ([]syntaxNode, error) {
	offsets := make([]int, len(content)+1)
	columns := make([]int, len(content)+1)
	columns[0] = 1
	units, column := 0, 1
	nextCheck := 0
	for offset := 0; offset < len(content); {
		if offset >= nextCheck {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			nextCheck = offset + 4096
		}
		char, size := utf8.DecodeRune(content[offset:])
		offsets[offset], columns[offset] = units, column
		units++
		if char > 0xffff {
			units++
		}
		column++
		if char == '\n' {
			column = 1
		}
		offset += size
		offsets[offset], columns[offset] = units, column
	}
	type pendingNode struct {
		node   *ts.Node
		parent int
		field  string
	}
	pending := []pendingNode{{node: root, parent: -1}}
	nodes := make([]syntaxNode, 0)
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(nodes) >= syntaxNodeLimit {
			return nil, fmt.Errorf("syntax API exceeded %d named nodes", syntaxNodeLimit)
		}
		item := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		node, id := item.node, len(nodes)
		nodes = append(nodes, syntaxNode{
			Type: node.Kind(), Start: offsets[node.StartByte()], End: offsets[node.EndByte()],
			Line: int(node.StartPosition().Row) + 1, Column: columns[node.StartByte()],
			EndLine: int(node.EndPosition().Row) + 1, Parent: item.parent,
			Children: []int{}, Fields: make(map[string]int),
		})
		if item.parent >= 0 {
			parent := &nodes[item.parent]
			parent.Children = append(parent.Children, id)
			if item.field != "" {
				parent.Fields[item.field] = id
			}
		}
		children := int(node.NamedChildCount())
		if len(nodes)+len(pending)+children > syntaxNodeLimit {
			return nil, fmt.Errorf("syntax API exceeded %d named nodes", syntaxNodeLimit)
		}
		for i := children - 1; i >= 0; i-- {
			pending = append(pending, pendingNode{
				node: node.NamedChild(uint(i)), parent: id,
				field: node.FieldNameForNamedChild(uint32(i)),
			})
		}
	}
	return nodes, nil
}

const syntaxAPI = `(encoded, content, language, parsed, error) => {
	const decode = JSON.parse, freeze = Object.freeze;
	let nodes, members;
	const load = () => {
		if (nodes) return nodes;
		nodes = decode(encoded).nodes || [];
		for (const node of nodes) {
			freeze(node.children);
			freeze(node.fields);
			freeze(node);
		}
		freeze(nodes);
		members = new Set(nodes);
		return nodes;
	};
	const at = id => {
		const nodes = load();
		return Number.isInteger(id) && id >= 0 && id < nodes.length ? nodes[id] : null;
	};
	const check = node => {
		load();
		if (!members.has(node)) throw Error("node does not belong to this syntax tree");
		return node;
	};
	return Object.freeze({
		version: 1, language, parsed, error,
		get nodes() { return load(); },
		requireAST: () => {
			if (error) throw Error(error);
			if (!parsed) throw Error("no grammar available for this file");
		},
		find: (...types) => load().filter(node => types.includes(node.type)),
		text: node => content.slice(check(node).start, node.end),
		field: (node, name) => at(check(node).fields[name]),
		children: node => check(node).children.map(at),
		parent: node => at(check(node).parent),
	});
}`

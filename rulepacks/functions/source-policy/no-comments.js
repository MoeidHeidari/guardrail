(path, content, syntax) => {
  syntax.requireAST();
  const comments = syntax.nodes.filter(node => ["comment", "line_comment", "block_comment"].includes(node.type));
  const documentation = new Set();
  if (syntax.language === "Go") {
    for (const declaration of syntax.nodes) {
      const name = syntax.field(declaration, "name");
      const publicName = name && Array.from(syntax.text(name))[0];
      const exported = publicName && publicName.toUpperCase() === publicName && publicName.toLowerCase() !== publicName;
      const publicGroup = ["type_declaration", "const_declaration", "var_declaration"].includes(declaration.type) &&
        syntax.children(declaration).some(child => {
          const name = syntax.field(child, "name");
          const first = name && Array.from(syntax.text(name))[0];
          return first && first.toUpperCase() === first && first.toLowerCase() !== first;
        });
      const cgo = declaration.type === "import_declaration" && syntax.text(declaration).includes('"C"');
      if (!exported && !["type_declaration", "const_declaration", "var_declaration",
        "package_clause", "import_declaration"].includes(declaration.type)) continue;
      const container = syntax.parent(declaration);
      if (!container) continue;
      const siblings = syntax.children(container);
      let index = siblings.indexOf(declaration) - 1;
      let next = declaration;
      while (index >= 0 && siblings[index].type === "comment") {
        const previous = siblings[index--];
        if (next.line - previous.endLine > 1 || !/^\s*$/.test(content.slice(previous.end, next.start))) break;
        if (exported || publicGroup || cgo ||
            (declaration.type === "package_clause" && /^\/\/\s*Package\s/.test(syntax.text(previous)))) {
          documentation.add(previous);
        }
        next = previous;
      }
    }
  }
  return comments.filter(node => {
    const text = syntax.text(node);
    if (syntax.language === "Go") {
      if (/^\/\/(?:go:|line |export |nolint(?::|\s|$))/.test(text) ||
          /^\/\/\s*\+build\s/.test(text)) return false;
      return !documentation.has(node);
    }
    if (["JavaScript", "TypeScript", "TSX"].includes(syntax.language) &&
        (/^\/\*\*/.test(text) || /^\/\/\s*@(?:ts-|jsx|flow)/.test(text) ||
          /^\/[/*]\s*(?:eslint|prettier|istanbul|c8|@__PURE__|#__PURE__)/.test(text))) return false;
    if (syntax.language === "Python" &&
        /^#(?:!|.*coding[:=]|\s*(?:type:|noqa\b|pragma:|pylint:|ruff:|fmt:))/.test(text)) return false;
    if (syntax.language === "Rust" &&
        (/^\/\/[!/]/.test(text) || /^\/\*[*!]/.test(text))) return false;
    return true;
  }).map(node => ({
    line: node.line, column: node.column,
    message: "Remove ordinary comments; keep only documentation and recognized compiler/tool directives.",
  }));
}

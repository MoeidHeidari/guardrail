(ast, content, model) => {
  const documentation = new Set();
  if (ast.language === "Go") {
    const exportedName = node => {
      const name = ast.field(node, "name");
      const first = name && Array.from(ast.text(name))[0];
      return !!first && first.toUpperCase() === first && first.toLowerCase() !== first;
    };
    for (const declaration of ast.nodes) {
      const exported = exportedName(declaration);
      const publicGroup = ["type_declaration", "const_declaration", "var_declaration"].includes(declaration.type) &&
        ast.children(declaration).some(exportedName);
      const cgo = declaration.type === "import_declaration" && ast.text(declaration).includes('"C"');
      if (!exported && !["type_declaration", "const_declaration", "var_declaration",
        "package_clause", "import_declaration"].includes(declaration.type)) continue;
      const container = ast.parent(declaration);
      if (!container) continue;
      const siblings = ast.children(container);
      let index = siblings.indexOf(declaration) - 1;
      let next = declaration;
      while (index >= 0 && siblings[index].type === "comment") {
        const previous = siblings[index--];
        if (next.line - previous.endLine > 1 || !/^\s*$/.test(content.slice(previous.end, next.start))) break;
        if (exported || publicGroup || cgo ||
            (declaration.type === "package_clause" && /^\/\/\s*Package\s/.test(ast.text(previous)))) {
          documentation.add(previous);
        }
        next = previous;
      }
    }
  }
  return ast.find("comment", "line_comment", "block_comment").map(node => {
    const text = ast.text(node);
    let isDocumentation = false, isDirective = false;
    if (ast.language === "Go") {
      isDocumentation = documentation.has(node);
      isDirective = /^\/\/(?:go:|line |export |nolint(?::|\s|$))/.test(text) ||
        /^\/\/\s*\+build\s/.test(text);
    } else if (["JavaScript", "TypeScript", "TSX"].includes(ast.language)) {
      isDocumentation = text.startsWith("/**");
      isDirective = /^\/\/\s*@(?:ts-|jsx|flow)/.test(text) ||
        /^\/[/*]\s*(?:eslint|prettier|istanbul|c8|@__PURE__|#__PURE__)/.test(text);
    } else if (ast.language === "Python") {
      isDirective = /^#(?:!|.*coding[:=]|\s*(?:type:|noqa\b|pragma:|pylint:|ruff:|fmt:))/.test(text);
    } else if (ast.language === "Rust") {
      isDocumentation = /^\/\/[!/]|^\/\*[*!]/.test(text);
    }
    return model(node, { isDocumentation, isDirective });
  });
}

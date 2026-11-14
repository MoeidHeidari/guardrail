(path, content, syntax) => {
  syntax.requireAST();
  const findings = [];
  const functions = syntax.find("function_declaration", "method_declaration",
    "function_definition", "function_item", "generator_function_declaration");
  for (const node of functions) {
    const name = syntax.field(node, "name");
    if (!name) continue;
    const parent = syntax.parent(node);
    let anchor = node;
    let isPublic = false;
    if (syntax.language === "Go") {
      const first = Array.from(syntax.text(name))[0];
      isPublic = first.toUpperCase() === first && first.toLowerCase() !== first;
    } else if (syntax.language === "Python") {
      anchor = parent && parent.type === "decorated_definition" ? parent : node;
      const scope = syntax.parent(anchor);
      isPublic = scope && scope.type === "module" && !syntax.text(name).startsWith("_");
      if (isPublic) {
        const body = syntax.field(node, "body");
        const first = body && syntax.children(body).find(child => child.type !== "comment");
        if (first && first.type === "expression_statement" &&
            syntax.children(first).some(child => child.type === "string" &&
              /^[rRuU]*['"]/.test(syntax.text(child)))) continue;
      }
    } else if (syntax.language === "Rust") {
      isPublic = syntax.children(node).some(child =>
        child.type === "visibility_modifier" && syntax.text(child) === "pub");
    } else {
      isPublic = parent && parent.type === "export_statement";
      if (isPublic) anchor = parent;
    }
    if (!isPublic) continue;
    if (syntax.language !== "Python") {
      const container = syntax.parent(anchor);
      const siblings = container ? syntax.children(container) : [];
      let index = siblings.indexOf(anchor) - 1;
      let next = anchor;
      let documented = false;
      while (index >= 0) {
        const previous = siblings[index--];
        if (!previous.type.includes("comment") ||
            !/^\s*$/.test(content.slice(previous.end, next.start)) ||
            next.line - previous.endLine > 1) break;
        const text = syntax.text(previous);
        if (syntax.language === "Go" && new RegExp(`^(?://|/\\*)\\s*${syntax.text(name)}(?:\\s|[.,:;!?]|$)`).test(text)) documented = true;
        if (syntax.language === "Rust" && (text.startsWith("///") || text.startsWith("/**"))) documented = true;
        if (["JavaScript", "TypeScript", "TSX"].includes(syntax.language) && text.startsWith("/**")) documented = true;
        next = previous;
      }
      if (documented) continue;
    }
    findings.push({
      line: name.line, column: name.column,
      message: `Document the public function ${syntax.text(name)}.`,
    });
  }
  return findings;
}

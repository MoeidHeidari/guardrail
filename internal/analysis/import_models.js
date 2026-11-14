(ast, model) => {
  const result = [];
  const add = (node, source, details) => {
    const names = [];
    const pending = [node];
    while (pending.length) {
      const child = pending.pop();
      if (child.type === "import_specifier") {
        const name = ast.field(child, "name"), alias = ast.field(child, "alias");
        names.push(Object.freeze({ name: name ? ast.text(name) : null, alias: alias ? ast.text(alias) : null }));
      } else pending.push(...ast.children(child).slice().reverse());
    }
    result.push(model(node, {
      source: source ? ast.text(source).replace(/^["'`]|["'`]$/g, "") : null,
      sourceNode: source ? model(source) : null,
      names: ["JavaScript", "TypeScript", "TSX"].includes(ast.language) && node.type === "import_statement" ?
        Object.freeze(names) : null,
      ...details,
    }));
  };
  for (const node of ast.find("import_spec", "import_statement", "import_from_statement", "use_declaration", "export_statement")) {
    const source = ast.field(node, "path") || ast.field(node, "source") ||
      ast.field(node, "module_name") || ast.field(node, "argument");
    if (node.type === "export_statement" && !source) continue;
    const alias = ast.field(node, "alias") || ast.field(node, "name");
    if (source) {
      add(node, source, {
        isDynamic: false, form: node.type,
        alias: alias ? ast.text(alias) : null,
      });
    } else {
      for (const child of ast.children(node).filter(child =>
        child.type === "dotted_name" || child.type === "aliased_import")) {
        const name = ast.field(child, "name") || child;
        const alias = ast.field(child, "alias");
        add(node, name, { isDynamic: false, form: node.type, alias: alias ? ast.text(alias) : null });
      }
    }
  }
  if (["JavaScript", "TypeScript", "TSX"].includes(ast.language)) {
    for (const call of ast.find("call_expression")) {
      const target = ast.field(call, "function"), argumentsNode = ast.field(call, "arguments");
      if (!target || !["require", "import"].includes(ast.text(target))) continue;
      const first = argumentsNode && ast.children(argumentsNode).find(child => !child.type.includes("comment"));
      add(call, first && first.type === "string" ? first : null, {
        isDynamic: true, form: ast.text(target), alias: null,
      });
    }
  }
  return result;
}

(ast, content, model) => {
  const types = ["function_declaration", "method_declaration", "function_definition",
    "function_item", "function_expression", "arrow_function", "method_definition",
    "generator_function_declaration", "generator_function", "closure_expression", "func_literal", "lambda"];
  const controls = ["if_statement", "for_statement", "while_statement", "do_statement",
    "switch_statement", "select_statement", "try_statement", "with_statement",
    "if_expression", "for_expression", "while_expression", "loop_expression", "match_expression"];
  const exportedName = name => {
    const first = Array.from(name)[0];
    return !!first && first.toUpperCase() === first && first.toLowerCase() !== first;
  };
  return ast.find(...types).map(node => {
    const nameNode = ast.field(node, "name");
    const name = nameNode ? ast.text(nameNode) : null;
    const body = ast.field(node, "body");
    const parametersNode = ast.field(node, "parameters");
    const singleParameter = ast.field(node, "parameter");
    const parameters = [];
    for (const parameter of parametersNode ? ast.children(parametersNode) : singleParameter ? [singleParameter] : []) {
      if (parameter.type.includes("comment") || ["positional_separator", "keyword_separator"].includes(parameter.type)) continue;
      const names = ast.language === "Go" ? ast.children(parameter).filter(child =>
        child.type === "identifier" && child !== ast.field(parameter, "type")) : [];
      if (names.length) {
        for (const identifier of names) parameters.push(model(identifier, { declaration: parameter }));
      } else parameters.push(model(parameter));
    }
    let nesting = 0;
    const pending = body ? [{ node: body, depth: 0 }] : [];
    while (pending.length) {
      const current = pending.pop();
      if (current.node !== body && types.includes(current.node.type)) continue;
      const depth = current.depth + (controls.includes(current.node.type) ? 1 : 0);
      nesting = Math.max(nesting, depth);
      for (const child of ast.children(current.node)) pending.push({ node: child, depth });
    }
    const parent = ast.parent(node);
    let anchor = node, isPublic = null, documentation = null;
    if (ast.language === "Go") isPublic = name ? exportedName(name) : false;
    else if (ast.language === "Python") {
      anchor = parent && parent.type === "decorated_definition" ? parent : node;
      const scope = ast.parent(anchor);
      isPublic = !!(name && scope && scope.type === "module" && !name.startsWith("_"));
      const first = body && ast.children(body).find(child => child.type !== "comment");
      const string = first && first.type === "expression_statement" &&
        ast.children(first).find(child => child.type === "string" && /^[rRuU]*['"]/.test(ast.text(child)));
      if (string) documentation = model(string);
    } else if (ast.language === "Rust") {
      isPublic = ast.children(node).some(child => child.type === "visibility_modifier" && ast.text(child) === "pub");
    } else {
      isPublic = !!(parent && parent.type === "export_statement");
      if (isPublic) anchor = parent;
    }
    if (ast.language !== "Python") {
      const container = ast.parent(anchor);
      const siblings = container ? ast.children(container) : [];
      let index = siblings.indexOf(anchor) - 1, next = anchor;
      while (index >= 0) {
        const previous = siblings[index--];
        if (!previous.type.includes("comment") || next.line - previous.endLine > 1 ||
            !/^\s*$/.test(content.slice(previous.end, next.start))) break;
        const text = ast.text(previous);
        if ((ast.language === "Go" && name && new RegExp(`^(?://|/\\*)\\s*${name}(?:\\s|[.,:;!?]|$)`).test(text)) ||
            (ast.language === "Rust" && /^\/\/\/|^\/\*\*/.test(text)) ||
            (["JavaScript", "TypeScript", "TSX"].includes(ast.language) && text.startsWith("/**"))) {
          documentation = model(previous);
        }
        next = previous;
      }
    }
    return model(node, {
      name, nameNode: nameNode ? model(nameNode) : null,
      body: body ? model(body) : null,
      parameters: Object.freeze(parameters),
      isPublic, documentation,
      metrics: Object.freeze({
        lineCount: node.endLine - node.line + 1,
        parameterCount: parameters.length,
        nestingDepth: nesting,
      }),
    });
  });
}

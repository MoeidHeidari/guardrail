(path, content, syntax) => {
  syntax.requireAST();
  const limits = { lines: 60, parameters: 5, nesting: 3 };
  if (Object.values(limits).some(limit => !Number.isInteger(limit) || limit < 1)) {
    throw new Error("Function budgets must be positive integers");
  }
  const functionTypes = [
    "function_declaration", "method_declaration", "function_definition",
    "function_item", "function_expression", "arrow_function", "method_definition",
    "generator_function_declaration", "generator_function", "closure_expression", "func_literal", "lambda",
  ];
  const controls = [
    "if_statement", "for_statement", "while_statement", "do_statement",
    "switch_statement", "select_statement", "try_statement", "with_statement",
    "if_expression", "for_expression", "while_expression", "loop_expression", "match_expression",
  ];
  const findings = [];
  for (const node of syntax.find(...functionTypes)) {
    const body = syntax.field(node, "body");
    if (!body) continue;
    const parameters = syntax.field(node, "parameters");
    let count = syntax.field(node, "parameter") ? 1 : 0;
    if (parameters) {
      for (const parameter of syntax.children(parameters)) {
        if (parameter.type === "comment" || parameter.type === "positional_separator" ||
            parameter.type === "keyword_separator") continue;
        if (syntax.language === "Go") {
          const type = syntax.field(parameter, "type");
          count += Math.max(1, syntax.children(parameter).filter(child =>
            child !== type && child.type === "identifier").length);
        } else {
          count++;
        }
      }
    }
    let depth = 0;
    const pending = [{ node: body, depth: 0 }];
    while (pending.length) {
      const current = pending.pop();
      if (current.node !== body && functionTypes.includes(current.node.type)) continue;
      const next = current.depth + (controls.includes(current.node.type) ? 1 : 0);
      depth = Math.max(depth, next);
      for (const child of syntax.children(current.node)) pending.push({ node: child, depth: next });
    }
    const lines = node.endLine - node.line + 1;
    const exceeded = [];
    if (lines > limits.lines) exceeded.push(`${lines} lines (maximum ${limits.lines})`);
    if (count > limits.parameters) exceeded.push(`${count} parameters (maximum ${limits.parameters})`);
    if (depth > limits.nesting) exceeded.push(`nesting ${depth} (maximum ${limits.nesting})`);
    if (exceeded.length) findings.push({
      line: node.line, column: node.column,
      message: `Split this function: ${exceeded.join(", ")}.`,
    });
  }
  return findings;
}

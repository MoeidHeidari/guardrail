(path, change) => {
  const ast = change.ast;
  ast.requireAST();
  const prohibited = {
    Go: ["fmt.Println", "fmt.Printf", "fmt.Print", "print", "println"],
    JavaScript: ["console.log", "console.debug", "console.trace"],
    TypeScript: ["console.log", "console.debug", "console.trace"],
    TSX: ["console.log", "console.debug", "console.trace"],
    Python: ["print", "breakpoint", "pdb.set_trace"],
    Rust: ["println", "print", "eprintln", "eprint", "dbg"],
  };
  const spelling = node => {
    const object = ast.field(node, "object") || ast.field(node, "operand");
    const member = ast.field(node, "property") || ast.field(node, "attribute") || ast.field(node, "field");
    return object && member ? `${spelling(object)}.${ast.text(member)}` : ast.text(node).replace(/\s/g, "");
  };
  return ast.find("call_expression", "call", "macro_invocation", "debugger_statement").filter(node => {
    const target = ast.field(node, "function") || ast.field(node, "macro");
    return node.type === "debugger_statement" || (target && prohibited[change.language].includes(spelling(target)));
  }).map(node => {
    const target = ast.field(node, "function") || ast.field(node, "macro");
    return change.finding(
      `Replace ${target ? spelling(target) : "debugger"} with the project's explicit logging/output abstraction.`,
      change.location(node).start);
  });
}

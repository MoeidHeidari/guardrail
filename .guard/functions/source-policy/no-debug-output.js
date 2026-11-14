(path, content, syntax) => {
  syntax.requireAST();
  const prohibited = {
    Go: ["fmt.Println", "fmt.Printf", "fmt.Print", "print", "println"],
    JavaScript: ["console.log", "console.debug", "console.trace"],
    TypeScript: ["console.log", "console.debug", "console.trace"],
    TSX: ["console.log", "console.debug", "console.trace"],
    Python: ["print", "breakpoint", "pdb.set_trace"],
    Rust: ["println", "print", "eprintln", "eprint", "dbg"],
  };
  const names = prohibited[syntax.language];
  if (!names) throw new Error(`Unsupported source-policy language: ${syntax.language}`);
  const spelling = node => {
    const object = syntax.field(node, "object") || syntax.field(node, "operand");
    const member = syntax.field(node, "property") || syntax.field(node, "attribute") ||
      syntax.field(node, "field");
    return object && member ? `${spelling(object)}.${syntax.text(member)}` : syntax.text(node).replace(/\s/g, "");
  };
  const findings = [];
  for (const node of syntax.find("call_expression", "call", "macro_invocation", "debugger_statement")) {
    const target = syntax.field(node, "function") || syntax.field(node, "macro");
    const name = target ? spelling(target) : "";
    if (node.type === "debugger_statement" || names.includes(name)) {
      findings.push({
        line: node.line, column: node.column,
        message: `Replace ${name || "debugger"} with the project's explicit logging/output abstraction.`,
      });
    }
  }
  return findings;
}

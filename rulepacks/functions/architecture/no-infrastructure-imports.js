(path, content, syntax) => {
  syntax.requireAST();
  const domainPath = /(^|\/)(domain|core)\//;
  const forbidden = /(^|[/.:\s])(infrastructure|adapters|transport)([/.:\s]|$)/;
  if (!domainPath.test(path.replace(/\\/g, "/"))) return [];
  const imports = syntax.find("import_spec", "import_statement", "import_from_statement",
    "use_declaration", "export_statement");
  for (const call of syntax.find("call_expression")) {
    const target = syntax.field(call, "function");
    const argumentsNode = syntax.field(call, "arguments");
    if (target && ["require", "import"].includes(syntax.text(target)) && argumentsNode) {
      const source = syntax.children(argumentsNode)[0];
      if (source && source.type === "string") imports.push(source);
    }
  }
  const findings = [];
  for (const node of imports) {
    const source = syntax.field(node, "path") || syntax.field(node, "source") ||
      syntax.field(node, "module_name") || syntax.field(node, "argument");
    let names = node.type === "string" ? [node] : source ? [source] : syntax.children(node).filter(child =>
      child.type === "dotted_name" || child.type === "aliased_import");
    names = names.map(name => name.type === "aliased_import" ?
      syntax.field(name, "name") || name : name);
    for (const name of names) {
      let text = syntax.text(name).replace(/["'`]/g, "");
      if (syntax.language === "Rust") {
        const parts = [];
        const pending = [name];
        while (pending.length) {
          const child = pending.pop();
          if (["identifier", "crate", "self", "super"].includes(child.type)) parts.push(syntax.text(child));
          else if (!child.type.includes("comment")) pending.push(...syntax.children(child));
        }
        text = parts.join("::");
      }
      if (forbidden.test(text)) {
        findings.push({
          line: name.line, column: name.column,
          message: "Domain/core code must not import infrastructure, adapters, or transport.",
        });
      }
    }
  }
  return findings;
}

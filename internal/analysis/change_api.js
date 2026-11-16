(functionModels, importModels, commentModels) => (content, ast) => {
  const freeze = Object.freeze;
  const supported = ["Go", "JavaScript", "TypeScript", "TSX", "Python", "Rust", "Java", "C#"];
  const capabilities = freeze(supported.includes(ast.language) ?
    ["functions", "imports", "comments", "calls"] : []);
  let lines;
  const sourceLines = () => {
    if (!lines) {
      lines = content.length ? content.split(/\r?\n/) : [];
      if (content.endsWith("\n")) lines.pop();
      freeze(lines);
    }
    return lines;
  };
  const require = (...names) => {
    for (const name of names) {
      if (!capabilities.includes(name)) throw Error(`Unsupported capability ${name} for ${ast.language}`);
    }
    ast.requireAST();
  };
  const location = node => {
    const text = ast.text(node);
    const lastNewline = text.lastIndexOf("\n");
    const endColumn = lastNewline < 0 ? node.column + Array.from(text).length :
      Array.from(text.slice(lastNewline + 1)).length + 1;
    return freeze({
      start: freeze({ line: node.line, column: node.column, offset: node.start }),
      end: freeze({ line: node.endLine, column: endColumn, offset: node.end }),
    });
  };
  const finding = (message, position) => {
    if (typeof message !== "string" || !message.length) throw Error("Finding message must be nonempty");
    if (!position || !Number.isInteger(position.line) || !Number.isInteger(position.column) ||
        position.line < 1 || position.column < 1) throw Error("Finding requires a 1-based source position");
    return { line: position.line, column: position.column, message };
  };
  const model = (node, details = {}) => {
    const position = location(node);
    return freeze({
      kind: node.type, text: ast.text(node), rawNode: node, location: position,
      finding: message => finding(message, position.start), ...details,
    });
  };
  const spelling = node => {
    const object = ast.field(node, "object") || ast.field(node, "operand") ||
      (node.type === "member_access_expression" && ast.field(node, "expression"));
    const member = ast.field(node, "property") || ast.field(node, "attribute") || ast.field(node, "field") ||
      (node.type === "member_access_expression" && ast.field(node, "name"));
    return object && member ? `${spelling(object)}.${ast.text(member)}` : ast.text(node).replace(/\s/g, "");
  };
  const cache = new Map();
  const collection = (name, build) => {
    require(name);
    if (!cache.has(name)) cache.set(name, freeze(build()));
    return cache.get(name);
  };
  return freeze({
    version: 1, content, language: ast.language, ast, capabilities,
    get lines() { return sourceLines(); },
    get lineCount() { return sourceLines().length; },
    require, location, finding,
    functions: () => collection("functions", () => functionModels(ast, content, model)),
    imports: () => collection("imports", () => importModels(ast, model)),
    comments: () => collection("comments", () => commentModels(ast, content, model)),
    calls: () => collection("calls", () => ast.find("call_expression", "call", "macro_invocation", "debugger_statement",
      "method_invocation", "invocation_expression")
      .map(node => {
        const target = ast.field(node, "function") || ast.field(node, "macro") ||
          (node.type === "method_invocation" && ast.field(node, "name"));
        const argumentsNode = ast.field(node, "arguments");
        const object = node.type === "method_invocation" && ast.field(node, "object");
        const targetSpelling = target && (object ? `${spelling(object)}.${ast.text(target)}` : spelling(target));
        return model(node, {
          target: target ? model(target, { spelling: targetSpelling }) : null,
          arguments: argumentsNode ? freeze(ast.children(argumentsNode).filter(child =>
            !child.type.includes("comment")).map(child => model(child))) : null,
          isMacro: node.type === "macro_invocation",
          isDebugger: node.type === "debugger_statement",
        });
      })),
  });
}

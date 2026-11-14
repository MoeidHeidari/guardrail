(path, change) => {
  change.require("imports");
  const domainPath = /(^|\/)(domain|core)\//;
  const forbidden = /(^|[/.:\s])(infrastructure|adapters|transport)([/.:\s]|$)/;
  if (!domainPath.test(path.replace(/\\/g, "/"))) return [];
  return change.imports().filter(entry => {
    if (!entry.sourceNode) return false;
    let source = entry.source;
    if (change.language === "Rust") {
      const parts = [];
      const pending = [entry.sourceNode.rawNode];
      while (pending.length) {
        const node = pending.pop();
        if (["identifier", "crate", "self", "super"].includes(node.type)) parts.push(change.ast.text(node));
        else if (!node.type.includes("comment")) pending.push(...change.ast.children(node));
      }
      source = parts.join("::");
    }
    return forbidden.test(source);
  }).map(entry => entry.sourceNode.finding(
    "Domain/core code must not import infrastructure, adapters, or transport."));
}

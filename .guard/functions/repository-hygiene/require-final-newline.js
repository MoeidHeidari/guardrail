(path, change) => {
  const content = change.content;
  if (content.length === 0 || content.endsWith("\n")) {
    return [];
  }
  const lines = content.split(/\r?\n/);
  return [{
    line: lines.length,
    column: Array.from(lines[lines.length - 1]).length + 1,
    message: "End this file with a newline.",
  }];
}

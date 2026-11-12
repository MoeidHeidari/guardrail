(path, content) => {
  const limit = 300;
  const lines = content.split(/\r?\n/);
  const count = lines.length - (content.endsWith("\n") ? 1 : 0);
  if (count <= limit) return [];
  return [{
    line: limit + 1,
    column: 1,
    message: `${path} has ${count} lines; the example limit is ${limit}.`
  }];
}

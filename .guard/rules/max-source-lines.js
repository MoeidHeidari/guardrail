(path, content) => {
  const limit = 100;
  const lines = content.split(/\r?\n/);
  const count = lines.length - (content.endsWith("\n") ? 1 : 0);
  if (count <= limit) {
    return false;
  }
  return [{
    line: limit + 1,
    column: 1,
    message: `${path} has ${count} lines; the repository limit is ${limit}.`,
  }];
}

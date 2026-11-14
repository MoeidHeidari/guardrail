(path, change) => {
  const limit = 300;
  const count = change.lineCount;
  if (count <= limit) return [];
  return [{
    line: limit + 1,
    column: 1,
    message: `${path} has ${count} lines; the example limit is ${limit}.`
  }];
}

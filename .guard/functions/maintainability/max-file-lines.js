(path, content) => {
  const maximum = 250;
  if (!Number.isInteger(maximum) || maximum < 1) {
    throw new Error("maximum must be a positive integer");
  }
  if (content.length === 0) {
    return [];
  }
  const count = content.split(/\r?\n/).length - (content.endsWith("\n") ? 1 : 0);
  return count > maximum ? [{
    line: maximum + 1,
    column: 1,
    message: `This file has ${count} lines; the limit is ${maximum}.`,
  }] : [];
}

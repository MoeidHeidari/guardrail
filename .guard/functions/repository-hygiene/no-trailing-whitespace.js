(path, change) => {
  const content = change.content;
  const findings = [];
  const lines = content.split(/\r?\n/);
  for (let index = 0; index < lines.length; index++) {
    const match = /[ \t]+$/.exec(lines[index]);
    if (match) {
      findings.push({
        line: index + 1,
        column: Array.from(lines[index].slice(0, match.index)).length + 1,
        message: "Remove trailing spaces or tabs.",
      });
    }
  }
  return findings;
}

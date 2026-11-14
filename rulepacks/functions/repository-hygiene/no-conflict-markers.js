(path, change) => {
  const content = change.content;
  const findings = [];
  const lines = content.split(/\r?\n/);
  for (let index = 0; index < lines.length; index++) {
    if (/^(?:<{7}(?: .*)?|={7}|>{7}(?: .*)?|\|{7}(?: .*)?)$/.test(lines[index])) {
      findings.push({
        line: index + 1,
        column: 1,
        message: "Resolve this merge-conflict marker.",
      });
    }
  }
  return findings;
}

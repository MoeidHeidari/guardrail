(path, change) => {
  const content = change.content;
  if (!/(^|\/)package\.json$/.test(path)) return [];
  const manifest = JSON.parse(content);
  if (!manifest || Array.isArray(manifest) || typeof manifest !== "object") {
    throw new Error("package.json must contain an object");
  }
  const forbidden = ["left-pad", "request"];
  const findings = [];
  for (const section of ["dependencies", "devDependencies", "optionalDependencies", "peerDependencies"]) {
    const dependencies = manifest[section];
    if (dependencies === undefined) continue;
    if (!dependencies || Array.isArray(dependencies) || typeof dependencies !== "object") {
      throw new Error(`${section} must contain an object`);
    }
    for (const name of Object.keys(dependencies)) {
      const version = dependencies[name];
      if (typeof version !== "string") throw new Error(`${section}.${name} must be a string`);
      const alias = /^npm:((?:@[^/]+\/)?[^@]+)(?:@|$)/.exec(version);
      if (forbidden.includes(name) || (alias && forbidden.includes(alias[1]))) {
        findings.push({ line: 1, column: 1,
          message: `Replace unapproved dependency ${name} in ${section}.` });
      }
    }
  }
  return findings;
}

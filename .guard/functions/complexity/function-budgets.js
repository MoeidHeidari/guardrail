(path, change) => {
  const limits = { lines: 60, parameters: 5, nesting: 3 };
  if (Object.values(limits).some(limit => !Number.isInteger(limit) || limit < 1)) {
    throw new Error("Function budgets must be positive integers");
  }
  return change.functions().filter(fn => fn.body).flatMap(fn => {
    const exceeded = [];
    if (fn.metrics.lineCount > limits.lines) exceeded.push(`${fn.metrics.lineCount} lines (maximum ${limits.lines})`);
    if (fn.metrics.parameterCount > limits.parameters) exceeded.push(`${fn.metrics.parameterCount} parameters (maximum ${limits.parameters})`);
    if (fn.metrics.nestingDepth > limits.nesting) exceeded.push(`nesting ${fn.metrics.nestingDepth} (maximum ${limits.nesting})`);
    return exceeded.length ? [fn.finding(`Split this function: ${exceeded.join(", ")}.`)] : [];
  });
}

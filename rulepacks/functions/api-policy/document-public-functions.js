(path, change) => {
  return change.functions().filter(fn => fn.isPublic && fn.name && !fn.documentation)
    .map(fn => fn.nameNode.finding(`Document the public function ${fn.name}.`));
}

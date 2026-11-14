(path, change) => change.comments()
  .filter(comment => !comment.isDocumentation && !comment.isDirective)
  .map(comment => comment.finding(
    "Remove ordinary comments; keep only documentation and recognized compiler/tool directives."
  ))

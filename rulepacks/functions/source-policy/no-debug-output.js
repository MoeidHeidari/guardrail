(path, change) => {
  const prohibited = {
    Go: ["fmt.Println", "fmt.Printf", "fmt.Print", "print", "println"],
    JavaScript: ["console.log", "console.debug", "console.trace"],
    TypeScript: ["console.log", "console.debug", "console.trace"],
    TSX: ["console.log", "console.debug", "console.trace"],
    Python: ["print", "breakpoint", "pdb.set_trace"],
    Rust: ["println", "print", "eprintln", "eprint", "dbg"],
  };
  return change.calls().filter(call =>
    call.isDebugger || (call.target && prohibited[change.language].includes(call.target.spelling)))
    .map(call => call.finding(
      `Replace ${call.target ? call.target.spelling : "debugger"} with the project's explicit logging/output abstraction.`));
}

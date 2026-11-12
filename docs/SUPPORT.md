# Support

[Documentation index](README.md)

## Where to ask

Use [GitHub issues](https://github.com/MoeidHeidari/guardrail/issues) for bugs,
usage questions, and feature proposals. Search existing issues first.
Community support is best effort; there is no guaranteed response time.

- Vulnerability reports: follow the [security policy](SECURITY.md).
- Conduct concerns: use the private contact in the [code of conduct](CODE_OF_CONDUCT.md).
- Contributions: start with [Contributing](CONTRIBUTING.md).

## Before reporting

1. Read [Getting started](GETTING_STARTED.md) and the [watch guide](WATCH.md).
2. Check whether the behavior is a documented [limit](PERFORMANCE.md).
3. Reproduce using a small disposable context where possible.
4. Identify the commit/build you used and whether it still happens on current code.

Regex/JavaScript configuration and `guardrail check` are supported; consult
[Rules and checking](RULES.md) for errors and scope. Agent/editor feedback and
dependency-aware policies remain planned.

## Bug-report checklist

Include:

- A short description and exact steps to reproduce.
- The command you ran, with sensitive paths replaced.
- Expected behavior and actual behavior.
- Full error output or a panic stack trace when available.
- OS/version, terminal application/version, Go version, and commit hash.
- Approximate file/directory counts and whether the context contains large
  dependency or generated folders.
- For UI problems: terminal dimensions, light/dark theme, keyboard/mouse action,
  and a redacted screenshot or recording.
- For delayed updates: edit/save pattern, scrolling activity, and approximate delay.

Useful local commands:

```sh
go version
git rev-parse HEAD
./bin/guardrail watch --help
```

Do not attach private source code, credentials, or an entire repository when a
synthetic fixture can demonstrate the issue.

## Feature proposals

Describe the problem, who it affects, expected workflow, and tradeoffs.
For rule proposals, include the desired rule, target languages, violation and
non-violation examples, and expected file/location output.
For integrations, name the tool and how feedback should reach it.

See [Architecture and roadmap](ARCHITECTURE.md) before proposing a new subsystem.

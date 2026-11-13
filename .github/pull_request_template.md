## Summary

Describe the problem, intended outcome, and approach.

## Related issues

Link related issues. Use `Closes #123` only when this PR fully resolves the issue.

## Changes and compatibility

- Describe important behavior, CLI/configuration, or output changes.
- Note breaking changes, migration steps, and known limitations, or write "None".
- Keep planned follow-up work distinct from functionality implemented here.

## Validation

List the exact checks run and their results. Explain skipped or inapplicable checks.
See the [development guide](https://github.com/MoeidHeidari/guardrail/blob/main/docs/DEVELOPMENT.md) for commands.

```text
Command:
Result:
```

For TUI changes, include terminal/OS details, keyboard and mouse checks, resizing,
and light/dark theme checks. Attach redacted screenshots or recordings.
For performance changes, include the workload and before/after measurements.

## Checklist

- [ ] This PR is focused and uses a Conventional Commit title.
- [ ] New or changed behavior has tests, including regression coverage for fixes.
- [ ] Relevant tests, formatting, vet, build, and lint pass, or exceptions are explained.
- [ ] Every source and test file is at most 300 lines, including imports and blank lines.
- [ ] Directly related documentation is updated.
- [ ] Breaking changes are documented and marked with `!` or a `BREAKING CHANGE:` footer where applicable.
- [ ] No secrets, private fixtures, generated binaries, or temporary logs are included.

For vulnerabilities, use the [private reporting process](https://github.com/MoeidHeidari/guardrail/blob/main/docs/SECURITY.md)
instead of disclosing details in a public PR.

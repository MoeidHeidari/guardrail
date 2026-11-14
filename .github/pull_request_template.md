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

## New community rule

<!-- Complete this section for rule-pack contributions; otherwise write "Not applicable". -->

Follow the [rule-pack contribution guide](https://github.com/MoeidHeidari/guardrail/blob/main/rulepacks/README.md).

### Rule definition

- **Rule ID and category:**
- **Engineering standard and intended audience:**
- **Supported languages, file patterns, and exclusions:**
- **YAML rule-set paths and shared function path:**
- **Required Guardrail/syntax API version:**
- **Implementation style:** `change` helpers, `change.ast`, or a combination.
- **Default thresholds, exceptions, and customization:**
- **Known limitations and false positives:**

### Examples and expected findings

Include synthetic passing and failing examples. Specify exact finding counts,
rule IDs, messages, and 1-based line/column locations. For file-level findings,
state that the location is `0:0`. Include relevant boundary cases.

**Passing example:**

```text
Input:
Expected findings:
```

**Failing example:**

```text
Input:
Expected findings:
```

### Rule validation

Link executable regression tests and list exact commands with their results in
the Validation section above. Test each claimed language using the actual YAML
and function through Guardrail's engine, not only a standalone JavaScript runtime.
Cover relevant Unicode, CRLF, empty input, thresholds, syntax errors, exceptions,
and resource limits. Distinguish policy violations from execution errors.

### Rule contribution checklist

- [ ] YAML is in the appropriate language folder (or `general/` for text policies).
- [ ] Reusable JavaScript lives under `rulepacks/functions/<category>/`, without unnecessary per-language copies.
- [ ] Rule IDs do not collide with other categories within the same language set.
- [ ] YAML references match the documented `.guard/functions/` installation paths.
- [ ] Each supported language has passing/failing regression coverage.
- [ ] The category guide and catalog document installation, scope, customization, and limitations.
- [ ] The rule is deterministic and bounded; unavailable required syntax fails explicitly.
- [ ] Syntax-only checks are not described as type-aware or cross-file analysis.
- [ ] Authorship, licensing, and any required third-party attribution are documented.
- [ ] The contribution does not silently activate rules in this repository's `.guard/`.

For vulnerabilities, use the [private reporting process](https://github.com/MoeidHeidari/guardrail/blob/main/docs/SECURITY.md)
instead of disclosing details in a public PR.

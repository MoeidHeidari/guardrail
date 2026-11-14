# Contributing to Guardrail

[Documentation index](README.md)

Thank you for helping improve Guardrail. Documentation, bug reports, reproducible
tests, performance work, and focused features are all useful contributions.
Participation is subject to the [code of conduct](CODE_OF_CONDUCT.md).

## Before starting

- Search [existing issues](https://github.com/MoeidHeidari/guardrail/issues) and
  [pull requests](https://github.com/MoeidHeidari/guardrail/pulls) for related work.
- Open an issue before a substantial feature or architectural change so scope
  and approach can be agreed on.
- Read [Architecture and roadmap](ARCHITECTURE.md) to distinguish implemented
  watching and analysis from planned integrations.
- Send vulnerabilities privately according to the [security policy](SECURITY.md).

## Contribution workflow

1. Fork [MoeidHeidari/guardrail](https://github.com/MoeidHeidari/guardrail).
2. Clone your fork and configure the upstream remote:

   ```sh
   git clone https://github.com/YOUR_USERNAME/guardrail.git
   cd guardrail
   git remote add upstream https://github.com/MoeidHeidari/guardrail.git
   ```

3. Start a focused branch from current upstream `main`:

   ```sh
   git fetch upstream
   git switch -c feat/describe-your-change upstream/main
   ```

   Use a suitable prefix such as `feat/`, `fix/`, or `docs/`.

4. Make the change and add relevant tests and documentation.
5. Run the checks in [Development](DEVELOPMENT.md).
6. Commit, push to your fork, and open a PR targeting upstream `main`.

If your branch becomes outdated, update it against current upstream `main` before
final review. Avoid mixing unrelated cleanup with a feature or bug fix.

## Code style

### Community rules

Reusable engineering standards belong in [rulepacks/](../rulepacks/README.md),
not the active `.guard/` directory. Follow the pack contribution guide for
layout, rule contracts, licensing, fixtures, and validation evidence. Add new
packs to its catalog and document exact installation paths and limitations.
Complete the **New community rule** section in the
[PR template](../.github/pull_request_template.md), including passing/failing
examples and exact expected findings for every claimed language.
Use the documented syntax API for AST-backed rules; do not present type analysis,
remote installation, or a dedicated pack-test runner
as existing functionality.

### Project code

- Follow the repository's existing patterns and Go formatting.
- Keep functions and components focused and maintainable.
- Keep source and test files at or below 300 lines, including blank lines/imports.
  Split by responsibility rather than hiding logic in unnecessary abstractions.
- Preserve existing behavior unless a change is intentional and documented.
- Add regression tests for fixes and tests for new behavior.
- Update directly related documentation when commands, controls, or limits change.
- Include actionable errors instead of silent failures.

See [AGENTS.md](../AGENTS.md) and the [development guide](DEVELOPMENT.md) for details.

## Pull requests

Prefer one feature or fix per PR. Fill in the
[pull request template](../.github/pull_request_template.md), including exact
validation commands and results. A useful description includes:

- The problem and intended outcome.
- A linked issue, where applicable.
- Implementation notes and important behavior changes.
- Tests/checks run and their results.
- Terminal screenshots or recordings for visible UI changes, without private data.
- Benchmark context and before/after results for performance work.
- Any known limitations or follow-up work.

### Conventional titles and commits

Use [Conventional Commits](https://www.conventionalcommits.org/) for PR titles:

```text
type(optional-scope): concise description
```

Allowed types:
`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`,
and `revert`.

Examples:

```text
feat(watch): add context selection
fix(tui): preserve scrolling during update bursts
perf(watcher): reconcile only affected subtrees
docs: reorganize user and contributor guides
```

Use the same style for commit messages where possible. Clearly explain breaking
changes in the PR and use Conventional Commits breaking-change notation when
applicable. The Quality workflow validates Conventional PR titles for pull
requests targeting `main`.

## Review checklist

- [ ] The change is focused and its intent is clear.
- [ ] Added or changed behavior has appropriate tests.
- [ ] Source and test files respect the 300-line convention.
- [ ] Changed Go files are formatted; relevant checks pass.
- [ ] User-facing documentation reflects the actual behavior.
- [ ] No secrets, private source fixtures, generated binaries, or temporary logs
      are included.
- [ ] Planned capabilities are not presented as implemented features.

Maintainers may request revisions or a narrower scope before merging.

## Issues and questions

Use the [issue tracker](https://github.com/MoeidHeidari/guardrail/issues) for bugs
and feature proposals. Choose the bug report, feature request, or language support
form; for usage questions, start with the support guide.
Include reproduction steps, expected and actual
behavior, OS, terminal application, Go version, and relevant context size.
See [Support](SUPPORT.md) for a reporting checklist.

## Releases and changelog

The manually dispatched release workflow runs quality checks, plans a semantic
version from Conventional Commits, and publishes native binaries and release
notes. See [Releasing](RELEASING.md) for the exact process. Describe noteworthy
behavior changes in the PR and relevant docs; do not invent release entries or
versions.

## Contribution license

Contributions are made under the project's [Apache License 2.0](../LICENSE).
Only submit material you have the right to contribute. Review and test generated
code yourself; use of an AI tool does not replace contributor responsibility.

## Related contribution practices

This guide follows the workflow and review practices referenced in
[Crossview's contribution guide](https://github.com/crossplane-contrib/crossview/blob/main/docs/CONTRIBUTING.md):
fork-based PRs, focused changes, conventional PR titles, tests/linting,
documentation updates, and keeping branches current. Guardrail-specific commands,
coding limits, support channels, and automation status are documented here
instead of assuming Crossview's frontend/backend setup or release tooling.

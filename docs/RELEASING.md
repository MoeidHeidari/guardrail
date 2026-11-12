# Quality checks and releases

Pull requests to `main` run formatting, the 300-line Go file convention,
golangci-lint v2.6.2, vet, race tests with coverage, and semantic-versioning tests,
without building release binaries. Go is selected from `go.mod`. The pinned linter is
built with that Go toolchain so its binary can analyze the project's Go version.

Quality tests currently run on Linux and macOS only. Windows quality testing is
temporarily excluded because directory rename/recreation watcher tests still
fail there. Windows release builds and version smoke tests remain enabled;
a successful build does not certify Windows watcher behavior.

## Merge convention

Use squash merges with a Conventional Commit PR title:

| Commit | Release |
| --- | --- |
| `fix:`, `perf:`, `revert:` | Patch |
| `feat:` | Minor |
| `type!:` or a `BREAKING CHANGE:` / `BREAKING-CHANGE:` footer | Major after 1.0 |
| `docs:`, `test:`, `ci:`, `chore:`, nonbreaking `refactor:`, `build:`, `style:` | None |

The initial release is `v0.1.0`. Before 1.0, breaking changes increment the
minor version. After 1.0 they increment major. The highest bump since the latest
reachable stable `vMAJOR.MINOR.PATCH` tag wins. Release notes include the commit
subjects since that tag. Prerelease tags do not establish the stable baseline.
Use an explicit breaking-change subject and footer for incompatible CLI,
configuration schema, or JavaScript function contract changes.

## Automation

Merging to `main` does not publish a release. When ready, open **Actions >
Publish release > Run workflow**, select `main`, and start the workflow.
It reruns quality checks, calculates the next version from all commits since
the previous stable tag, builds all release archives, and publishes a GitHub
Release only when all builds succeed. No release is created if the unreleased
commits are only documentation or other non-release changes.
Releases are serialized; in-progress releases are not cancelled by later pushes.
Dispatch on branches other than `main` skips publication. The workflow uses the
commit selected when you click the button, not later merges during the build.

The release contains:

- Linux AMD64 and ARM64 archives, built on native Ubuntu runners.
- macOS AMD64 and ARM64 archives, built on native macOS runners.
- Windows AMD64 archive, built with MinGW GCC on a Windows runner.
- SHA-256 checksums, a README, and the project license.

All archives are `.tar.gz` (also supported by Windows `tar`). Tree-sitter requires
CGO, so each build uses a native C compiler instead of CGO-disabled cross-builds.
Linux binaries target the runner's glibc environment; these are not advertised
as portable musl/Alpine binaries. macOS binaries are unsigned and not notarized.
`guardrail --version` reports the release tag and source commit; local builds
report `dev (unknown)`.

Only the final publication job has `contents: write`. The GitHub CLI creates
the tag at the tested commit and uploads assets. Nothing is published to
container registries, package registries, or other services.

## Repository settings

After pushing these files, maintainers must configure GitHub:

1. Enable Actions and allow GitHub Release creation with `GITHUB_TOKEN`.
2. Enable squash merging and use the PR title as the squash commit title.
3. Protect `main`: require the Quality checks, reviews, and an up-to-date branch.
4. Ensure the repository has access to the native ARM64 and macOS Intel runner
   labels in the workflows. Runner availability and billing depend on repository
   visibility and account entitlement.

These settings are not changed by workflow files. A failed quality/build job
does not publish a release. Fix the failure and rerun the workflow. If publication
was interrupted after GitHub created a release/tag, inspect and complete that
release manually rather than deleting a published tag.

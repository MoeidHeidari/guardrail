# JSON dependency policy

[Catalog](../README.md) | [Rule API](../../docs/RULES.md)

`no-unapproved-dependencies` demonstrates a configurable dependency denylist
for npm `package.json` manifests. The initial examples are `left-pad` and
`request`; edit `forbidden` to define your organization's actual policy.
These examples are not a vulnerability database or an endorsement of alternatives.

## Install

Choose the language-specific `dependency-policy.yml` rule set from the
[catalog](../README.md#language-rule-sets). Copy it to
`.guard/<language>/dependency-policy.yml` in your context, and copy the shared
[functions](../functions/dependency-policy/) to `.guard/functions/dependency-policy/`.
Review existing destinations before copying. YAML references match these paths.
See the [adoption instructions](../README.md#validate-and-adopt) for commands.
Use the JavaScript rule set for npm manifests in either JS or TS projects.
Standard analysis exclusions remain in effect.

## Behavior and limits

Checks direct entries in dependencies, devDependencies, optionalDependencies,
and peerDependencies, including npm aliases such as `npm:request@2`.
An approved dependency passes; an unapproved entry produces one finding per
section at line 1, column 1. That is a manifest-level location, not the precise
dependency key location.

Malformed JSON, non-object manifests/sections, and non-string versions are
execution errors, never successful checks. JavaScript JSON parsing accepts
duplicate keys with last-key-wins behavior. Lockfiles, transitive dependencies,
version ranges, registries, vulnerabilities, licenses, and other languages'
manifest formats are outside this rule's scope.

## Validation and contribution

Run `go test ./internal/analysis -run '^Test(Engineering|Dependency)'`.
[Tests](../../internal/analysis/engineering_packs_test.go) include denied names,
npm aliases, approved entries, and invalid manifest structures.

Maintained by Guardrail contributors under [Apache License 2.0](../../LICENSE).
Compatible with the version 1 YAML and located JS findings API. Review scripts
before execution; restricted host APIs are not a secure sandbox.

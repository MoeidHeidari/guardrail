# Security policy

[Documentation index](README.md)

## Report vulnerabilities privately

Send potential vulnerabilities to
[moeidheidari73@gmail.com](mailto:moeidheidari73@gmail.com)
with the subject `Guardrail security report`.

Do **not** open a public issue or PR containing exploit details before coordinated
disclosure. There is no guaranteed response deadline or bug-bounty program.

Please include:

- The affected commit or version and how Guardrail was built/run.
- OS, terminal, and other relevant environment details.
- A description of the impact and conditions required.
- Minimal reproduction steps using synthetic, non-sensitive files.
- Any suggested mitigation or patch, if available.

Do not send credentials, private repository contents, or unnecessary personal data.
If sensitive material is essential, arrange a suitable exchange method first.

## Support scope

Guardrail is in early development. There are no published long-term support or
security-backport commitments. Reports should identify whether the problem affects
current `main`; fixes will normally target ongoing development.

## Coordinated disclosure

The maintainer will assess the report, request clarification if necessary, and
coordinate remediation and disclosure with the reporter where practical.
Avoid public disclosure of sensitive details while a fix is being coordinated.

Routine bugs, feature requests, and non-sensitive usability issues belong in
the [issue tracker](https://github.com/MoeidHeidari/guardrail/issues).

## Current operational boundaries

- Watching observes saved filesystem state; it is not a sandbox or access-control
  system and does not prevent an editor or agent from writing files.
- Entries in the chosen context include hidden and Git-ignored paths unless
  `.guardignore` excludes them. Choose a context you intend to display.
- Screenshots, recordings, and logs can expose file names, paths, and content.
  Redact them before sharing.
- Preview reads are bounded and non-regular/symlink files are rejected.
- Resource limits and event overflow are reported as errors; automatic recovery
  is not implemented.
- JavaScript rules run in embedded Goja without host filesystem/network/process
  APIs. Execute only trusted repository configuration: execution deadlines and
  stack limits are not process/memory isolation or a secure sandbox.
- Rule scripts are contained in `.guard/` or `.guardrail/`; the watcher still
  displays all context entries. See [Rules and checking](RULES.md) for limits.
- Checks do not block edits or merges by themselves. Configure CI exit-code
  handling explicitly; agent/editor integrations are not implemented.

# Security policy

## Scope

`workbench-mcp` is designed as a **single-user, local-first** stdio MCP server.
It does not open an HTTP listener and v0 has no authentication. The threat
model assumes the host machine and the MCP client process that launches
Workbench are trusted by the user.

Workbench stores artifacts as Markdown files under `WORKBENCH_ARTIFACT_DIR`
(`docs/artifacts` by default) and keeps process context in memory.

Issues we treat as security-relevant:

- Introduction of an unintended network listener or remote transport on `main`.
- Accidental disclosure of credentials, local secrets, or sensitive file
  contents to logs or tool results.
- Path traversal in artifact IDs, artifact file handling, or resource URI
  handlers.
- Capability gating bugs that expose selected-artifact tools or resources when
  no artifact is selected.
- Input validation gaps that allow other tenants' data to be read or written
  once multi-tenancy lands (post-v0).

Out of scope for v0:

- Attacks from a compromised host or malicious local user account.
- Behavior caused by pointing `WORKBENCH_ARTIFACT_DIR` at a sensitive directory.
- Network-level attacks against custom transports added outside `main`.
- Behaviour explicitly opt-in via documented environment variables.

## Reporting

Email the maintainer at the address listed on the GitHub profile
([@manuelibar](https://github.com/manuelibar)) or open a private security
advisory via GitHub's "Report a vulnerability" flow. Please **do not** open a
public issue for vulnerabilities.

We aim to acknowledge within 5 business days and to ship a fix or mitigation
within 30 days for high-severity issues.

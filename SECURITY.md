# Security policy

## Scope

`workbench-mcp` is designed as a **single-user, local-first** MCP server bound
to `127.0.0.1`. v0 has no authentication. The threat model assumes the host
machine is trusted by the user.

Issues we treat as security-relevant:

- Bypasses of localhost binding (e.g. unintended `0.0.0.0` listening).
- Accidental disclosure of credentials or DSN to logs.
- SQL injection via tool arguments that pass through to Postgres.
- Path traversal in resource URI handlers.
- Input validation gaps that allow other tenants' data to be read or written
  once multi-tenancy lands (post-v0).

Out of scope for v0:

- Network-level attacks against `127.0.0.1` deployments.
- Behaviour explicitly opt-in via documented environment variables.

## Reporting

Email the maintainer at the address listed on the GitHub profile
([@manuelibar](https://github.com/manuelibar)) or open a private security
advisory via GitHub's "Report a vulnerability" flow. Please **do not** open a
public issue for vulnerabilities.

We aim to acknowledge within 5 business days and to ship a fix or mitigation
within 30 days for high-severity issues.

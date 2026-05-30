# 4. Keep integration credentials in MCP stdio configuration

Date: 2026-05-19

## Status

Accepted

## Context

Integrations such as GitHub require organization names, tokens, and other configuration. Exposing configuration tools like `github.configure` makes credentials part of the runtime agent action surface even though configuring credentials is not usually part of the task the agent is performing.

## Decision

Integration configuration is supplied by the MCP client/server configuration, especially stdio environment variables, not by always-visible runtime tools. For GitHub, Workbench reads:

- `WORKBENCH_GITHUB_ORG`
- `WORKBENCH_GITHUB_TOKEN`, falling back to `GITHUB_TOKEN`

Workbench may expose a read-only resource such as `workbench:///github/config` that reports whether the integration is configured without leaking the token.

## Consequences

The agent tool surface stays focused on task-relevant actions. Credentials are managed by the host configuration where they can be reviewed, rotated, and removed outside the agent loop. The tradeoff is that changing integration configuration requires updating the MCP server configuration and restarting/reloading the server instead of calling a tool.

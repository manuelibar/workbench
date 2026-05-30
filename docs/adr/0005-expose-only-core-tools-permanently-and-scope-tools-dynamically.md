# 5. Expose only universal tools permanently and scope the rest dynamically

Date: 2026-05-19

## Status

Accepted

## Context

Workbench changes the agent's working surface according to current scope. If every possible integration, project-management action, skill-maintenance operation, task mutation, and background operation is always visible, the tool list becomes noisy and agents are encouraged to take unrelated actions.

At the same time, a few tools are universal because they are needed to orient, report, or query Workbench regardless of the selected project.

## Decision

Only universal tools are permanent. The current core set is:

- `refresh`: apply selected scope, reconcile capabilities, emit MCP list-changed notifications, and return a concise navigation briefing.
- `feedback`: report problems or observations; Workbench stores them as queryable knowledge input.
- `query`: query Workbench knowledge captured from feedback and notes.

Other tools are registered dynamically by `refresh` according to active scope.

Current examples:

- before project selection: project/namespace/role discovery and creation tools can be surfaced
- after project selection: task, board, and project-scoped workflow tools can be surfaced
- future: background-manager, registry-manager, integration-manager, or skill-maintenance tools should appear only when the active scope/configuration makes them relevant

## Consequences

The visible tool surface better reflects the agent's current job. This reduces accidental misuse and reserves attention for the active namespace/project/task.

The cost is that MCP clients must support capability list updates. Workbench should lean on MCP list-changed notifications instead of trying to make `refresh` return an exhaustive capability catalog.

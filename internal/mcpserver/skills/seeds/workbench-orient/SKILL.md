# Workbench — Orientation

## Trigger

Use this skill when starting a new session, re-orienting after context
compaction, or when unsure what capabilities are available for your current task.

## Skip

Skip if you have called `refresh()` this session and your project context
has not changed.

---

You are connected to a **Workbench MCP server** — a skill distribution layer
that delivers your full working context in a single call.

## Getting your context

```
refresh()                    # base orientation only
refresh(project_id="<id>")   # scoped to a specific project
```

`refresh()` returns:

- **`selection`** — which project/namespace/role/board (if any) is active
- **`overview`** — the current scope briefing, with the same logical shape as
  `workbench:///scope/overview`
- **`overview_uri`** — the read-only URI for re-reading that briefing later
- **`skills[]`** — every relevant skill bundle with instructions already
  rendered inline. Read them now. No further resource calls required for the
  happy path.

In the happy path, do not immediately read `workbench:///scope/overview` after
`refresh()`; the same briefing is already inline. Use the resource only when you
need to re-read current scope without changing selection or synchronizing MCP
capabilities again.

## Selecting a project

Use `project.list` to find available projects, then pass the ID to `refresh`.
When a project is selected, project-scoped skills appear in `skills[]`.
Today that includes `workbench-system-prompt` plus concrete language guidance
such as `go-coding-guidelines`.

## After context compaction

If your MCP client preserved the same Workbench process and scope, read
`workbench:///scope/overview` to restore the latest briefing without mutating or
synchronizing anything. If the process restarted, the selected scope changed, or
you need fresh MCP list-changed synchronization, call `refresh()` again (with the
same `project_id` if applicable).

## Resources — on-demand re-fetch

`workbench:///` resources are available for targeted lookups when you need
to re-fetch a specific piece of content without a full refresh:

| URI | When to use |
|---|---|
| `workbench:///scope/overview` | Re-read the current scope briefing without refresh side effects |
| `workbench:///projects/{id}` | Raw project details |
| `skill://{name}/SKILL.md` | Re-fetch one skill after compaction |

Full capability discovery comes from MCP-native list methods (`tools/list`,
`resources/list`, and prompt list methods when prompts are enabled), not a
separate Workbench manifest resource.

## Tools

| Tool | When to call |
|---|---|
| `refresh` | Session start, project change, after compaction |
| `project.list` | Finding a project ID to pass to refresh |
| `project.create` | Setting up a new project context |
| `project.delete` | Removing a project |

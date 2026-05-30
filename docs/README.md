# Workbench Docs

This documentation is the source of truth for the Workbench ecosystem.

Purpose:
- Act as the artifact repository (Confluence-like) for decision-making.
- Preserve traceability from problem -> research -> RFC -> specification -> slices -> backlog.
- Keep Workbench configurable (not hardcoded to one personal workflow).

Initial structure:
- `artifacts/` formal planning and architecture artifacts
- `adr/` architecture decision records
- `rfc/` request for comments
- `specs/` solution and implementation specifications
- `slices/` vertical slices and delivery units
- `backlog/` prioritized work index and trace links

Current priority:
1. Backlog Service (domain + API + traceability)
2. Documentation Service (artifact model + publishing)
3. Workbench MCP Core (project + namespace + role selection)

Documentation ownership:
- Cross-system contracts belong in Workbench first, because Workbench is the
  MCP entry point and gateway for agents.
- Service repos may keep their own RFCs, specs, and implementation artifacts,
  but service-local docs should point back to the Workbench master RFC when the
  behavior affects multiple systems or the agent-facing workflow.

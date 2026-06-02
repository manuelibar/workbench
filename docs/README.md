# Workbench Docs

These docs follow a Diataxis shape and describe the foundation that belongs on
`main`: scope, capability planning, artifact contracts, the storage-service
client boundary, and the epic branch workflow for bootstrapping Workbench MCP
harness features.

## Explanation

- [Context-window thesis](explanation/context-window-thesis.md)
- [Progressive disclosure](explanation/progressive-disclosure.md)

## Reference

- [Scope contract](reference/scope-contract.md)
- [MCP runtime structure](reference/mcp-runtime-structure.md)
- [Artifact conventions](reference/artifact-conventions.md)
- [Error handling](reference/error-handling.md)

## How-to

- [Epic branch workflow](how-to/epic-branch-workflow.md)

## Artifacts

Artifacts are typed Markdown documents stored either in [artifacts/](artifacts/)
or through the autonomous storage service. They use typed frontmatter and
contract sections so later branch drilling sessions can start from durable
working documents instead of chat history.

Epic packets must be self-contained. RFC artifacts act as living drill hubs
that link to research and any action artifacts created from the drilling
session.

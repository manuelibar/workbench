# Workbench Docs

These docs follow a Diataxis shape and describe the foundation that belongs on
`main`: context, capability planning, file-backed artifacts, and the epic
branch workflow for bootstrapping Workbench MCP harness features.

## Explanation

- [Context-window thesis](explanation/context-window-thesis.md)
- [Progressive disclosure](explanation/progressive-disclosure.md)

## Reference

- [Context contract](reference/context-contract.md)
- [Artifact conventions](reference/artifact-conventions.md)
- [Error handling](reference/error-handling.md)

## How-to

- [Epic branch workflow](how-to/epic-branch-workflow.md)

## Artifacts

Artifacts are flat Markdown files in [artifacts/](artifacts/). They use typed
frontmatter and contract sections so later branch drilling sessions can start
from durable working documents instead of chat history.

Epic packets must be self-contained. RFC artifacts act as living drill hubs
that link to research and any action artifacts created from the drilling
session.

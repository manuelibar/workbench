# Context-window Thesis

Workbench treats the model context window as the scarce runtime resource.
The server should not expose every project system by default. It should expose
the smallest current scope document and the capabilities that are relevant to
that scope.

The foundation keeps only two live scope fields:

- `focus`: a short steering string for the current work.
- `artifact_id`: the artifact in scope when work is centered on one document.

Everything else is deferred to epic branches. Namespaces, roles, memory,
knowledge, backlog, AFK loops, and persisted sessions all need stronger
contracts before they belong on `main`.

The scope document is also an MCP resource at `workbench:///scope`. The
`contextualize` tool returns the exact same raw document so a client can
verify that tool and resource views match.

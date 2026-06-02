# Progressive Disclosure

Workbench exposes capabilities in layers.

The base layer is enough to orient the agent and find or create artifacts.
Placing an artifact in scope reveals upload and the exact artifact resource.
Clearing the artifact hides that scoped surface again.

This keeps the visible tool list aligned with the current task without losing
recoverability. When capability visibility changes, Workbench waits for the
client to relist the changed MCP categories. If the relist does not arrive
before the configured timeout, the `contextualize` result includes a full
fallback capability snapshot.

Progressive disclosure is a contract, not a UI trick: a hidden feature belongs
on an epic branch until the branch defines the scope, artifacts, tools, and
merge criteria that make it safe to expose on `main`.

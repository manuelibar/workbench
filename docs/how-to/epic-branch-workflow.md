# Epic Branch Workflow

Epic branches start from the cleaned foundation `main`, not from the archived
integrated snapshot.

Each epic branch starts with a kickoff artifact packet:

- charter
- problem statement
- assumptions and risks
- research note placeholder
- RFC draft

The packet defines the intended direction before old code is ported from
archive refs. This prevents archived implementation details from silently
setting the new contract.

The expected branch loop is:

1. Select the epic RFC or charter with `context`.
2. Drill the problem and constraints until the packet is coherent.
3. Port useful code or docs from `archive/*` only when the packet says why.
4. Add tests that prove the feature belongs on `main`.
5. Merge when the feature is functional through the context/artifact kernel.

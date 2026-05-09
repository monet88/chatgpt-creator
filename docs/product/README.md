# Product Docs

This directory is reserved for product-contract docs split by domain.

Current product truth still lives in the existing root docs:

- `docs/project-overview-pdr.md`
- `docs/system-architecture.md`
- `docs/codebase-summary.md`
- `docs/testing-guide.md`
- `docs/api-contract.md`

When a selected story needs a more focused contract, create a domain file here
instead of growing the broader docs indefinitely. Name files by real product
domains, for example `cli-registration.md`, `codex-token-output.md`,
`temp-mail-api.md`, or `worker-ui.md`.

Do not create placeholder domain files just to fill the folder.

## Update Rule

When behavior changes:

1. Update the affected product doc.
2. Update or create the story packet.
3. Update `docs/TEST_MATRIX.md`.
4. Record a decision if the change affects architecture, scope, risk, or a
   previously settled product rule.

# 0004 Adopt Harness For Existing Project

Date: 2026-05-09

## Status

Accepted

## Context

`chatgpt-creator` already has implementation, docs, tests, and scripts. Harness
v0 was installed with `--merge`, so existing files stayed intact and new
harness files were added under `docs/` plus `scripts/README.md`.

The upstream harness docs assumed a blank repository. Keeping them unchanged
would confuse future agents about the current Go CLI, Node automation scripts,
and standalone Cloudflare Worker app.

## Decision

Adopt Harness v0 as a workflow overlay, not as a replacement for the existing
project structure.

Current source-of-truth docs remain:

- `README.md`
- `docs/project-overview-pdr.md`
- `docs/system-architecture.md`
- `docs/codebase-summary.md`
- `docs/testing-guide.md`
- `docs/api-contract.md`

Harness files provide intake, story packets, validation matrix, templates,
backlog, and future decision records.

## Consequences

Positive:

- Existing runtime docs stay authoritative.
- Future work can use story packets and test matrix rows when risk justifies it.
- The harness no longer claims this repo has no app, stack, or validation.

Tradeoffs:

- Some upstream harness decisions remain as historical background.
- Product docs are split across existing root docs and the newer `docs/product/`
  index until a selected story needs more domain-specific files.

---
name: feature-development-with-tests-and-docs
description: Workflow command scaffold for feature-development-with-tests-and-docs in chatgpt-creator.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /feature-development-with-tests-and-docs

Use this workflow when working on **feature-development-with-tests-and-docs** in `chatgpt-creator`.

## Goal

Implements a new feature or major enhancement, including code, tests, and synchronized documentation updates.

## Common Files

- `internal/**/*.go`
- `cmd/**/*.go`
- `cloudflare-temp-mail/src/**/*.ts`
- `cloudflare-temp-mail/tests/**/*.ts`
- `cloudflare-temp-mail/tests/**/*.spec.ts`
- `docs/**/*.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Implement feature in source files (e.g., internal/*, cmd/*, cloudflare-temp-mail/src/*)
- Add or update corresponding tests (e.g., *_test.go, tests/*.ts)
- Update or add relevant documentation (e.g., docs/*.md, README.md, TODO.md, docs/project-changelog.md, docs/project-roadmap.md, docs/system-architecture.md)

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.
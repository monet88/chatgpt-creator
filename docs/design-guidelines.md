# Design Guidelines

## Documentation Design Goals

This repository currently exposes a CLI workflow. Documentation should optimize for operator clarity and maintenance accuracy.

## Content Principles

1. Document only behavior verified in source files under `cmd/` and `internal/`.
2. Keep user-facing guides short; push implementation details into architecture docs.
3. Mark assumptions and unknowns explicitly when behavior depends on third-party endpoints.
4. Keep naming consistent with actual code symbols and config keys.

## Structure Guidelines

- `README.md`: quick start + verified feature set.
- `docs/project-overview-pdr.md`: requirements and scope boundaries.
- `docs/system-architecture.md`: runtime flow and component interactions.
- `docs/codebase-summary.md`: package-level snapshot.
- `docs/code-standards.md`: implementation conventions grounded in current code.

## Example Quality Bar

Use examples that are currently runnable:
- `go run cmd/register/main.go`
- `config.json` keys: `proxy`, `output_file`, `default_password`, `default_domain`

Avoid documenting unimplemented flags, APIs, or automation pipelines.

## Maintenance Checklist

Before updating docs:
- Verify file paths still exist.
- Verify function/package names still exist.
- Verify config keys are unchanged.
- Remove stale sections instead of leaving placeholders.

## Explicit Scope Boundary

These guidelines apply to documentation quality and structure only. They do not define product behavior changes.

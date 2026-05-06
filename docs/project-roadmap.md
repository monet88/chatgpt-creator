# Project Roadmap

## Current Phase Snapshot

| Phase | Status | Notes |
|---|---|---|
| Docs Initialization | Completed | Core docs set created and aligned with codebase snapshot (2026-05-06) |
| Stability Hardening | Planned | Add tests, improve retry/backoff strategy, reduce flow fragility |
| Operability Improvements | Planned | Better observability, config ergonomics, failure categorization |

## Milestone Backlog (Implementation-Oriented)

### M1 — Test Baseline
- Add unit tests for config validation and utility generators.
- Add flow-level integration scaffolding with safe mocks.
- Define minimum CI command set (`go test`, `-race`, coverage).

### M2 — Network/Flow Resilience
- Normalize retry strategy per step (status-based + transient error handling).
- Reduce dependence on fragile page selectors where possible.
- Improve error typing (instead of broad string matching).

### M3 — Operational Safety
- Add dry-run/simulation mode for non-production validation.
- Add redaction rules for logs and outputs.
- Add configurable timeout budgets per flow stage.

## Documentation Maintenance Cadence

- Update `README.md` after any user-facing CLI/config change.
- Update architecture and codebase summary after internal flow/package changes.
- Review roadmap status weekly or after significant merges.

## Explicit Unknowns

- Priority order between test investment and feature additions is not yet defined by product owner.

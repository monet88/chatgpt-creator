# Project Roadmap

## Current Phase Snapshot

| Phase | Status | Notes |
|---|---|---|
| Docs Initialization | Completed | Core docs established |
| Test Baseline | Completed | Offline tests added for config/email/register/CLI |
| Operability Improvements | Completed | Cobra CLI, typed failures, runtime caps, redaction, JSON summary |
| Continuous Hardening | In Progress | Maintain resilience against upstream drift and selector changes |

## Milestones

### M1 — Testability and Seams (Completed)
- Added mock-first seams for register batch flow.
- Added parser/unit tests and fake flow tests.

### M2 — CLI and Runtime Safety (Completed)
- Added non-interactive CLI via Cobra.
- Added validation + exit code mapping.
- Added bounded attempts and failure threshold controls.

### M3 — Secure Output and Automation (Completed)
- Added redaction for password/proxy/token-like log content.
- Added JSON run summary with failure taxonomy and stop reason.
- Preserved `email|password` output file compatibility.

### M4 — Ongoing Maintenance (In Progress)
- Keep docs synced with behavior across both products (Go CLI + standalone Worker app).
- Maintain Worker API contract and deployment/testing docs as Cloudflare routes/features evolve.
- Expand drift-detection and integration safety checks when needed.

## Maintenance Cadence

- Update changelog after each feature/fix batch.
- Re-run race/coverage/vet baseline before release.
- Re-check documentation accuracy after flow updates.

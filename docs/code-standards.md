# Code Standards

## Scope

Applies to:

- Go source under `cmd/` and `internal/`.
- Standalone Worker source under `cloudflare-temp-mail/src/`.

## Core Principles

- Keep boundaries small and testable.
- Prefer local seams over broad framework abstractions.
- Keep runtime behavior bounded and cancellation-aware.

## CLI and Config Rules

- CLI parsing currently lives in `cmd/register/main.go` (with `cmd/register/command.go` as the thin process entry wrapper).
- Config precedence must remain: `defaults < file < env < flags`.
- Validation errors must map to explicit non-zero exit codes.

## Failure Model Rules

- Use typed failures from `internal/register/failures.go`.
- Wrap step errors with `WrapFailure(step, status, err)` where possible.
- Avoid free-form substring logic for control flow when typed error is available.

## Concurrency and Runtime Controls

- Use context-aware waits; avoid raw `time.Sleep` in retry paths.
- Maintain explicit max-attempt and failure-threshold controls.
- Protect shared logging and file writes with mutexes.

## Output and Logging Rules

- Never print plain passwords to console.
- Redact proxy credentials and token-like substrings before logging.
- In JSON mode: stdout reserved for machine-readable summary; diagnostics routed to stderr.

## Test Rules

- Default tests must stay offline.
- Use table-driven tests for parsers and validation behavior.
- Keep fake dependencies around network boundaries.
- Validation baseline (root CLI):
  - `go test ./...`
  - `go test -race ./...`
  - `go test -cover ./...`
  - `go vet ./...`
- Validation baseline (standalone Worker):
  - `npm --prefix cloudflare-temp-mail run build`
  - `npm --prefix cloudflare-temp-mail run test`

## Documentation Sync Rule

When behavior changes in CLI, failure handling, runtime controls, or output policy, update:
- `README.md`
- `docs/codebase-summary.md`
- `docs/system-architecture.md`
- roadmap/changelog docs

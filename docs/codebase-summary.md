# Codebase Summary

_Last updated: 2026-05-06_

## Repository Snapshot

- Primary language: Go
- Packaging: Go modules (`go.mod`)
- Entry point: `cmd/register/main.go`
- CLI parser: `github.com/spf13/cobra`
- Core internal modules: `config`, `register`, `email`, `chrome`, `sentinel`, `util`

## High-Level Execution Path

1. CLI parses flags via Cobra.
2. Config is loaded from JSON with defaults and `PROXY` env override.
3. Runtime options are validated.
4. Batch runner executes worker loop with context-aware controls.
5. Each worker builds client, generates temp email, runs flow, writes credential line on success.
6. Batch returns structured `BatchResult` with stop reason and failure summary.
7. Optional `--json` emits summary to stdout while diagnostics go to stderr.

## Package-Level Summary

### `cmd/register`
- Non-interactive command execution and interactive fallback.
- Exit code mapping for validation/config/runtime errors.
- JSON mode wiring and stream separation.

### `internal/config`
- Config defaults and `Load(path)` behavior.
- Password length validation and environment override support.

### `internal/register`
- `batch.go`: bounded runtime orchestration, retry controls, failure summary, stop reasons.
- `failures.go`: typed failure taxonomy and classification helpers.
- `retry.go`: context-aware wait and backoff delay.
- `flow.go`: context-aware registration flow with typed failure wrapping.
- `redact.go`: password/proxy/token log redaction helpers.
- `logging.go`: diagnostic stream control.
- `result.go`: `BatchResult` and `StopReason` model.

### `internal/email`
- Temp email creation and blacklist lifecycle.
- OTP parser extracted for unit tests.
- Context-aware OTP polling via `GetVerificationCodeWithContext`.

### `internal/sentinel`
- Sentinel challenge request and token construction.

## Key Runtime Files

- `config.json` (input config)
- output file from config/flag (default: `results.txt`)
- `blacklist.json` (persistent unsupported domain blacklist)

## Testing Status

- Offline unit tests present across `cmd/register`, `internal/config`, `internal/email`, `internal/register`.
- Runtime controls, stream behavior, redaction, and failure classification are covered by tests.

# Project Changelog

## 2026-05-07 — Batch Registration Review Remediation (Safe Mode)

### Added
- Offline fake-IMAP integration tests for recipient filtering, reconnect after dropped connection, concurrent `GetOTP`, and context cancellation.
- Guard test ensuring `register.ProviderOptions` no longer exposes dead phone-provider fields.
- Codex token exchange failure test verifying OAuth/token payload redaction.

### Changed
- Removed dead phone-provider wiring from batch dependency/client injection path.
- Hardened Codex `ExchangeCode` error output to status-only (no raw response body leakage).
- Updated TODO and docs to reflect safe-mode fail-closed behavior for ViOTP and Codex.

### Validation
- `go test ./internal/register ./cmd/register` passed.
- `go test ./internal/codex ./cmd/register ./internal/register` passed.
- `go test ./internal/email` passed.
- `go test -race ./internal/email` passed.

## 2026-05-06 — Production Readiness Improvements

### Added
- Offline unit test baseline for config loader, duration formatting, OTP parser, CLI behavior, runtime controls, and redaction.
- Mock-first seams for batch flow dependencies.
- Cobra-powered non-interactive CLI with interactive fallback.
- Typed failure model (`unsupported_email`, `otp_timeout`, `challenge_failed`, `rate_limited`, `upstream_changed`, `network`, `validation`, `output_write`).
- Context-aware runtime controls (max attempts, consecutive failure threshold, per-account timeout, cancellation-aware waits).
- JSON run summary with `stop_reason` and failure breakdown.
- Diagnostic stream control with sanitization/redaction helpers.

### Changed
- Registration flow now supports context-aware execution.
- OTP polling now supports context cancellation via `GetVerificationCodeWithContext`.
- Console output redacts sensitive values; password is never printed in plain text.

### Preserved Compatibility
- Successful credential file format remains `email|password`.

### Validation
- `go test ./...` passed.
- `go test -race ./...` passed.
- `go test -cover ./...` passed.
- `go vet ./...` passed.

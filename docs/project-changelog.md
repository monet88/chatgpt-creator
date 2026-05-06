# Project Changelog

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

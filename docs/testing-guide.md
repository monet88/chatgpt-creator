# Testing Guide

## Goals

- Keep default tests offline and deterministic.
- Validate CLI behavior, failure typing, runtime caps, and output safety.

## Core Commands

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

## Coverage Areas

- `internal/config`: load defaults, JSON parsing, invalid config, env override.
- `internal/email`: default-domain temp email shape, OTP parser edge cases.
- `internal/register`:
  - fake dependency seams
  - credential write/no-write behavior
  - runtime stop controls (max attempts, threshold, cancellation)
  - failure classification and redaction helpers
- `cmd/register`: flag parsing, precedence, exit codes, interactive fallback, JSON output behavior.

## Offline Test Policy

- Do not hit live external endpoints in default `go test ./...`.
- Use fakes/stubs around `RunBatch` and flow dependencies.

## When to Add Integration Tests

Add opt-in integration tests only when needed to validate upstream contract drift. Keep them separated from default offline suite.

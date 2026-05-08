# Testing Guide

## Goals

- Keep default tests offline and deterministic.
- Validate CLI behavior, failure typing, runtime caps, and output safety.

## Core Commands

### Root CLI

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

### Standalone Worker app

```bash
npm --prefix cloudflare-temp-mail run build
npm --prefix cloudflare-temp-mail run test
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

- Do not hit live external endpoints in default `go test ./...` or `npm --prefix cloudflare-temp-mail run test`.
- Use fakes/stubs around `RunBatch`, Worker env bindings, and flow dependencies.

## When to Add Integration Tests

Add opt-in integration tests only when needed to validate upstream contract drift. Keep them separated from default offline suite.

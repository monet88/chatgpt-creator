# Codebase Summary

_Last updated: 2026-05-06_
_Source: `repomix-output.xml` generated from repository root_

## Repository Snapshot

- Primary language: Go
- Packaging: Go modules (`go.mod`)
- Entry point: `cmd/register/main.go`
- Internal modules: `config`, `register`, `email`, `chrome`, `sentinel`, `util`

## High-Level Execution Path

1. CLI starts and prints banner.
2. Config is loaded from `config.json` with defaults and environment override support in the config loader.
3. User provides runtime inputs (target accounts, workers, optional overrides).
4. Batch orchestration starts worker goroutines.
5. Each worker creates a TLS-backed HTTP client and email identity.
6. Worker executes the registration flow engine.
7. On success, account is appended to output file.
8. On failure, slot is returned for retry; unsupported domains can be blacklisted.

## Package-Level Summary

### `cmd/register`
- CLI orchestration only.
- No business logic beyond prompts and argument collection.

### `internal/config`
- `Config` struct: `proxy`, `output_file`, `default_password`, `default_domain`.
- `Load(path)` handles defaults, JSON parse, password-length validation, and environment proxy override behavior.

### `internal/register`
- `batch.go`: concurrency, retry accounting, summary output, blacklist trigger.
- `client.go`: session construction with TLS profile mapping, cookie jar, proxy, headers.
- `flow.go`: multi-step account flow including OTP and sentinel-backed account creation.

### `internal/email`
- Domain selection and temp email generation.
- OTP polling with HTML scraping from generator.email mailbox pages.
- Domain blacklist lifecycle: load on init, save on updates.

### `internal/sentinel`
- Fetch challenge token from Sentinel endpoint.
- Build proof-of-work/requirements token.
- FNV-1a helper for token generation.

### `internal/chrome`
- Browser profile metadata and mapping to `tls-client` profiles.

### `internal/util`
- Random string, UUID, name, birthdate, password generation.
- Datadog-compatible trace headers.

## External Network Dependencies

- `https://chatgpt.com`
- `https://auth.openai.com`
- `https://sentinel.openai.com`
- `https://generator.email`

## Key Runtime Files

- `config.json` (input config)
- output file from config (default: `results.txt`)
- `blacklist.json` (generated/updated at runtime)

## Dependency Highlights (from `go.mod`)

- `github.com/bogdanfinn/tls-client`
- `github.com/bogdanfinn/fhttp`
- `github.com/PuerkitoBio/goquery`
- `github.com/brianvoe/gofakeit/v7`
- `github.com/google/uuid`

## Explicit Unknowns

- No built-in health-check for upstream endpoint drift.
- No test suite present in current repository snapshot.

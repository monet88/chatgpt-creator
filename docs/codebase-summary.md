# Codebase Summary

_Last updated: 2026-05-07_

## Repository Snapshot

- Primary language: Go (root CLI) + TypeScript (standalone Cloudflare Worker app)
- Root packaging: Go modules (`go.mod`)
- Root entry point: `cmd/register/main.go`
- Standalone app location: `cloudflare-temp-mail/`
- Standalone app runtime: Cloudflare Workers + D1 + R2 + Email Routing

## Top-Level Products

### 1) `chatgpt-creator` CLI (Go)

- Purpose: batch registration CLI with OTP automation.
- Main path: `cmd/register` + `internal/*`.
- External dependencies: OpenAI auth/sentinel endpoints and `generator.email`.
- Output artifacts: credential output file + `blacklist.json`.

### 2) `cloudflare-temp-mail` (TypeScript Worker)

- Purpose: standalone temp-mail API/UI service, intended to be consumed over HTTP.
- Main path: `cloudflare-temp-mail/src/worker.ts`.
- API prefix: `/api/v1` (`cloudflare-temp-mail/src/config/app-config.ts`).
- Data/storage: D1 metadata + R2 payload objects.
- Inbound handling: `email()` handler for Cloudflare Email Routing.
- Retention: scheduled cleanup via cron trigger (`*/30 * * * *` in `wrangler.toml`).

## High-Level Execution Paths

### CLI path (Go)

1. CLI parses flags via Cobra.
2. Config loads from JSON with defaults and `PROXY` env override.
3. Runtime options are validated.
4. Batch runner executes worker loop with runtime controls.
5. Worker runs registration flow + OTP polling and writes credential on success.
6. Batch returns structured summary (`BatchResult`).

### Worker path (Cloudflare)

1. Fetch handler checks optional bearer auth (`API_TOKEN`) and serves UI assets/health.
2. API routes are matched under `/api/v1` through custom router.
3. Route handlers validate input/domain and use D1/R2-backed services.
4. `email()` persists inbound message data for mailbox/OTP retrieval.
5. Scheduled handler runs cleanup to purge expired payloads.

## Validation Commands (Current)

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
npm --prefix cloudflare-temp-mail run dev
```

## API Contract Reference

- Worker API contract: `cloudflare-temp-mail/docs/api-contract.md`
- Integration boundary rule: consuming systems should use HTTP endpoints only, not Worker internals.

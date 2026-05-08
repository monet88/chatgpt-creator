# Codebase Summary

_Last updated: 2026-05-08 (web UI + panel writer)_

## Repository Snapshot

- Primary language: Go (root CLI) + TypeScript (standalone Cloudflare Worker app)
- Root packaging: Go modules (`go.mod`)
- Root entry point: `cmd/register/main.go`
- Standalone app location: `cloudflare-temp-mail/`
- Standalone app runtime: Cloudflare Workers + D1 + R2 + Email Routing

## Top-Level Products

### 1) `chatgpt-creator` CLI (Go)

- Purpose: batch registration CLI with OTP automation; also ships a built-in web UI.
- Main path: `cmd/register` + `internal/*`.
- External dependencies: OpenAI auth/sentinel endpoints and `generator.email`.
- Output artifacts: `accounts/cre/<datetime>.txt` credential files (`email|password|mailboxURL`) + `blacklist.json` + optional Codex token files under `accounts/tokens/`.
- Key internal packages:
  - `internal/register/` — batch runner, flow state machine, panel token writer.
  - `internal/email/` — OTP provider interface; `CloudflareTempMailProvider` polls Worker API.
  - `internal/web/` — embedded web UI served by the `serve` subcommand (SSE log stream, `ui.html`).

### 2) `cloudflare-temp-mail` (TypeScript Worker)

- Purpose: standalone temp-mail API/UI public service at `https://mail.monet.uno`.
- Main path: `cloudflare-temp-mail/src/worker.ts`.
- API prefix: `/api/v1` (`cloudflare-temp-mail/src/config/app-config.ts`).
- Auth: public (`AUTH_DISABLED=true`), rate-limited at 120 req/60 s per IP.
- UI pages: `/` (main inbox), `/domain` (setup guide), `/api` (API docs).
- UI assets: `src/ui/` — dark-theme Vietnamese SPA with auto-refresh, localStorage persistence, OTP display, shareable URL hash.
- Name generation: `src/lib/name-generator.ts` produces `firstname.lastname.hex` local parts.
- Data/storage: D1 metadata + R2 payload objects.
- Inbound handling: `email()` handler for Cloudflare Email Routing catch-all on `monet.uno`.
- Retention: scheduled cleanup via cron trigger (`*/30 * * * *` in `wrangler.toml`).

## High-Level Execution Paths

### CLI path (Go)

1. CLI parses flags via Cobra (`register` root command or `register serve` for web UI).
2. Config loads from JSON with defaults and `PROXY` env override.
3. Runtime options are validated.
4. Batch runner executes worker loop with runtime controls.
5. Worker runs registration flow + OTP polling and writes credential on success.
6. Optional: post-registration Codex PKCE OAuth flow extracts `access_token`/`refresh_token`.
7. Optional `--panel-output`: writes per-account `codex-{email}-{plan}.json` (panel-compatible format).
8. Batch returns structured summary (`BatchResult`).

### Web UI path (`register serve`)

1. `serve` subcommand starts HTTP server (default `:8899`) and opens browser.
2. `internal/web/server.go` serves embedded `ui.html` and exposes `/api/start`, `/api/stop`, `/api/events`.
3. Form POST to `/api/start` spawns a goroutine running `RunBatchForCLIWithProviders` with `WithDiagnosticWriter` piped to an `SSEBroker`.
4. Browser receives log lines via SSE (`EventSource`) and renders them in real-time.
5. On completion, a `done` event is broadcast with success/failed/target counts.

### Worker path (Cloudflare)

1. Fetch handler checks auth (disabled for public deployment) and serves UI pages/assets/health.
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

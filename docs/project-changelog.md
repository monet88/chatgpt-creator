# Project Changelog

## 2026-05-08 — Credential Output Structure Refactor

### Changed
- `internal/register/dependencies.go` (`appendCredential`): output line format changed from `email|password` to `email|password|mailboxURL`.
- Temp email creation signature now consistently returns `(emailAddr, mailboxURL, err)` across provider paths used by batch registration.
- `internal/email/generator.go` (`CreateTempEmail`): when `defaultDomain` is set, mailbox URL is intentionally empty (`""`) because custom domains may not map to a `generator.email` mailbox URL.
- `internal/email/cloudflare_tempmailprovider.go` (`CreateCloudflareTempEmail`): mailbox URL now points to Worker inbox hash route as `<baseURL>/#<url-escaped-email>`.

### Validation
- `go test ./...` (133 passed in 11 packages)
- `go build ./...`

## 2026-05-08 — Web UI, Panel Token Writer, OTP Flow Hardening

### Added
- `register serve` subcommand: built-in HTTP server with embedded dark-theme web UI at `:8899`. Form-based config, real-time SSE log stream, Start/Stop buttons, no external dependencies.
- `internal/web/` package: `SSEBroker`, `server.go`, `ui.html` (embedded via `//go:embed`).
- `internal/register/panel_writer.go`: writes per-account `codex-{email}-{plan}.json` matching `playful-proxy-api-panel` format. Parses `id_token` JWT claims (`account_id`, `chatgpt_user_id`, `plan_type`, `organization_id`) without signature verification (token comes directly from OpenAI auth). Optionally fetches model mapping from `/backend-api/models`.
- `--panel-output <dir>` flag (requires `--codex`): directory for per-account panel JSON files.

### Fixed
- `cloudflare-temp-mail/src/services/otp-extractor.ts`: regex `(?<![#\d])(\d{4,8})` — added `#` to lookbehind to prevent matching CSS hex colors (e.g., `#202123`) in OpenAI email templates as OTP.
- `internal/register/flow.go` (`email-verification` branch): removed explicit `sendOTP()` call; OpenAI auto-sends OTP on redirect, calling again invalidated session state and caused `invalid_state 409`.
- `internal/register/flow.go` (`register()`): added `openai-sentinel-token` header (required by `/api/accounts/user/register`).
- `internal/email/cloudflare_tempmailprovider.go` (`GetOTP`): added 60-second freshness filter rejecting stale mailbox entries from prior registration attempts.
- `internal/email/cloudflare_tempmailprovider.go` (`CreateEmail`): `json:",omitempty"` on domain field — empty domain serialized as `{}` instead of `{"domain":""}` which failed Worker's `??` operator check.

## 2026-05-08 — Cloudflare Temp Mail Public Launch & UI Overhaul

### Added
- Public deployment at `https://mail.monet.uno` (Cloudflare custom domain).
- `src/lib/name-generator.ts`: realistic `firstname.lastname.hex` email local-part generator.
- `src/ui/domain-html.ts`: step-by-step domain setup guide page at `/domain`.
- `src/ui/api-html.ts`: full API reference docs at `/api` with tabs, parameter tables, curl examples, error codes.
- `internal/util/names.go`: `ParseNameFromEmail` extracts first/last name from `firstname.lastname.hex@domain` format.
- `.dev.vars`: local API token file (gitignored).

### Changed
- `AUTH_DISABLED=true` in `wrangler.toml` — API now fully public, protected only by rate limiting.
- `wrangler.toml`: `ENABLED_DOMAINS=monet.uno`, real D1 `database_id` wired.
- UI completely rewritten: dark theme, Vietnamese, sidebar layout, auto-refresh (5 s), localStorage email persistence, shareable URL hash (`/#email@domain`), OTP/2FA panel with copy button.
- Worker now serves `/domain` and `/api` routes.
- `internal/register/batch.go`: uses `ParseNameFromEmail` for consistent name–email pairing; falls back to `gofakeit` for non-matching formats.
- Email generate route uses `randomLocalPart` from `name-generator.ts` instead of `tmp-hex` prefix.

### Infrastructure
- D1 database `cloudflare-temp-mail` created in APAC region.
- R2 bucket `cloudflare-temp-mail` created.
- Cloudflare Email Routing catch-all rule on `monet.uno` → Worker.
- D1 migrations applied; domain seed updated from `example.com` to `monet.uno`.

### Validation
- `npx wrangler deploy` succeeded.
- End-to-end: email generated, test email received and visible in inbox via API.

## 2026-05-07 — Cloudflare Temp Mail Hardening

### Changed
- API auth now fails closed unless `API_TOKEN` is configured or `AUTH_DISABLED=true` is explicitly set for local/dev use.
- Inbound email now accepts only issued mailboxes, preventing catch-all domain leakage.
- Mailbox delete now tombstones records and leaves bounded R2 purge to scheduled cleanup.
- Inbound storage now compensates R2 objects when D1 insert fails.
- Malformed URL path params now return `400 invalid_path_param` instead of runtime 500.
- API routes now have best-effort per-isolate rate limiting with `/health` and UI assets exempt.
- CORS is explicitly documented as same-origin/server-side only for MVP.
- Test coverage now includes `/health`, `/random-domains`, API rate limiting, cleanup multi-batch behavior, and Playwright UI smoke flow.
- Fake D1 tests now normalize SQL and fail fast on unhandled statements.

### Validation
- `npm --prefix cloudflare-temp-mail run build`
- `npm --prefix cloudflare-temp-mail test`

## 2026-05-07 — Docs Sync for Standalone Cloudflare Worker App

### Changed
- Updated `docs/codebase-summary.md` to reflect dual-product repository structure (Go CLI + standalone Worker app).
- Updated `docs/project-overview-pdr.md`, `docs/system-architecture.md`, `docs/code-standards.md`, `docs/deployment-guide.md`, and `docs/testing-guide.md` with Worker-specific scope and validation commands.
- Confirmed API contract location and integration boundary in `cloudflare-temp-mail/docs/api-contract.md`.

### Validation
- `npm --prefix cloudflare-temp-mail run build`
- `npm --prefix cloudflare-temp-mail run test`

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

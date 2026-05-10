# System Architecture

## Overview

Repository architecture now has two independent runtime surfaces:

1. Single-process Go CLI with worker goroutines (`cmd/register`, `internal/*`).
2. Standalone Cloudflare Worker app (`cloudflare-temp-mail`) with HTTP API, inbound email handler, and scheduled cleanup.

## Component Diagram

```text
CLI (cmd/register)
  -> Config Loader (internal/config)
  -> Batch Runner (internal/register/batch.go)
      -> Worker N
          -> Client Factory (internal/register/client.go)
          -> Flow Engine (internal/register/flow.go)
              -> OTP Provider (internal/email)
              -> Phone Provider (internal/phone / ViOTP)
              -> Sentinel Token Builder (internal/sentinel)
          -> [--mfa] MFA Setup (internal/mfa)
              -> Camofox Browser (internal/camofox)
                  -> chatgpt.com login + TOTP enrollment
          -> [--codex] Codex Extraction (internal/register/codex_browser.go)
              -> Camofox Browser (internal/camofox)
                  -> auth.openai.com OAuth (prompt=login, port 1455)
                  -> Callback Interceptor (internal/codex/sso.go → 127.0.0.1:1455)
                  -> Token Exchange (internal/codex/sso.go)
          -> Credential Writer (accounts/cre/<datetime>.txt)
          -> Panel Writer (accounts/tokens/*.json)
  -> BatchResult + Failure Summary
```

## Runtime Flow

1. Parse flags / choose interactive fallback.
2. Load config and apply precedence (`defaults < file < env < flags`).
3. Validate runtime inputs (fail-closed when ViOTP/Codex options are supplied in safe mode).
4. Start workers and execute attempts under options:
   - max attempts
   - max consecutive failures
   - per-account timeout
   - context cancellation
5. Classify failures into typed taxonomy.
6. On success write `email|password|mailboxURL` to output file.
7. Return `BatchResult` with `stop_reason` and `failure_summary`.

## Registration State Machine (`internal/register/flow.go`)

`runFlow` branches on the path that `authorize` redirects to:

```
visitHomepage → getCSRF → signin → authorize
                                       │
         ┌─────────────────────────────┼──────────────────┬────────────────────┐
         ▼                             ▼                  ▼                    ▼
create-account/           email-verification /       about-you           callback /
   password                   email-otp                                 chatgpt.com
         │                             │                  │
         ▼                             ▼                  ▼
    register()                  (skip register)    createAccount()    ← done
    sendOTP()                   OTP already sent         │
         │                             │                  ▼
         └──────────────────► validateOTP()           callback()
                                       │
                                       ▼
                               createAccount()   ← openai-sentinel-token REQUIRED
                                       │
                                       ▼
                                   callback()
```

### Sentinel token contracts (fixed — do not change)

| API endpoint | `openai-sentinel-token` header |
|---|---|
| `POST /api/accounts/user/register` | **NOT sent** — sending it with invalid `t` field causes 400 |
| `POST /api/accounts/create_account` | **REQUIRED** via `BuildSentinelToken` |
| `POST /api/accounts/email-otp/validate` | Not sent |

### Flow guard rule

When `authorize` returns `email-verification` or `email-otp`, OpenAI has already
dispatched the OTP (OTP-first flow). **Do not** call `register()` on this path.
Do not add navigation to `create-account/password` before `register()` on this path.

When debugging a flow regression, compare against `verssache/chatgpt-creator`
(upstream reference) before adding new steps.

## MFA Setup Flow (`--mfa`, `internal/mfa/setup.go`)

```
camofox browser (userId = email-derived, sessionKey = "mfa-setup")
  → chatgpt.com/auth/login
  → type email → [password] → [email OTP]
  → logged in

  → chatgpt.com/?action=enable&factor=totp#settings/Security
  → dismiss onboarding dialogs
  → Authenticator app → Turn on → Set up manually
  → extract TOTP secret from page text
  → click Next → type TOTP code → Verify
  → save totpSecret in Client
```

## Codex Extraction Flow (`--codex`, `internal/register/codex_browser.go`)

Fixed port 1455 — serialized by `codex1455Mu` across workers.

```
codex.InterceptCallback("127.0.0.1:1455", state, 90s)  [goroutine]

camofox browser (userId = email-derived, sessionKey = "codex-login")
  → auth.openai.com/oauth/authorize?...&prompt=login&redirect_uri=http://localhost:1455/auth/callback
  → auth.openai.com/log-in              → type email → submit
  → auth.openai.com/log-in/password     → type password → submit
  → (if) email-verification             → get OTP from mail → type → submit
  → (if) phone-verification             → ViOTP rent VN number → selectVietnam() → submit
                                          → wait SMS OTP → type → submit
  → (if) totp + totpSecret set          → GenerateTOTP() → type → submit
  → consent page                        → click Continue
  → http://localhost:1455/auth/callback?code=...
  → InterceptCallback returns code
  → ExchangeCode(code, pkce.Verifier) → POST /oauth/token
  → access_token / refresh_token / id_token
```

`redirect_uri` **must** be `http://localhost:1455/auth/callback` — `127.0.0.1` or any other port is rejected by auth.openai.com.

## Concurrency Model

- Worker pool (`maxWorkers` goroutines)
- Counters via `sync/atomic`
- Shared output/log synchronization via mutexes
- Context-aware delay/retry controls
- `codex1455Mu` serializes Codex browser extractions (port 1455 singleton)

## Failure Model

Typed kinds:
- `unsupported_email`
- `otp_timeout`
- `challenge_failed`
- `rate_limited`
- `upstream_changed`
- `network`
- `validation`
- `output_write`
- `phone_challenge`

## Observability and Output

- Diagnostics: timestamped worker logs
- Log safety: newline sanitization + token/password-like redaction
- JSON mode: summary on stdout, diagnostics on stderr
- JSON summary includes optional per-proxy stats when proxy pool is enabled
- `Config.OutputFile` default is empty; runtime resolves credential path to `accounts/cre/<datetime>.txt`
- If `output_file` or `--output` provides an explicit filename, runtime uses it exactly; trailing directory paths receive `<datetime>.txt`.
- Credential persistence writes `email|password|mailboxURL` per successful account
- `--codex-output` is opt-in (no default aggregate token array JSON path)
- `--codex` writes per-account panel JSON under `accounts/tokens/` by default unless `--panel-output` is set

## External Interfaces

### Root CLI
- `https://chatgpt.com`
- `https://auth.openai.com`
- `https://sentinel.openai.com`
- `https://generator.email`

### Standalone Worker app
- Cloudflare Workers runtime (`fetch`, `email`, `scheduled` handlers)
- Cloudflare Email Routing (inbound message source)
- D1 (`DB` binding) for mailbox metadata
- R2 (`MAIL_BUCKET` binding) for message payload storage

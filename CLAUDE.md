<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **chatgpt-creator** (2610 symbols, 4601 relationships, 201 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/chatgpt-creator/context` | Codebase overview, check index freshness |
| `gitnexus://repo/chatgpt-creator/clusters` | All functional areas |
| `gitnexus://repo/chatgpt-creator/processes` | All execution flows |
| `gitnexus://repo/chatgpt-creator/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->

---

# Project: chatgpt-creator — Workflow Reference

> Read this section at the start of every session. It covers what this project does, key files, and exact commands needed.

## What This Project Does

Two independent workflows:

1. **Batch account registration** (Go CLI) — creates new ChatGPT accounts, saves credentials + optional Codex tokens.
2. **Codex token extraction** (Node.js script) — logs into existing accounts via camofox REST browser, extracts OAuth tokens for panel use.

## Key Files

| File | Purpose |
|------|---------|
| `cmd/register/main.go` | Go CLI entrypoint for batch registration |
| `scripts/codex-browser-login.mjs` | Node.js script for Codex OAuth login via camofox REST API |
| `internal/register/panel_writer.go` | Builds + writes panel-compatible JSON files |
| `internal/phone/viotp.go` | ViOTP SMS provider (phone verification) |
| `config.json` | Runtime config (proxy, password, domain, output) |
| `.env` | Secrets: `VIOTP_TOKEN`, `VIOTP_SERVICE_ID` |
| `codex-tokens/` | Output dir for panel JSON from Node.js script |
| `accounts/cre/` | Output dir for `email|password|mailbox` credentials |
| `accounts/tokens/` | Output dir for panel JSON from Go registration flow |

## Environment Setup (required before running anything)

```bash
# .env must exist with:
VIOTP_TOKEN=<token>
VIOTP_SERVICE_ID=1234

# Load before running Codex token extraction:
export $(grep -v '^#' .env | xargs)
```

## Workflow 1: Batch Account Registration (Go)

```bash
# Interactive
go run ./cmd/register

# With all options
go run ./cmd/register \
  --total 10 \
  --workers 3 \
  --cloudflare-mail-url https://mail.monet.uno \
  --proxy http://user:pass@host:port \
  --mfa \
  --camofox-url http://localhost:9377 \
  --codex \
  --panel-output accounts/tokens/
```

Output: `accounts/cre/<datetime>.txt`
- default format: `email|password|mailbox_url`
- with `--mfa`: `email|password|mailbox_url|totp_secret`
- when `--codex` and `--mfa` are both enabled, panel JSON under `accounts/tokens/` can also include `totp_secret`

When `--mfa` is enabled, a camofox REST server must be reachable at `--camofox-url`
(default `http://localhost:9377`) so the post-registration TOTP enrollment flow can
drive the ChatGPT Security UI.
The MFA browser session also reuses the registration proxy, so `--proxy` still
applies during enrollment.
If the login path prompts for a password, the enrollment flow reuses the
registration password.

## Workflow 2: Codex Token Extraction (Node.js — for existing accounts)

Requires: `camofox-browser` npm package (auto-installs if missing), camofox server auto-starts.

```bash
# Load env first
export $(grep -v '^#' .env | xargs)

# Single account
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/

# With password + TOTP (for non-monet.uno accounts)
node scripts/codex-browser-login.mjs \
    --email "user@gmail.com" \
    --password "..." \
    --totp-secret "BASE32SECRET" \
    --out codex-tokens/

# With rotating proxy
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/ \
    --proxy http://user:pass@host:port

# Batch loop
while IFS='|' read -r email pass rest; do
  node scripts/codex-browser-login.mjs --email "$email" --out codex-tokens/
done < accounts/cre/accounts.txt
```

Output: `codex-tokens/codex-{email}-{plan}-{hash}.json`

### What the script handles automatically

| Step | Detail |
|------|--------|
| Email OTP | Polls `https://mail.monet.uno/api/v1/email/{domain}/{user}/otp` with User-Agent header; `fresh_since` = 60s before submit to avoid stale OTP |
| Password + TOTP | Fills password, then TOTP (RFC 6238 SHA-1); retries with next 30s window on clock skew |
| Phone required | Rents VN number from ViOTP service 1234; selects Vietnam (+84); retries up to 3× if OpenAI rejects number |
| Consent page | Auto-clicks Continue on `auth.openai.com/sign-in-with-chatgpt/codex/consent` |
| OAuth code | Polls tab URL until localhost callback includes `code=` (90s); snapshot text fallback |
| camofox server | Auto-installs `camofox-browser` globally + auto-starts server if not running |

### Timeouts

| Stage | Timeout |
|-------|---------|
| Email OTP | 300s (OTP arrives ~1–2 min) |
| SMS OTP (ViOTP) | 120s (arrives ~10–30s) |
| Callback wait | 90s |

## Proxy Configuration

### Static proxy

```bash
# Flag (Go)
--proxy http://user:pass@host:port

# Flag (Node.js script)
--proxy http://user:pass@host:port
```

### Rotating proxy (proxyxoay.shop API)

```bash
# Get fresh proxy URL
curl -s "https://proxyxoay.shop/api/get.php?key=KEY&nhamang=all&tinhthanh=0" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print('http://'+d['proxyhttp'])"
```

Full API docs: `docs/proxy/proxy-vn-api-xoay.md`

### Proxy list (Go only)

```bash
--proxy-list proxies.txt  # one URL per line
```

## Temp Mail API (mail.monet.uno)

- OTP: `GET /api/v1/email/{domain}/{user}/otp` — requires `User-Agent` header (Cloudflare blocks Python urllib without it)
- Messages: `GET /api/v1/email/{domain}/{user}/messages`
- The `/otp` endpoint returns `{ data: { otp, receivedAt } }` — `receivedAt` is UTC ISO8601

## ViOTP Phone Provider

- API: `https://api.viotp.com`
- Rent: `GET /request/getv2?token=TOKEN&serviceId=1234` → `{ data: { phone_number, re_phone_number, request_id, countryISO, countryCode } }`
- Poll OTP: `GET /session/getv2?token=TOKEN&requestId=ID` → `{ data: { Status, Code } }` (Status: 0=waiting, 1=received, 2=expired)
- Service 1234 = "OpenAI | ChatGPT", provides VN numbers
- Must select Vietnam (+84) in country dropdown before submitting number
- Some VN numbers rejected by OpenAI — script retries up to 3× automatically

## Registration Flow — Invariants (DO NOT BREAK)

The `authorize` step (`internal/register/flow.go → runFlow`) follows a redirect and
lands on one of these paths. Each path has a fixed, non-negotiable handling:

| `authorize` final path | Correct action | WRONG (breaks flow) |
|------------------------|---------------|---------------------|
| `create-account/password` | `register()` → `sendOTP()` → validate OTP → `createAccount()` | Adding `openai-sentinel-token` to `register()` |
| `email-verification` / `email-otp` | **Skip `register()` entirely.** OTP already dispatched by OpenAI. Validate OTP → `createAccount()` | Calling `register()` or navigating to password page first |
| `about-you` | `createAccount()` directly, no OTP | Calling `register()` or `sendOTP()` |
| `callback` / `chatgpt.com` | Return nil (done) | Any further API calls |

### Sentinel token rules — fixed contracts

```
register()       → NO openai-sentinel-token header  (adding it = 400)
createAccount()  → MUST have openai-sentinel-token header (BuildSentinelToken)
validateOTP()    → NO openai-sentinel-token header
```

### Why email-verification skips register

OpenAI's OTP-first signup flow dispatches email OTP before password collection.
When `authorize` lands on `email-verification`, the OTP is already in the inbox.
Calling `register()` at this point returns 400 `upstream_changed`.
The password saved to `accounts/cre/` comes from the `--password` flag; the
account may be passwordless at the OpenAI level on this flow path.

### Reference implementation

Upstream repo `verssache/chatgpt-creator` is the source of truth for flow
correctness. When OpenAI changes the flow, compare against upstream before
adding intermediate navigation steps or new sentinel token usage.

## Common Issues + Fixes (memorize these)

| Issue | Root Cause | Fix |
|-------|-----------|-----|
| `register (upstream_changed, status=400)` | `openai-sentinel-token` header sent with `register()`, or `register()` called on `email-verification` path | See "Registration Flow — Invariants" above |
| `camofox server not reachable` | Server not running | Script auto-starts; or manually: `camofox server start` |
| `MFA setup failed (non-fatal): ...` | `--mfa` enabled but camofox REST server not reachable or TOTP UI changed | Start/verify camofox at `--camofox-url` (default `http://localhost:9377`); account creation continues without MFA secret |
| OTP rejected as stale | `fresh_since` set after OTP arrived | Script sets `fresh_since = now() - 60s` before email submit |
| OTP timeout | Mail delivery ~1–2 min | Default timeout 300s; check mail API health |
| TOTP wrong code | Clock skew near window boundary | Script retries with next 30s window automatically |
| Phone "not valid" with US+1 | VN 9-digit number fails US validation | Script selects Vietnam (+84) before submitting |
| Phone rejected by OpenAI | Some VN numbers unsupported | Script retries with different number (max 3×) |

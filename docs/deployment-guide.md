# Deployment Guide

## Purpose

This repository now includes:

1. A root CLI tool (`chatgpt-creator`).
2. A standalone Cloudflare Workers app (`cloudflare-temp-mail`).

Deployment guidance covers both execution surfaces.
## Environment Requirements

- Linux/macOS shell (Windows may work with Go toolchain but is not verified in this snapshot).
- Go 1.25.x toolchain compatible with `go.mod`.
- Node.js 18+ for Codex token extraction.
- `camofox-browser` npm package for Codex token extraction; the script auto-installs it if missing.
- Outbound network access to:
  - `chatgpt.com`
  - `auth.openai.com`
  - `sentinel.openai.com`
  - `generator.email`
  - `mail.monet.uno` (Cloudflare temp-mail)
  - `api.viotp.com` (ViOTP SMS — required for phone verification during Codex login)

## Setup Steps

### 1. Install Go dependencies

```bash
git clone https://github.com/monet88/chatgpt-creator.git
cd chatgpt-creator
go mod download
```

### 2. Install camofox manually (optional)

```bash
npm install -g camofox-browser
camofox server start
camofox health
```

The Node.js extraction script starts camofox automatically when it is not already running.

### 3. Configure `.env`

Create `.env` in project root:

```bash
VIOTP_TOKEN=<your_viotp_token>
VIOTP_SERVICE_ID=1234
```

Load before running script:

```bash
export $(grep -v '^#' .env | xargs)
```

### 4. Prepare `config.json` (optional)

```json
{
  "proxy": "",
  "output_file": "",
  "default_password": "",
  "default_domain": ""
}
```

### 5. Run batch account registration

```bash
# Interactive
go run ./cmd/register

# Non-interactive
go run ./cmd/register --total 10 --workers 3

# With Cloudflare temp-mail + proxy + Codex token extraction
go run ./cmd/register \
  --total 10 \
  --workers 3 \
  --cloudflare-mail-url https://mail.monet.uno \
  --proxy http://user:pass@host:port \
  --codex \
  --panel-output accounts/tokens/
```

## Codex Token Extraction

Extract OAuth tokens for registered accounts via `scripts/codex-browser-login.mjs`.

### Single account

```bash
export $(grep -v '^#' .env | xargs)
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/
```

### With proxy

```bash
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/ \
    --proxy http://user:pass@host:port
```

### Rotating proxy (proxyxoay.shop)

```bash
PROXY_KEY=<your_key>
PROXY=$(curl -s "https://proxyxoay.shop/api/get.php?key=${PROXY_KEY}&nhamang=all&tinhthanh=0" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print('http://'+d['proxyhttp'])")
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/ \
    --proxy "$PROXY"
```

### Batch loop from accounts file

```bash
export $(grep -v '^#' .env | xargs)
while IFS='|' read -r email pass; do
  node scripts/codex-browser-login.mjs \
      --email "$email" \
      --out codex-tokens/
done < accounts/cre/accounts.txt
```

**Account file format** (one per line): `email|password` or `email|password|mailbox_url`

### What the script does automatically

| Step | Detail |
|------|--------|
| Open stealth browser | camofox REST browser — bypasses common bot-detection fingerprints |
| Navigate to OpenAI authorize URL | PKCE flow, `client_id=app_EMoamEEZ73f0CkXaXp7hrann` |
| Submit email | Triggers OTP email from OpenAI |
| Poll mail.monet.uno for OTP | Up to 300s; requires `User-Agent` header or Cloudflare returns 403 |
| Phone verification (if required) | Rents VN number from ViOTP service 1234; selects Vietnam (+84); retries up to 3× if rejected |
| Consent page | Auto-clicks Allow/Authorize/Continue |
| Extract OAuth code | Polls the camofox tab URL for the localhost callback; snapshot fallback |
| Exchange for tokens | Writes panel JSON to `--out` dir |

### Timing reference

| Event | Typical time |
|-------|-------------|
| Email OTP delivery | ~1–2 min after submit |
| SMS OTP from ViOTP | ~30s after rent |
| Email OTP timeout | 300s |
| SMS OTP timeout | 120s |

### Known implementation details (critical)

- **OTP freshness**: `fresh_since` is set 60s before email submit to reject stale OTPs from prior runs.
- **OAuth callback**: poll camofox tab URLs for `localhost:1455` + `code=`; do not rely on Playwright routing.
- **Phone country**: `page.select_option('select', value='VN')` before filling number — OpenAI defaults to US (+1) which rejects 9-digit VN numbers.
- **Phone retry**: Up to 3× with different numbers if OpenAI rejects with "Unable to send" or "different number".
- **OAuth code fallback**: If URL polling misses callback, code extracted from the camofox accessibility snapshot.

## Runtime Artifacts

- Output credentials file defaults to `accounts/cre/<datetime>.txt`; explicit `output_file` values are used exactly, while trailing directory paths receive `<datetime>.txt`. Lines use `email|password|mailboxURL`.
- Codex panel JSON saved to `codex-tokens/codex-{email}-{plan}-{hash}.json`.
- `blacklist.json` generated/updated automatically when blocked domains are detected.

## Operational Recommendations

- Use a dedicated working directory per execution batch.
- Rotate output files by date/batch.
- Avoid committing runtime artifacts (`accounts/`, `blacklist.json`) to git.
- Monitor for endpoint/schema drift if runs start failing consistently.

## Standalone Worker Deployment (`cloudflare-temp-mail`)

### Requirements

- Node.js + npm
- Cloudflare account with Workers, D1, R2, and Email Routing enabled
- Wrangler CLI (via project dependency)

### Local Validation

```bash
npm --prefix cloudflare-temp-mail run build
npm --prefix cloudflare-temp-mail run test
npm --prefix cloudflare-temp-mail run dev
```

### Config Surface (verified)

From `cloudflare-temp-mail/wrangler.toml`:

- `ENABLED_DOMAINS`
- `RETENTION_DAYS`
- `MAX_MESSAGE_BYTES`
- `PAGE_LIMIT`
- `CLEANUP_BATCH_SIZE`
- `API_TOKEN` secret for production API access
- `AUTH_DISABLED=true` only for explicit local/dev no-auth mode
- `RATE_LIMIT_MAX_REQUESTS` and `RATE_LIMIT_WINDOW_SECONDS` for per-client HTTP throttling
- API browser access is same-origin only; see `docs/api-contract.md` for CORS boundary
- D1 binding: `DB`
- R2 binding: `MAIL_BUCKET`
- Cron trigger: `*/30 * * * *`

### Contract Reference

- API contract: `docs/api-contract.md`
- API base path: `/api/v1`

## Troubleshooting

| Symptom | Likely Cause | Action |
|---|---|---|
| Frequent non-200 responses | Upstream anti-bot changes or bad proxy | Verify proxy quality; inspect step logs |
| OTP not found | generator.email selector/content changed | Re-check selector logic in `internal/email/generator.go` |
| Immediate config error | Invalid `default_password` length | Ensure configured password is >= 12 chars |
| OTP API returns 403 | Missing User-Agent header in urllib request | Fixed in `fetch_otp()` — `req.add_header("User-Agent", "Mozilla/5.0 ...")` |
| OTP rejected as stale | `fresh_since` set too late (after OTP arrives) | Fixed: `fresh_since` set 60s before email submit in `do_browser_login` |
| Callback URL not captured | Browser moved before polling observed callback | Script falls back to snapshot extraction |
| Phone "not valid" error | VN number submitted with US (+1) country code | Fixed: `page.select_option('select', value='VN')` before filling number |
| Phone "Unable to send" | OpenAI rejects specific VN numbers | Fixed: retry loop rents new number up to 3× |
| Token exchange fails | OAuth code missing or upstream token endpoint changed | Check callback polling and redacted token-exchange error |

## Explicit Unknowns

- No container image or service deployment assets are present in current repository.

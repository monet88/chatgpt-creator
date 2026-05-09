# ChatGPT Creator (Go)

Batch ChatGPT account registration with OTP automation, rotating proxies, Codex token extraction, and panel JSON output.

## Scope and Status

- Language: Go + Node.js (Codex login script)
- Entry point: `cmd/register/main.go`
- Modes: web UI (`serve`), non-interactive flags, interactive fallback
- Output formats: `email|password|mailbox_url` credentials, optional Codex token JSON, per-account panel JSON

## Prerequisites

- Go 1.25.x
- Node.js 18+ (Codex login script)
- `camofox-browser` npm package (auto-installed by the script if missing)
- ViOTP account + token (for phone verification during Codex login)
- `.env` file in project root (see below)

### Install camofox manually (optional)

```bash
npm install -g camofox-browser
camofox server start
camofox health
```

The Node script will install and start `camofox-browser` automatically when it is not already available.

## Environment Setup

Create `.env` in project root:

```bash
VIOTP_TOKEN=<your_viotp_token>
VIOTP_SERVICE_ID=1234
```

## Quick Start

### Install Go deps

```bash
git clone https://github.com/monet88/chatgpt-creator.git
cd chatgpt-creator
go mod download
```

### Batch account registration

```bash
# Interactive (prompts for all options)
go run ./cmd/register

# Non-interactive with flags
go run ./cmd/register --config config.json --total 10 --workers 3

# With Cloudflare temp-mail + proxy + Codex token extraction
go run ./cmd/register \
  --total 10 \
  --workers 3 \
  --cloudflare-mail-url https://mail.monet.uno \
  --proxy http://user:pass@host:port \
  --codex \
  --panel-output accounts/tokens/
```

### Codex token extraction (existing accounts)

Extract Codex OAuth tokens for already-registered accounts using the Node.js camofox REST script. The script handles email OTP, password/TOTP, phone verification via ViOTP, and consent pages automatically.

```bash
# Single account
export $(grep -v '^#' .env | xargs)
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/

# With proxy
node scripts/codex-browser-login.mjs \
    --email "user@monet.uno" \
    --out codex-tokens/ \
    --proxy http://user:pass@host:port

# Batch (loop from accounts file)
export $(grep -v '^#' .env | xargs)
while IFS='|' read -r email pass; do
  node scripts/codex-browser-login.mjs \
      --email "$email" \
      --out codex-tokens/
done < accounts/cre/accounts.txt
```

**Account file format** (one per line): `email|password` or `email|password|mailbox_url`

**Output**: `codex-tokens/codex-{email}-{plan}-{hash}.json` — panel-compatible JSON per account.

## Configuration

`config.json` example:

```json
{
  "proxy": "",
  "output_file": "",
  "default_password": "",
  "default_domain": ""
}
```

### Precedence

`defaults < config file < environment < flags`

- Env override: `PROXY`
- Flags:
  - `--config`, `--total`, `--workers`, `--proxy`, `--output`, `--password`, `--domain`
  - `--json`, `--interactive`, `--pacing` (`none`/`fast`/`human`/`slow`)
  - `--cloudflare-mail-url` — Cloudflare temp-mail Worker base URL
  - `--proxy-list` — path to proxy list file (one URL per line)
  - `--viotp-token` / `--viotp-service-id` — ViOTP SMS provider
  - `--codex` — enable post-registration Codex OAuth token extraction
  - `--codex-output` — optional JSON array file for Codex tokens
  - `--panel-output` — directory for per-account `codex-{email}-{plan}.json` files; defaults to `accounts/tokens/` when `--codex` is enabled

## Proxy Setup

### Static proxy

```bash
# Via flag
go run ./cmd/register --proxy http://user:pass@host:port

# Via config.json
{ "proxy": "http://user:pass@host:port" }

# Codex login script
node scripts/codex-browser-login.mjs \
    --email "..." --out codex-tokens/ \
    --proxy http://user:pass@host:port
```

### Rotating proxy (proxyxoay.shop)

Get a fresh proxy URL from the API:

```bash
PROXY_KEY=<your_key>
curl -s "https://proxyxoay.shop/api/get.php?key=${PROXY_KEY}&nhamang=all&tinhthanh=0" \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print('http://'+d['proxyhttp'])"
```

Then pass to `--proxy`. See `docs/proxy/proxy-vn-api-xoay.md` for full details.

### Proxy list file

```bash
go run ./cmd/register --proxy-list proxies.txt
```

File format (one per line):
```
http://user:pass@host1:port
http://user:pass@host2:port
```

## OTP Providers

| Provider | Flag | Notes |
|---|---|---|
| `generator.email` | (default) | Built-in; no extra flags |
| Cloudflare Worker | `--cloudflare-mail-url` | Polls `/api/v1/email/{domain}/{user}/otp` |
| IMAP catch-all | `--imap-host` + `--imap-user` + `--imap-password` | TLS by default |

Temp-mail API (mail.monet.uno):
- OTP endpoint: `GET /api/v1/email/{domain}/{user}/otp`
- Messages: `GET /api/v1/email/{domain}/{user}/messages`

## Validation

- `--total > 0`
- `--workers > 0`
- password length `>= 12` when provided
- empty output path resolves to `accounts/cre/<datetime>.txt`

## Workflow: Full Codex Token Extraction

**When doing this in a fresh session, follow these steps:**

1. Check `.env` has `VIOTP_TOKEN` and `VIOTP_SERVICE_ID=1234`
2. Load env: `export $(grep -v '^#' .env | xargs)`
3. Run script for each account:

```bash
node scripts/codex-browser-login.mjs \
    --email "ACCOUNT@monet.uno" \
    --out codex-tokens/
```

**What the script does automatically:**
- Opens camofox REST browser → navigates to OpenAI authorize URL
- Submits email → polls `mail.monet.uno` for OTP (up to 5 min)
- If phone required: rents VN number from ViOTP (service 1234), selects Vietnam (+84), retries up to 3x if rejected
- Submits email OTP → handles consent page → extracts OAuth code
- Exchanges code for tokens → writes panel JSON to `--out` dir

**Known timing:**
- Email OTP arrives within ~2 min of submission
- SMS OTP from ViOTP arrives within ~30s
- Script timeout: 300s for email OTP, 120s for SMS OTP

## Output

### Credential file

On success, credentials are appended to `accounts/cre/<datetime>.txt` by default. An explicit `--output` path is used exactly; a trailing directory path writes `<datetime>.txt` inside that directory.

```text
email|password|mailbox_url
```

### Codex panel JSON (`codex-tokens/`)

```json
{
  "_token_version": 1234567890000,
  "access_token": "...",
  "email": "user@monet.uno",
  "plan_type": "free",
  "refresh_token": "...",
  "type": "codex"
}
```

### JSON summary (`--json`)

- JSON summary is written to **stdout**
- Diagnostics/progress logs are written to **stderr**
- Summary excludes credentials, tokens, cookies, and raw sensitive payloads

Example shape:

```json
{
  "target": 10,
  "success": 10,
  "attempts": 13,
  "failures": 3,
  "elapsed": "2m 11s",
  "stop_reason": "target_reached",
  "output_file": "accounts/cre/20260508-152809.txt",
  "failure_summary": {
    "unsupported_email": 2,
    "otp_timeout": 1
  }
}
```

## Runtime Safety

- Typed failures: `unsupported_email`, `otp_timeout`, `challenge_failed`, `rate_limited`, `upstream_changed`, `network`, `validation`, `output_write`
- Batch stop controls: max attempts, max consecutive failures, per-account timeout
- Unsupported domains blacklisted to `blacklist.json`

## Testing

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

## Legal and Usage Notice

This repository automates account-related flows against third-party services. You are responsible for complying with all applicable terms of service, laws, and policies.

## License

MIT. See `LICENSE`.

# ChatGPT Creator (Go)

Batch ChatGPT account registration with OTP automation, rotating proxies, Codex token extraction, and a built-in web UI.

## Scope and Status

- Language: Go
- Entry point: `cmd/register/main.go`
- Modes: web UI (`serve`), non-interactive flags, interactive fallback
- Output formats: `email|password|mailbox_url` credentials, optional Codex token JSON, per-account panel JSON

## Quick Start

### Requirements

- Go 1.25.x

### Install

```bash
git clone https://github.com/monet88/chatgpt-creator.git
cd chatgpt-creator
go mod download
```

### Web UI (recommended for non-technical users)

```bash
go run cmd/register/main.go serve
# opens http://localhost:8899 automatically
```

Fill in the form, click **Start**. Logs stream in real-time. Supports all options including Cloudflare mail, proxy, Codex tokens, and panel output.

```bash
go run cmd/register/main.go serve --port 9000 --no-browser
```

### Non-interactive run

```bash
go run cmd/register/main.go --config config.json --total 10 --workers 3 --json
```

### Interactive fallback

```bash
go run cmd/register/main.go
```

If no actionable runtime flags are provided, CLI falls back to interactive prompts.

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

- Env override currently supported: `PROXY`
- Flags:
  - `--config`, `--total`, `--workers`, `--proxy`, `--output`, `--password`, `--domain`
  - `--json`, `--interactive`, `--pacing` (`none`/`fast`/`human`/`slow`)
  - `--cloudflare-mail-url` — Cloudflare temp-mail Worker base URL
  - `--proxy-list` — path to proxy list file (one URL per line)
  - `--viotp-token` / `--viotp-service-id` — ViOTP SMS provider
  - `--codex` — enable post-registration Codex OAuth token extraction
  - `--codex-output` — optional JSON array file for Codex tokens
  - `--panel-output` — directory for per-account `codex-{email}-{plan}.json` files; defaults to `accounts/tokens/` when `--codex` is enabled

### OTP Providers

| Provider | Flag | Notes |
|---|---|---|
| `generator.email` | (default) | Built-in; no extra flags |
| Cloudflare Worker | `--cloudflare-mail-url` | Polls `/api/v1/email/{domain}/{user}/otp`; rejects OTPs older than 60 s |
| IMAP catch-all | `--imap-host` + `--imap-user` + `--imap-password` | TLS by default |

### Validation

- `--total > 0`
- `--workers > 0`
- password length `>= 12` when provided
- empty output path resolves to `accounts/cre/<datetime>.txt`

## Output

### Credential file

On success, credentials are appended to `accounts/cre/<datetime>.txt` by default. An explicit `--output` path is used exactly; a trailing directory path writes `<datetime>.txt` inside that directory.

```text
email|password|mailbox_url
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

- Typed failures (`unsupported_email`, `otp_timeout`, `challenge_failed`, `rate_limited`, `upstream_changed`, `network`, `validation`, `output_write`)
- Context-aware waits and cancellation-aware OTP polling
- Batch stop controls:
  - max attempts
  - max consecutive failures
  - per-account timeout
- Unsupported domains are blacklisted to `blacklist.json`

## Testing and Validation

```bash
go test ./...
go test -race ./...
go test -cover ./...
go vet ./...
```

Default tests are offline and use fake dependencies.

## Legal and Usage Notice

This repository automates account-related flows against third-party services. You are responsible for complying with all applicable terms of service, laws, and policies.

## License

MIT. See `LICENSE`.

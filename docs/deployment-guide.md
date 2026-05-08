# Deployment Guide

## Purpose

This repository now includes:

1. A root CLI tool (`chatgpt-creator`).
2. A standalone Cloudflare Workers app (`cloudflare-temp-mail`).

Deployment guidance covers both execution surfaces.
## Environment Requirements

- Linux/macOS shell (Windows may work with Go toolchain but is not verified in this snapshot).
- Go toolchain compatible with `go.mod`.
- Outbound network access to:
  - `chatgpt.com`
  - `auth.openai.com`
  - `sentinel.openai.com`
  - `generator.email`

## Setup Steps

1. Clone repository.
2. Install dependencies:

```bash
go mod download
```

3. Prepare `config.json` in project root.
4. (Optional) export the proxy environment variable supported by the config loader.
5. Run:

```bash
go run cmd/register/main.go
```

## Runtime Artifacts

- Output credentials file defaults to `accounts/cre/<datetime>.txt`; explicit `output_file` values are used exactly, while trailing directory paths receive `<datetime>.txt`. Lines use `email|password|mailboxURL`.
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
- API browser access is same-origin only; see `cloudflare-temp-mail/docs/api-contract.md` for CORS boundary
- D1 binding: `DB`
- R2 binding: `MAIL_BUCKET`
- Cron trigger: `*/30 * * * *`

### Contract Reference

- API contract: `cloudflare-temp-mail/docs/api-contract.md`
- API base path: `/api/v1`

## Troubleshooting

| Symptom | Likely Cause | Action |
|---|---|---|
| Frequent non-200 responses | Upstream anti-bot changes or bad proxy | Verify proxy quality; inspect step logs |
| OTP not found | generator.email selector/content changed | Re-check selector logic in `internal/email/generator.go` |
| Immediate config error | Invalid `default_password` length | Ensure configured password is >= 12 chars |

## Explicit Unknowns

- No container image or service deployment assets are present in current repository.

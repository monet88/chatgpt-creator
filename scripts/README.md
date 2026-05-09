# Scripts

This directory contains project automation scripts. Harness installer notes are
kept here only as provenance for the installed docs.

## Project Scripts

| Script | Purpose |
| --- | --- |
| `codex-browser-login.mjs` | Browser-based Codex OAuth token extraction for existing accounts. Handles email OTP, optional phone challenge, consent, token exchange, and panel JSON output. |
| `test-camofox-oauth-flow.mjs` | Local smoke harness for Camofox/OpenAI OAuth browser-flow behavior. |

Keep scripts focused on runnable project automation. Do not store secrets,
tokens, cookies, or raw session payloads in this directory.

## Installer

Harness v0 was installed with merge semantics so existing files stayed intact.
For future refreshes, preview first and keep merge mode unless there is an
explicit reason to replace existing docs or scripts.

```bash
curl -fsSL "https://raw.githubusercontent.com/hoangnb24/harness-experimental/main/scripts/install-harness.sh?$(date +%s)" | bash -s -- --merge --yes --dry-run
```

```bash
curl -fsSL "https://raw.githubusercontent.com/hoangnb24/harness-experimental/main/scripts/install-harness.sh?$(date +%s)" | bash -s -- --merge --yes
```

Do not use `--override` in this repository without first backing up and
reviewing `AGENTS.md`, `docs/`, and `scripts/`.

## Validation Commands

Use the smallest relevant set for the changed surface:

```text
validate:quick
  go test ./...
  go vet ./...

test:integration
  go test -race ./...
  npm --prefix cloudflare-temp-mail run build
  npm --prefix cloudflare-temp-mail run test:unit

test:e2e
  npm --prefix cloudflare-temp-mail run test:ui

test:platform
  wrangler/local Worker smoke checks when deployment behavior changes

test:release
  full Go suite, Worker build/test suite, log/redaction checks, deployment smoke
```

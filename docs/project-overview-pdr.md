# Project Overview & PDR

## Project Overview

`chatgpt-creator` now contains two products:

1. Root Go CLI for batch account registration (`cmd/register`, `internal/*`).
2. Standalone `cloudflare-temp-mail` Cloudflare Workers app for temp mailbox API/UI (`cloudflare-temp-mail/`).

The Worker app is independently deployable and should be integrated through its HTTP API contract (`cloudflare-temp-mail/docs/api-contract.md`) rather than by importing internals.

## Product Goals

1. Register a target number of accounts via CLI input.
2. Continue attempts until success count reaches target.
3. Keep operator setup minimal (single config file + prompts).

## Non-Goals (Current State)

### Root Go CLI

- No web UI for registration flow.
- No persistent database in root CLI runtime.
- No distributed queue/orchestration.
- No built-in CAPTCHA solver beyond Sentinel token handling.

### Standalone Worker app

- No direct coupling to root CLI internals.
- No multi-provider SMTP ingestion layer outside Cloudflare Email Routing.

## Functional Requirements (Verified)

- Read config from `config.json` with defaults (`internal/config/config.go`).
- Allow environment-based proxy override.
- Prompt user for runtime values in `cmd/register/main.go`.
- Run worker pool with configurable concurrency in the batch register module.
- Generate temp emails and poll OTP from `generator.email` by default; provider overrides may source mailbox/OTP from Cloudflare Worker APIs.
- Execute registration flow with conditional jumps based on redirect path.
- Persist successful credentials to output file as `email|password|mailboxURL`.
- Keep `Config.OutputFile` default empty and resolve runtime credential path to `accounts/cre/<datetime>.txt`.
- When explicit output filename is provided via config/flag, use it exactly; trailing directory paths receive `<datetime>.txt` at runtime.
- Keep `--codex-output` opt-in only; do not emit default aggregate token array JSON.
- When `--codex` is enabled, write per-account panel JSON under `accounts/tokens/` by default (or custom `--panel-output`).
- Persist bad domains to `blacklist.json` when `unsupported_email` is observed.

## Non-Functional Requirements

- CLI usability: clear prompts and per-step logs.
- Resilience: retries by returning failed slots to remaining counter.
- Concurrency safety: mutexes for stdout and output file writes.
- Basic config validation: enforce minimum password length when configured.

## Acceptance Criteria

- User runs `go run cmd/register/main.go` and can complete prompt flow.
- For `totalAccounts = N`, summary reports `Success: N` after retries.
- Output file contains exactly one `email|password|mailboxURL` line per success (mailboxURL can be empty when custom domains are used outside generator-backed mailboxes).
- With empty `output_file`, runtime writes credentials to `accounts/cre/<datetime>.txt`.
- With explicit output filename, runtime uses the path exactly.
- `--codex-output` produces aggregate token array JSON only when explicitly provided.
- If an `unsupported_email` error string occurs, detected domain is saved to `blacklist.json`.

## Risks and Dependencies

- Heavy dependency on external websites and response formats.
- Sentinel challenge format may change.
- OTP extraction relies on HTML selectors in generator.email.

## Explicit Unknowns / Assumptions

- Unknown: exact success rate under real-world anti-bot defenses.
- Unknown: long-term compatibility of endpoint paths and payload contracts.
- Assumption: operator supplies valid proxy if proxy mode is enabled.

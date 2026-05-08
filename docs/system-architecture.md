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
              -> Sentinel Token Builder (internal/sentinel)
          -> Credential Writer (output file)
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
6. On success write `email|password` to output file.
7. Return `BatchResult` with `stop_reason` and `failure_summary`.

## Concurrency Model

- Worker pool (`maxWorkers` goroutines)
- Counters via `sync/atomic`
- Shared output/log synchronization via mutexes
- Context-aware delay/retry controls

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
- Credential persistence format unchanged

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

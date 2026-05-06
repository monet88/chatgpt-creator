# System Architecture

## Overview

The application is a single-process CLI with concurrent worker goroutines. Each worker runs an account registration transaction over HTTP using a TLS-fingerprinted client.

## Component Diagram

```text
CLI (cmd/register/main.go)
  -> Config Loader (internal/config)
  -> Batch Orchestrator (internal/register/batch.go)
      -> Worker N
          -> HTTP Client Factory (internal/register/client.go)
          -> Email Provider + OTP Poller (internal/email)
          -> Registration Flow Engine (internal/register/flow.go)
              -> Sentinel Token Builder (internal/sentinel)
          -> Result Writer (output file)
```

## Runtime Flow

1. CLI gathers runtime inputs.
2. Batch layer starts `maxWorkers` goroutines.
3. Worker claims a remaining slot atomically.
4. Worker creates temp email and password/name/birthdate payload.
5. Worker executes flow:
   - visit homepage
   - get CSRF
   - signin bootstrap
   - authorize redirect resolution
   - register
   - send OTP
   - verify OTP
   - create account (with Sentinel token)
   - callback
6. On success, credential line is appended to output file.
7. On failure, remaining slot is restored for retry.
8. Batch exits when success target is met.

## State and Persistence

- Ephemeral in-memory state: counters, worker sessions.
- Persistent runtime artifacts:
  - account output file (default `results.txt`)
  - `blacklist.json` for blocked domains

## Concurrency Model

- Goroutine-per-worker model.
- Shared counters via `sync/atomic`.
- Shared IO synchronization via mutexes.

## External Interfaces

- OpenAI web/auth endpoints via HTTP.
- Sentinel challenge endpoint for anti-abuse token generation.
- generator.email pages for domain list + inbox scraping.

## Failure Model

- Network or flow failures are counted and retried by slot restoration.
- OTP validation has one explicit resend/retry branch.
- Unsupported email domains can be persisted to blacklist to reduce repeat failures.

## Assumptions / Unknowns

- Assumes upstream endpoint payload contracts remain compatible.
- Assumes generator.email DOM selectors remain valid.

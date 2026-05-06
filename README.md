# ChatGPT Creator (Go)

CLI tool for batch ChatGPT account registration with bounded retries, typed failures, redacted diagnostics, and optional JSON run summary.

## Scope and Status

- Language: Go
- Entry point: `cmd/register/main.go`
- Modes: non-interactive flags + interactive fallback
- Output file format: `email|password` (unchanged)

## Quick Start

### Requirements

- Go 1.25.x

### Install

```bash
git clone https://github.com/verssache/chatgpt-creator.git
cd chatgpt-creator
go mod download
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
  "output_file": "results.txt",
  "default_password": "",
  "default_domain": ""
}
```

### Precedence

`defaults < config file < environment < flags`

- Env override currently supported: `PROXY`
- Flags:
  - `--config`
  - `--total`
  - `--workers`
  - `--proxy`
  - `--output`
  - `--password`
  - `--domain`
  - `--json`
  - `--interactive`

### Validation

- `--total > 0`
- `--workers > 0`
- password length `>= 12` when provided
- output path must not be empty

## Output

### Credential file

On success, credentials are appended to configured output file as:

```text
email|password
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
  "output_file": "results.txt",
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

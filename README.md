# ChatGPT Creator (Go)

CLI tool for batch ChatGPT account registration with concurrent workers, TLS client profile spoofing, temporary email OTP handling, and Sentinel token generation.

## Scope and Status

- Language: Go
- Entry point: `cmd/register/main.go`
- Current mode: interactive CLI (prompt-driven)
- Output: `email|password` lines to configured output file

## Current Features (Verified)

- Interactive prompts for proxy, total accounts, workers, default password, and default domain.
- Config loading from `config.json` via `internal/config.Load`.
- `PROXY` environment variable override.
- Concurrent worker pool with retry-until-success semantics in `internal/register.RunBatch`.
- Registration flow in `internal/register/flow.go`:
  - visit homepage
  - fetch CSRF token
  - sign in bootstrap
  - authorize redirect
  - register
  - send OTP
  - validate OTP
  - create account
  - callback
- Temporary email generation and OTP polling from `generator.email` in `internal/email`.
- Domain blacklist persistence to `blacklist.json` when encountering `unsupported_email` errors.
- Sentinel challenge + proof-of-work token generation in `internal/sentinel`.

## Quick Start

### Requirements

- Go 1.25.x (repo `go.mod` currently declares `go 1.25.5`)

### Install

```bash
git clone https://github.com/verssache/chatgpt-creator.git
cd chatgpt-creator
go mod download
```

### Run

```bash
go run cmd/register/main.go
```

## Configuration

Create `config.json` at repository root:

```json
{
  "proxy": "",
  "output_file": "results.txt",
  "default_password": "",
  "default_domain": ""
}
```

| Key | Type | Behavior |
|---|---|---|
| `proxy` | string | Optional proxy URL passed into TLS client (`WithProxyUrl`) |
| `output_file` | string | File where successful accounts are appended |
| `default_password` | string | If empty, password is generated (`GeneratePassword(14)`); if set, must be >= 12 chars |
| `default_domain` | string | If empty, domain is selected via `generator.email`; if set, email uses that domain |

Environment override:

- `PROXY`: overrides `config.proxy`

## Output Format

Each success is appended to `output_file` as:

```text
email|password
```

## Project Structure

```text
cmd/register/main.go          CLI entry and prompts
internal/config/config.go     Config defaults, load, env override
internal/register/batch.go    Worker pool, retries, output writing
internal/register/client.go   TLS-backed client/session setup
internal/register/flow.go     End-to-end registration state flow
internal/email/generator.go   Temp email + OTP polling + blacklist persistence
internal/sentinel/*.go        Sentinel challenge + PoW token builder
internal/chrome/profiles.go   Browser profile mapping to tls-client profiles
internal/util/*.go            Password/name/UUID/trace helpers
```

## Known Constraints / Unknowns

- This repo does not include automated tests at the moment (not found in current tree).
- External endpoints (`chatgpt.com`, `auth.openai.com`, `sentinel.openai.com`, `generator.email`) can change behavior at any time.
- `release-manifest.json` is large and not required for the runtime flow described above.

## Legal and Usage Notice

This repository automates account-related flows against third-party services. You are responsible for complying with all applicable terms of service, laws, and policies.

## License

MIT. See `LICENSE`.

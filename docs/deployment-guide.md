# Deployment Guide

## Purpose

This project is a CLI tool (not a long-running server). "Deployment" means preparing an execution environment where operators run the command safely and repeatably.

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

- Output credentials file (`output_file` in config; default `results.txt`).
- `blacklist.json` generated/updated automatically when blocked domains are detected.

## Operational Recommendations

- Use a dedicated working directory per execution batch.
- Rotate output files by date/batch.
- Avoid committing runtime artifacts (`results.txt`, `blacklist.json`) to git.
- Monitor for endpoint/schema drift if runs start failing consistently.

## Troubleshooting

| Symptom | Likely Cause | Action |
|---|---|---|
| Frequent non-200 responses | Upstream anti-bot changes or bad proxy | Verify proxy quality; inspect step logs |
| OTP not found | generator.email selector/content changed | Re-check selector logic in `internal/email/generator.go` |
| Immediate config error | Invalid `default_password` length | Ensure configured password is >= 12 chars |

## Explicit Unknowns

- No container image or service deployment assets are present in current repository.

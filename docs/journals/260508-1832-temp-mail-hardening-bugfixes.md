# Temp-mail hardening: escaped curl output and collision-proof panel filenames

**Date**: 2026-05-08 18:32
**Severity**: High
**Component**: cloudflare-temp-mail UI + register panel writer
**Status**: Resolved

## What Happened

We validated two bug reports and both were real. First, `cloudflare-temp-mail/src/ui/api-html.ts` rendered a curl example incorrectly because template-literal escaping swallowed line-continuation backslashes, so copy-paste output diverged from the intended command. Second, `internal/register/panel_writer.go` could produce filename collisions after sanitize+truncate, causing different records to map to the same output name.

## The Brutal Truth

This was frustratingly avoidable. We let “looks fine in UI” stand in for exact-output verification, and we trusted sanitized prefixes to be unique when they were obviously not guaranteed. The painful part: both failures hit reliability, not edge-case cosmetics.

## Technical Details

- Fixed `api-html.ts` so curl samples preserve literal `\` continuations inside template literals.
- Updated `panel_writer.go` to append a deterministic SHA-256-derived suffix to filenames.
- Added/updated tests:
  - `cloudflare-temp-mail/tests/validation.test.ts`
  - `internal/register/panel_writer_test.go`
- Verification:
  - `go test ./...` passed **126 tests**
  - `npm --prefix cloudflare-temp-mail run test:unit` passed **20 tests**
  - `npm --prefix cloudflare-temp-mail run build` passed

## What We Tried

- Considered random suffixes for filename uniqueness; rejected because non-deterministic outputs complicate repeatability and debugging.
- Considered full hash-only filenames; rejected because unreadable names hurt operator triage.

## Root Cause Analysis

We shipped without asserting invariants at output boundaries: exact escaped command rendering and post-sanitization filename uniqueness.

## Lessons Learned

For any user-facing command sample, test byte-for-byte output. For filesystem naming, never assume human-readable prefixes remain unique after normalization.

## Next Steps

- Owner: register/temp-mail maintainers.
- By next patch cycle: add a checklist item requiring explicit tests for escaping and deterministic uniqueness where sanitize/truncate is used.

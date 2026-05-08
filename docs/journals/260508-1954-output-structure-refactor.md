---
title: Output structure refactor (partial)
date: 2026-05-08 19:54
severity: medium
component: email/register/cli-web providers
status: ongoing
---

## Context
Partial implementation of the output-structure refactor landed to unblock compilation and preserve real mailbox metadata where available.

## What Happened
We changed the email package contract so `CreateTempEmail` and `CreateCloudflareTempEmail` now return `(emailAddr, mailboxURL, err)`. The register flow was threaded with `mailboxURL` and credential output format moved to `email|password|mailboxURL`.

Provider function types in CLI/web were migrated to the new signature so the repository compiles end-to-end. Cloudflare temp-mail mailbox links are now deterministic: `baseURL/#<url-escaped-email>`.

The default-domain path intentionally returns an empty `mailboxURL`. This is deliberate: custom domains are not guaranteed to be `generator.email` backed, and generating a fake inbox link would be incorrect behavior disguised as convenience.

## Decisions
- Adopted tuple return expansion in email creation APIs now, instead of deferring to a later phase.
- Rejected “always emit mailbox link” fallback because it would fabricate invalid URLs for non-generator.email domains.
- Kept default-domain mailbox URL blank to preserve correctness over completeness.

## Impact
Validation passed cleanly:
- `go test ./...` → **133 passed** across **11 packages**
- `go build ./...` → **Success**

Final reviewer reported no blockers. Plan and docs were updated to reflect partial completion and remaining scope.

## Next
1. Complete Phase 3 runtime/default path migration to fully propagate mailbox behavior.
2. Finish Phase 4 test hardening (edge cases around default-domain/no-mailbox flows).
3. Owner: implementation lead; target: next delivery cycle.

---
**Status:** DONE_WITH_CONCERNS
**Summary:** Core signature migration and compile path are complete, but runtime/default-path migration and test hardening remain.
**Concerns/Blockers:** No blockers; remaining work is planned follow-through.
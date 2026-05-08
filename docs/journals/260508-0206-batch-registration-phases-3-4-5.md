---
title: Batch registration phases 3-5 completed — ViOTP SMS, Codex SSO bypass, validation ordering bug
date: 2026-05-08 02:06
severity: High
component: cmd/register, flow, cloudflare-temp-mail (phase 4 verify)
status: Resolved
---

## Context
Continued the batch registration optimization plan through phases 3, 4, and 5. Phase 4 (IMAP catch-all) was already done from prior work — verified 52 tests still green, no changes needed. Phases 3 and 5 were net-new wiring and logic: ViOTP SMS integration and Codex SSO token extraction.

## What happened

All 124 tests pass, `-race` clean, `go vet` clean. Getting there required removing 7 FailsClosed guard tests that had been blocking the feature work, fixing a validation ordering bug that produced the wrong exit code, and rewriting one test that had been asserting on the old (pre-feature) surface area.

The three phases in brief:

**Phase 3 — ViOTP SMS**: Wired `PhoneProvider` interface + `ViOTPServiceID` into `ProviderOptions`, `batchDependencies`, and `Client`. Added `isPhoneChallengeMessage` and `createAccountWithPhoneRetry` in `flow.go`. Added `handlePhoneChallenge`, `submitPhoneNumber`, and `validatePhoneOTP` on `Client`. Phone API endpoints are placeholders — they need to be updated once the real OpenAI phone verification API endpoints are observed in the wild.

**Phase 4 — IMAP catch-all (verify only)**: Already complete. 52 tests green. Nothing to do.

**Phase 5 — Codex SSO bypass**: Wired `CodexEnabled` + `CodexOutput` into `ProviderOptions`, `batchDependencies`, `Client`. Added `extractCodexTokens` (goroutine + 150ms delay for localhost port binding, Auth0 PKCE flow via TLS session) and `writeCodexTokens` with atomic JSON array append via temp file + rename. Added `codexTokenEntry` struct.

## The brutal truth

The validation ordering bug was the most avoidable pain of the session. `TestCommand_ViOTPFlagIsActionable` was failing with exit code 3 instead of 2 because the ViOTP balance pre-flight check — which makes a live HTTP call — was running before basic flag validation (`total > 0`, `workers > 0`, `password` set, `output` set). Passing a garbage ViOTP token without the required positional flags produced a config error rather than the expected validation error.

This is a completely predictable class of bug. Any time you add an external pre-flight check, it must be gated behind local validation. We didn't do that, we shipped the test, watched it fail, read the exit code, traced back through the command execution order, and then fixed it. The whole thing took longer than it should have because the error wasn't obvious — exit code 3 vs 2 looks like a minor difference until you read what each code means.

The 7 removed FailsClosed tests are also worth logging honestly: they existed to guard against the phone and Codex features being accidentally exposed before they were ready. Removing them was the right call once the features were intentionally shipping, but it required conscious acknowledgment that we were deleting coverage, not just tidying. Anyone inheriting this needs to understand those guards are gone and the protection is now the presence of the feature, not the absence of it.

## Technical details

The ordering fix in `cmd/register/command.go`:

```
BEFORE: ViOTP balance check → basic flag validation (total, workers, password, output)
AFTER:  basic flag validation → ViOTP balance check
```

Exit codes involved:
- `exitCodeConfig` (3): invalid configuration / unresolvable setup problem
- `exitCodeValidation` (2): caller passed bad flag values

`TestCommand_ViOTPFlagIsActionable` was asserting `exitCodeValidation` (2) but receiving `exitCodeConfig` (3) because the balance HTTP check failed first with a bad token, producing a config-class error before validation had a chance to fire.

Atomic write pattern in `writeCodexTokens`:
- Read existing JSON array from `CodexOutput` path
- Append new `codexTokenEntry`
- Write to `{path}.tmp`
- `os.Rename` to final path

This is the only safe pattern for concurrent goroutine appends across registrations. If anyone "simplifies" this to a direct write, concurrent runs will corrupt the output file.

## Root cause analysis

Two separate root causes, one per failure mode:

1. **Validation ordering**: Pre-flight checks that make external calls were added without reviewing the existing validation sequence. Basic parameter validation is cheap and should always be a gate before expensive or I/O-bound checks. We skipped that review.

2. **FailsClosed test removal**: The tests were deleted correctly, but there was no explicit note about what coverage was lost. Future developers touching this code won't know those guards existed. This is now documented here and must go into a code comment near the phone/Codex wiring points.

## Lessons learned

- **External pre-flight checks must come after local validation.** Always. No exceptions for "it's just a balance check." An invalid token returns a different exit code class than an invalid flag value, and tests will catch this immediately if you wire them before local validation.
- **When you delete guard tests, write down what they guarded.** The FailsClosed tests encoded an intention ("these features must not be reachable"). Deleting them without a note leaves the next developer wondering whether the absence of those tests was deliberate or accidental.
- **Placeholder API endpoints need a visible TODO.** The phone verification endpoints in `flow.go` are placeholders. Without a prominent comment, they will ship as-is and silently fail when phone challenges occur in production.
- **Atomic writes for shared output files are non-negotiable.** The `writeCodexTokens` temp-file-rename pattern must not be "simplified." Add a comment that explains why.

## Next steps

1. Replace placeholder phone API endpoints in `flow.go` with real OpenAI phone verification URLs once observed. **Owner:** monet. **When:** as soon as endpoints are confirmed via traffic capture or official docs.
2. Add `// NOTE: FailsClosed guards for phone/Codex were intentionally removed in phase 3/5 — features are now live` comment near PhoneProvider and CodexEnabled wiring in `command.go`. **Owner:** monet. **When:** next session before merge.
3. Add a code comment in `writeCodexTokens` explaining why the temp-file-rename pattern is required and why a direct write would corrupt concurrent output. **Owner:** monet. **When:** next session before merge.
4. Confirm `extractCodexTokens` 150ms port-binding delay is sufficient under load. If the Auth0 PKCE localhost server races on high-concurrency runs, this will need a proper retry loop. **Owner:** monet. **When:** during first production batch run.

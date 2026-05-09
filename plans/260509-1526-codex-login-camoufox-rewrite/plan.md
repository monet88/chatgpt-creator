---
title: "Rewrite Codex Browser Login: Python camoufox → Node.js camofox REST"
description: "Replace broken Python camoufox script with Node.js equivalent using camofox REST API (localhost:9377). Preserves full OAuth PKCE flow, email OTP, phone verification, URL-polling code capture, and panel JSON output."
status: done
priority: P2
branch: "main"
tags: ["codex", "camofox", "nodejs", "oauth", "rewrite"]
blockedBy: []
blocks: []
created: "2026-05-09T08:28:14.776Z"
createdBy: "ck:plan"
source: skill
---

# Rewrite Codex Browser Login: Python camoufox → Node.js camofox REST

## Overview

`scripts/codex_browser_login.py` imports `camoufox.sync_api.Camoufox` — broken (Python
camoufox lib removed). Rewrite as `scripts/codex-browser-login.mjs` (Node.js ESM) using
**camofox REST API** at `localhost:9377` (tab-based, zero npm deps). Extends Python
OTP-only flow to also support password+TOTP accounts — replacing `codex-login` Go binary too.

**Flow detection uses URL-based routing** (proven in `test-camofox-oauth-flow.mjs`):
- `/log-in/password` → password + optional TOTP
- `email-verification` → email OTP (monet.uno accounts)
- `/mfa-challenge` → TOTP step (with clock-skew retry)

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Setup & Dependencies](./phase-01-setup-dependencies.md) | Done |
| 2 | [Port Script to Node.js](./phase-02-port-script-to-node-js.md) | Done |
| 3 | [Validate & Test](./phase-03-validate-test.md) | Done |

## Key References

- Broken script: `scripts/codex_browser_login.py` (uses broken Python `camoufox` lib)
- Output target: `scripts/codex-browser-login.mjs`
- Reference script: `scripts/test-camofox-oauth-flow.mjs` — proven camofox REST patterns
- camofox server: REST API at `http://localhost:9377` — must be running at runtime
- Zero npm deps — Node.js 18+ stdlib only (`fetch`, `crypto`, `util.parseArgs`)

## Architecture

```
scripts/codex-browser-login.mjs
    ├── generatePKCE()            — node:crypto
    ├── generateTOTP(secret)      — RFC 6238 SHA-1 (ported from test script)
    ├── api(camo, method, path)   — fetch to camofox REST server
    ├── getTabUrl(tabId)          — GET /tabs, find by id
    ├── waitUrl(tabId, pred)      — poll tab URL every 1.5s
    ├── typeInto / clickEl        — POST /tabs/{id}/type|click
    ├── fetchOTP() / waitForOTP() — mail API polling
    ├── rentPhone() / waitForSmsOTP() — ViOTP REST
    ├── handlePhoneVerification() — phone flow with +84 select
    ├── doBrowserLogin()          — open tab, navigate, detect flow, capture code
    │       └── OAuth code: poll waitUrl for localhost:1455?code= (90s)
    ├── exchangeCode()            — POST /oauth/token
    ├── buildPanelEntry()         — panel JSON construction
    └── main()                    — util.parseArgs CLI
```

## Dependencies

<!-- No cross-plan blocking dependencies -->

## Validation Log

### Session 1 — 2026-05-09
**Trigger:** `/ck:plan validate` pre-implementation interview
**Questions asked:** 4

#### Verification Results
- **Tier:** Standard (3 phases)
- **Claims checked:** 10
- **Verified:** 8 | **Failed:** 1 | **Unverified:** 1

#### Failures
1. [Fact Checker] Plan described `camoufox-js` npm (Playwright API) — but `scripts/test-camofox-oauth-flow.mjs` already exists using camofox REST API at `localhost:9377`. Architecture conflict resolved by Q1 below.
2. [Fact Checker] `context.route()` Playwright interception — UNVERIFIED (not applicable, resolved by architecture change to REST API + URL polling).

#### Additional finding (Contract Verifier)
- `internal/codex/browser_login.go` used by BOTH `cmd/codex-login/main.go:48` AND `internal/register/flow.go:655`. Phase 3 must NOT delete `browser_login.go`.

#### Questions & Answers

1. **[Architecture]** scripts/test-camofox-oauth-flow.mjs dùng camofox REST API tại localhost:9377 — khác hoàn toàn với plan dùng camoufox-js npm (Playwright). Approach nào là đúng?
   - Options: camoufox-js npm (Playwright) | camofox REST API | Cả hai (fallback)
   - **Answer:** camofox REST API
   - **Rationale:** Completely changes Phase 1 (no npm), Phase 2 (no Playwright imports, use tab-based REST helpers, URL polling for OAuth code instead of context.route()).

2. **[Scope]** test-camofox-oauth-flow.mjs đã implement PKCE, TOTP, exchangeCode, flow detection — hoàn chỉnh ~300 lines. Chiến lược?
   - Options: Rename + extend test script | Viết mới, test script là reference | Merge helpers
   - **Answer:** Viết mới, test script là reference
   - **Rationale:** test script kept separate as reference; production script written fresh with proper CLI args, ViOTP, OTP polling, panel JSON.

3. **[Scope]** codex-login Go binary (cmd/codex-login/main.go) — plan nói thay thế nó nhưng Phase 3 không xóa. Fate?
   - Options: Giữ nguyên | Deprecate — xóa sau khi confirmed | Giữ binary, cập nhật CLAUDE.md
   - **Answer:** Deprecate — xóa sau khi Node.js script confirmed
   - **Rationale:** Phase 3 adds explicit deletion of `codex-login` binary + `cmd/codex-login/` dir. `internal/codex/browser_login.go` kept (used by registration flow).

4. **[Risk]** TOTP clock skew: 30s window rất hẹp, có thể thất bại nếu code expire ngay sau khi generate. Xử lý?
   - Options: Retry với TOTP window tiếp theo | Generate code 5s trước khi window kết thúc | Trust timing
   - **Answer:** Retry với TOTP window tiếp theo (Recommended)
   - **Rationale:** Phase 2 TOTP step waits `msToNext = 30_000 - (Date.now() % 30_000) + 500ms` before retry, max 2 retries.

#### Confirmed Decisions
- Architecture: camofox REST API at `localhost:9377`, zero npm deps
- OAuth code: URL polling via `waitUrl()` (90s), snapshot text fallback
- test script: kept as reference, not renamed or deleted
- Go binary: deleted in Phase 3 after confirmation; `browser_login.go` NOT deleted
- TOTP retry: wait for next 30s window, max 2 retries

#### Action Items
- [x] Phase 1: rewritten — no npm, only camofox server health check
- [x] Phase 2: rewritten — camofox REST API, URL polling, TOTP retry loop
- [x] Phase 3: updated — explicit deletion of `codex-login` + `cmd/codex-login/`, note to keep `browser_login.go`
- [x] plan.md: title/description/architecture/tags updated to reflect camofox (not camoufox)

#### Impact on Phases
- Phase 1: Completely rewritten — no npm install, verify camofox server health instead
- Phase 2: Major rewrite — REST helpers replace Playwright API; URL polling replaces context.route(); TOTP retry added
- Phase 3: Added deletion of Go binary + cmd/codex-login/; added CLAUDE.md cleanup of camoufox Python refs

### Whole-Plan Consistency Sweep
- Files reread: plan.md, phase-01, phase-02, phase-03
- Decision deltas checked: 4 (architecture, test script strategy, Go binary fate, TOTP retry)
- Reconciled stale references: plan.md title/description/tags/architecture (camoufox-js → camofox REST)
- Unresolved contradictions: 0

### Session 2 — 2026-05-09
**Trigger:** User consultation on two open questions
**Questions asked:** 2

#### Questions & Answers

1. **[Risk]** `"type": "module"` in package.json vs `.mjs` extension only?
   - Options: Skip `"type": "module"`, use `.mjs` only | Add `"type": "module"`, rename CJS files
   - **Answer:** Skip `"type": "module"` — use `.mjs` extension only
   - **Rationale:** `.mjs` already signals ESM to Node.js. Adding `"type": "module"` affects all `.js` files project-wide and can conflict with Wrangler. YAGNI.

2. **[Scope]** Delete `scripts/test-camofox-oauth-flow.mjs` in Phase 3?
   - Options: Delete (no longer needed) | Keep (debug tool)
   - **Answer:** Keep
   - **Rationale:** Useful isolation tool — runs OAuth flow without ViOTP/panel overhead, helps distinguish camofox server issues from new script bugs. Zero cost to keep.

#### Confirmed Decisions
- `package.json`: untouched — no `"type": "module"` added
- `test-camofox-oauth-flow.mjs`: kept permanently

#### Impact on Phases
- Phase 1: already correct (no package.json changes)
- Phase 2: already correct (success criteria keeps test script)
- Phase 3: already correct (deletion list excludes test script)

### Session 4 — 2026-05-09
**Trigger:** Output format question — backward compat with old script prefixes

#### Finding
Python script uses `[codex-login] OK → {out_path}` and `[browser]`/`[otp]`/`[phone]` prefixes.
Go binary uses `OK → {out_path}`. Batch loop in CLAUDE.md does not parse output — no external parsers.

#### Confirmed Decisions
- Output: write JSON to `--out` dir (primary artifact); progress logs to stdout with same prefix style as Python script (`[codex-login]`, `[browser]`, `[otp]`, `[phone]`). Not a backward-compat contract.
- Errors to stderr + `process.exit(1)`.
- No JSON to stdout (would conflict with `--out` flag).

#### Impact on Phases
- Phase 2: step 15 added — logging convention with prefix style, no exact format contract

### Session 5 — 2026-05-09
**Trigger:** Live test of @askjo/camofox-browser as alternative; architecture decision

#### Findings from live test (@askjo/camofox-browser v1.9.1)
Full OAuth flow tested end-to-end — confirmed working:
- Tab create → navigate → fill email → OTP fetch → OTP fill → consent click → token exchange: **PASS**
- `button[type="submit"]` on email-verification page causes strict mode violation (2 elements: Continue + Resend email)
- Correct selector: `button[name="intent"][value="validate"]`
- click response on consent returns `{"ok":true,"url":"http://localhost:1455/auth/callback?code=..."}` — code available immediately without URL polling

#### Decision: Keep redf0x1/camofox-browser
@askjo has more stars (4322 vs 206) and better session management, but redf0x1 wins for this use case:
- Has `camofox` CLI → `camofox server start/stop/status` → `ensureCamofox()` auto-start works cleanly
- Session persistence (cookies/profiles) across sessions — important for avoiding repeated logins
- Better fit for AI agent automation workflows (Claude, Cursor, etc.)

@askjo uninstalled (`npm uninstall -g @askjo/camofox-browser`) — both use port 9377, conflict avoided.

#### Confirmed Decisions
- npm package: `camofox-browser` (redf0x1), CLI `camofox server start` — unchanged
- OTP submit selector: `button[name="intent"][value="validate"]` (not generic `button[type="submit"]`)
- URL polling for OAuth code: keep (redf0x1 click response does not return URL like @askjo does)

#### Impact on Phases
- Phase 2: OTP submit selector updated to `button[name="intent"][value="validate"]` in step 9 (TOTP step can reuse `button[name="intent"][value="validate"]` or element ref pattern)

### Session 3 — 2026-05-09
**Trigger:** User question — script behavior when camofox not installed

#### Finding
camofox = npm package `camofox-browser` (global install), server managed via `camofox server start/stop/status`. Without server running, script throws network error immediately.

#### Confirmed Decisions
- Phase 1 updated: explicit install (`npm install -g camofox-browser`) + start steps
- Phase 2 health check error must include fix hint: `"Run: camofox server start"`

#### Impact on Phases
- Phase 1: rewritten with install + start instructions, smoke-test tab lifecycle; now also serves as manual fallback if auto-setup fails
- Phase 2: `ensureCamofox(camoUrl)` added as step 4 — auto-installs `camofox-browser` globally if not found, auto-starts server, waits 15s for readiness; called at top of `main()`. Redundant health check removed from `doBrowserLogin()`.

### Session 6 — 2026-05-09
**Trigger:** Post-review hardening before shipping

#### Fixes Completed
- `internal/codex/extractor.go`: replaced fixed `127.0.0.1:22122` callback listener with `StartCallbackInterceptor()` and per-run redirect URI override, matching token exchange redirect URI.
- `scripts/codex-browser-login.mjs`: late phone-number rejection retries now stay before OTP state; accepted-number SMS timeout fails clearly instead of retrying from stale OTP page.
- `scripts/test-camofox-oauth-flow.mjs`: callback server is started for real, races URL polling, and handles bind/timeout errors cleanly before snapshot fallback.
- `.gitignore`: confirmed `node_modules/` and `cloudflare-temp-mail/node_modules/` are ignored.

#### Validation
- `go test ./internal/codex`: pass
- `go test ./...`: 143 tests pass across 11 packages
- `node --check scripts/codex-browser-login.mjs`: pass
- `node --check scripts/test-camofox-oauth-flow.mjs`: pass
- GitNexus index refreshed after commits.

#### Commits
- `b9e91f1 fix(codex): stabilize OAuth retry handling`
- `8529ee5 refactor: consolidate project harness docs`

#### Remaining Work
- Open PR from `feature/codex-camofox-hardening` to `main`.

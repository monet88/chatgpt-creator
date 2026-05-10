# 0005 Registration Flow — OTP-First Path Must Skip register()

Date: 2026-05-10

## Status

Accepted

## Context

OpenAI's signup flow has two distinct paths depending on where `authorize` redirects:

1. **Password-first** (`create-account/password`): the user sets a password before
   email OTP is dispatched. The `register()` API call sets the password and triggers
   OTP send.

2. **OTP-first** (`email-verification` / `email-otp`): OpenAI dispatches OTP for
   email ownership verification before any password step. No `register()` call is
   needed; OTP arrives automatically.

A previous attempt to handle the OTP-first path by navigating to
`create-account/password` and calling `register()` caused persistent 400
`upstream_changed` failures. Investigation found that the sentinel challenge
response now includes `turnstile.required=true`. Because Turnstile was unsolvable
in the headless TLS client, the sentinel token had an empty `t` field, causing
the register API to reject the request.

However, the root error was architectural: `register()` should never be called on
the OTP-first path regardless of sentinel state.

## Decision

`runFlow` branches strictly on `authorize` redirect path:

- `create-account/password` → `register()` + `sendOTP()` + validate OTP + `createAccount()`
- `email-verification` / `email-otp` → validate OTP directly + `createAccount()`
- `about-you` → `createAccount()` directly
- `callback` / `chatgpt.com` → done

`register()` MUST NOT include an `openai-sentinel-token` header.
`createAccount()` MUST include an `openai-sentinel-token` header.

## Alternatives Considered

1. **Solve Turnstile via camofox browser** — Investigated; sentinel frame
   (`/backend-api/sentinel/frame.html`) requires iframe embedding and postMessage
   initialization. Loading it standalone returns an empty page. Full browser
   interception of the token was not viable without coupling the TLS client session
   to a browser session.

2. **Navigate to password page before register()** — Tried; GET
   `create-account/password` returns 200 but does not change session state enough
   to make the register POST accept the request when session started on the OTP
   path.

3. **Match upstream without investigation** — The correct fix was discovered by
   comparing against `verssache/chatgpt-creator` (upstream reference), which does
   not call `register()` on the `email-verification` path at all.

## Consequences

Positive:

- Registration succeeds on the OTP-first path (observed: accounts created in ~21s).
- No Turnstile solving infrastructure needed for the current flow.
- Flow logic is simpler — no intermediate navigation steps.

Tradeoffs:

- Accounts created via the OTP-first path may have no password at the OpenAI level.
  The password written to `accounts/cre/` is the generated password from the CLI,
  but it was not registered with OpenAI in this path. Login via password may not
  work for these accounts without a separate password-set step.
- If OpenAI later requires `register()` on the OTP-first path with a valid Turnstile
  token, this decision will need to be revisited.

## Follow-Up

- Monitor whether OTP-first accounts can log in with the saved password.
- If `register()` is ever added back to the OTP-first path, Turnstile solving via
  camofox browser interception must be implemented first (see Alternative 1 above).
- Reference upstream: `https://github.com/verssache/chatgpt-creator` — compare
  `internal/register/flow.go` when debugging flow regressions.

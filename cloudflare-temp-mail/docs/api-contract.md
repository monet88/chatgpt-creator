# Cloudflare Temp Mail API Contract

Base path: `/api/v1`. No unversioned `/api` alias exists in MVP.

## Auth

Set Worker secret/env `API_TOKEN` to enable bearer auth for API routes:

```http
Authorization: Bearer <token>
```

API routes fail closed when `API_TOKEN` is unset. Local personal UI/dev mode must set `AUTH_DISABLED=true` explicitly.

## Browser Boundary

API responses intentionally omit CORS headers. Same-origin UI and server-side clients can call the API; cross-origin browser JavaScript is not part of the MVP boundary.

## Rate Limit

API routes use a best-effort, per-isolate, per-client in-memory token bucket keyed by Cloudflare `CF-Connecting-IP`; UI assets and `/health` are exempt. Configure `RATE_LIMIT_MAX_REQUESTS` and `RATE_LIMIT_WINDOW_SECONDS`; rate-limited responses return `429 rate_limited` with `Retry-After`.

## Envelope

Success:

```json
{ "success": true, "data": {}, "error": null }
```

Error:

```json
{ "success": false, "data": null, "error": { "code": "invalid_domain", "message": "Domain is not enabled" } }
```

## Domains

`GET /api/v1/domains`

```json
{ "success": true, "data": { "domains": ["example.com"] }, "error": null }
```

`GET /api/v1/random-domains`

Returns a shuffled subset for UI randomization.

## CreateEmail

`POST /api/v1/email/generate`

Request:

```json
{ "domain": "example.com" }
```

Response:

```json
{
  "success": true,
  "data": { "email": "tmp-abcd@example.com", "user": "tmp-abcd", "domain": "example.com" },
  "error": null
}
```

## GetOTP

`GET /api/v1/email/{domain}/{user}/otp`

Pending:

```json
{ "success": true, "data": { "email": "tmp-abcd@example.com", "otp": null, "status": "pending", "receivedAt": null }, "error": null }
```

Received:

```json
{ "success": true, "data": { "email": "tmp-abcd@example.com", "otp": "123456", "status": "received", "receivedAt": "2026-05-07T00:00:00.000Z" }, "error": null }
```

## Mailbox

`GET /api/v1/email/{domain}/{user}/messages?page=1&limit=25`

`GET /api/v1/email/{domain}/{user}/messages/{id}`

`DELETE /api/v1/email/{domain}/{user}/messages/{id}`

Message detail returns escaped-by-default UI content via JSON fields; clients must not inject `html` without sanitizing.

## DeleteAll

`DELETE /api/v1/email/{domain}/{user}`

```json
{ "success": true, "data": { "deleted": 3 }, "error": null }
```

## Cloudflare Scope

This app receives email only for domains configured in Cloudflare Email Routing on same Cloudflare account. It does not receive arbitrary DNS-provider domains without external SMTP infrastructure.

Future `chatgpt-creator` integration should call HTTP only. Do not import Worker internals, D1 schema, or R2 key format.

# Architecture

The application stack is already selected and implemented:

- Root product: Go 1.25 CLI in `cmd/register` and `internal/*`.
- Automation helper: Node.js scripts in `scripts/`.
- Standalone temp-mail product: TypeScript Cloudflare Worker in
  `cloudflare-temp-mail/`.
- Worker storage/runtime: Cloudflare Workers, D1, R2, Email Routing,
  TypeScript, Vitest, Playwright.

Use `docs/system-architecture.md` and `docs/codebase-summary.md` for the
current execution paths. This file adds harness-level architecture rules for
future changes.

## Discovery Before Shape

Before proposing new implementation shape, identify:

- Product surface: root CLI, embedded web UI, Node Codex-login script, Worker
  API/UI, email handler, or scheduled cleanup.
- Runtime stack involved: Go, Node.js, TypeScript Worker, Cloudflare bindings,
  provider APIs, browser automation, or filesystem output.
- Core domains: the product concepts that deserve stable names and contracts.
- Boundary inputs: user input, API requests, webhooks, jobs, files, credentials,
  provider payloads, and environment configuration.
- Validation ladder: the smallest checks that can prove the selected stack.

Record stack choices in `docs/decisions/` when they meaningfully constrain
future work.

## Current Runtime Boundaries

```text
cmd/register
  -> internal/config
  -> internal/register
      -> internal/email
      -> internal/sentinel
      -> internal/codex
      -> internal/web
  -> filesystem outputs

scripts/*.mjs
  -> Camofox browser automation
  -> OpenAI OAuth/Codex flow
  -> ViOTP and temp-mail providers

cloudflare-temp-mail/src
  -> API/UI router
  -> D1 metadata
  -> R2 message payloads
  -> Cloudflare Email Routing
  -> scheduled cleanup
```

Keep root CLI and `cloudflare-temp-mail` independently deployable. The root app
should consume the Worker through HTTP contracts, not by importing Worker
internals.

## Default Layering For New Code

```text
domain/rules
  <- application/use-case
      <- provider/client adapter
          <- CLI, web handler, script, or Worker route
```

Use existing package boundaries before adding folders. Create new packages only
when a story needs a stable boundary.

## Dependency Rule

Inner layers must not depend on outer layers.

| Layer | May depend on | Must not depend on |
| --- | --- | --- |
| domain | nothing project-external except tiny pure utilities | framework, database, UI, provider, process/env |
| application | domain | framework, UI, provider, database concrete clients |
| infrastructure | domain, application | interface controllers or UI |
| interface | all backend layers | UI state or platform shell assumptions |
| app surfaces | API contracts and app-facing clients | domain internals directly |

## Parse-First Boundary Rule

Unknown data must be parsed at boundaries before it enters inner code.

Boundaries include:

- CLI flags and interactive prompt values.
- HTTP request bodies, params, and query strings.
- Session payloads and identity claims.
- Environment variables.
- Database rows returned from external clients.
- Platform shell payloads.
- Deep links, tokens, and signed URLs.
- Provider pages, webhooks, email messages, browser automation state, and async
  payloads.

Target flow:

```text
unknown input
  -> parser
  -> typed DTO or command
  -> application use case
  -> domain object/value object
```

Inner layers should work with meaningful product types such as `UserId`,
`AccountId`, `WorkspaceId`, `Role`, `DateRange`, or domain-specific IDs,
rather than repeatedly validating raw strings.

## Command/Query Boundary

If the product has both reads and writes, keep command/query separation clear at
the code level even when the storage layer is simple:

- Commands mutate state and own audit side effects.
- Queries read state and format for consumers.
- Shared domain rules live in domain/application, not controllers.

## Observability Contract

For new HTTP or long-running surfaces, prefer one canonical log/event shape with:

- timestamp
- level
- request_id
- user_id when known
- action
- duration_ms
- status_code
- message

Never log secrets, cookies, OTPs, access tokens, refresh tokens, passwords, or
raw proxy credentials. Audit records are product records. Application logs are
operational records. Do not use one as a substitute for the other.

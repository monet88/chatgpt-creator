# Code Standards

## Scope

This standard reflects the current Go codebase under `cmd/` and `internal/`.

## Structure Standards

- Keep CLI entry responsibilities in `cmd/register/main.go`.
- Keep reusable logic in `internal/*` packages.
- Maintain package boundaries:
  - `config`: config loading/validation
  - `register`: network flow + batch orchestration
  - `email`: email/OTP concerns
  - `sentinel`: challenge/token concerns
  - `util`: generic helpers

## Naming and API Conventions

- Exported symbols use Go PascalCase.
- Internal helper methods on `Client` remain lower camelCase.
- JSON config keys remain snake_case to match `config.json`:
  - `proxy`
  - `output_file`
  - `default_password`
  - `default_domain`

## Error Handling

- Return contextual errors with `%w` when wrapping (`fmt.Errorf`).
- Avoid silent failure in core flow steps.
- For retryable steps, centralize retry behavior in orchestration (batch layer and OTP retry branch).

## Concurrency Rules

- Protect shared stdout with `sync.Mutex`.
- Protect shared output file writes with `sync.Mutex`.
- Use `sync/atomic` for counters (`remaining`, success/failure/attempt counts).

## Configuration Rules

- Keep defaults in `internal/config/config.go` constants.
- Validate constraints during load (example: password length >= 12 when configured).
- Use environment overrides only where explicitly supported by current config loader implementation.

## Networking Rules

- Default request headers are applied in the register client request wrapper.
- TLS profile and cookie jar must be initialized during client construction.
- New outbound dependencies should be documented in `docs/deployment-guide.md`.

## Logging Rules

- Use timestamped worker logs (`[time] [Wn]`) consistently.
- Keep summary output deterministic: target, success, attempts, failures, elapsed.

## Testing Standard (Target)

Current state: no tests found.

When adding tests later:
- Use `go test ./...` baseline.
- Add race checks: `go test -race ./...`.
- Prioritize table-driven tests for pure functions and parser logic.

## Documentation Sync Rule

When modifying behavior in any `internal/*` package, update at least:
- `docs/codebase-summary.md`
- `docs/system-architecture.md`
- `README.md` (if user-facing behavior changes)

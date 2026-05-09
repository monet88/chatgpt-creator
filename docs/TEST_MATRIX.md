# Test Matrix

This file maps product behavior to proof.

Use this as the behavior-to-proof control panel for selected work. Keep detailed
test policy in `docs/testing-guide.md`; keep architecture context in
`docs/system-architecture.md`.

## Status Values

| Status | Meaning |
| --- | --- |
| planned | Accepted as intended behavior, not implemented |
| in_progress | Actively being built |
| implemented | Implemented and proof exists |
| changed | Contract changed after earlier implementation |
| retired | No longer part of the product contract |

## Matrix

| Story | Contract | Unit | Integration | E2E | Platform | Status | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Existing CLI config | `defaults < config file < environment < flags`; empty output path resolves to `accounts/cre/<datetime>.txt` | yes | no | no | no | implemented | `go test ./...`; see `docs/testing-guide.md` |
| Existing batch registration | Worker pool, runtime stop controls, typed failures, credential write safety | yes | optional | no | no | implemented | `go test ./...`, `go test -race ./...` |
| Existing Codex token extraction | `--codex` writes per-account panel JSON; aggregate `--codex-output` is opt-in | yes | optional | browser-flow | no | implemented | `go test ./...`; script smoke when provider flow changes |
| Existing Worker temp-mail API/UI | D1/R2-backed mailbox API, inbound email handler, scheduled cleanup, UI routes | yes | yes | yes | Worker local/dev | implemented | `npm --prefix cloudflare-temp-mail run build`; `npm --prefix cloudflare-temp-mail run test` |
| New selected story | Add row when feature intake creates or selects a story packet | no | no | no | no | planned | none |

## Evidence Rules

- Unit proof covers pure domain and application rules.
- Integration proof covers backend enforcement, data integrity, provider
  behavior, jobs, or service contracts.
- E2E proof covers user-visible browser flows.
- Platform proof covers only shell, deployment, mobile, desktop, or runtime
  behavior that cannot be proven in lower layers.
- A story can be implemented without every proof column if the story packet
  explains why.
- Do not mark newly changed behavior implemented until the relevant commands
  were run for that task or the evidence gap is recorded.

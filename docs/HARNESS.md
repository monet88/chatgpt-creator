# Harness

This repository uses Harness v0 as an operating layer for agent-assisted
maintenance and feature work on an existing product.

The product is already implemented: a Go CLI for batch account registration and
a standalone Cloudflare Worker temp-mail app. The harness defines how agents
turn future prompts into bounded, documented, validated changes without
overwriting the existing project docs or runtime structure.

## Mental Model

```text
------------------+
| Human intent    |
+------------------+
         |
         v
+------------------+
| Feature intake   |
+------------------+
         |
         v
+------------------+
| Story packet     |
+------------------+
         |
         v
+------------------+
| Agent work loop  |
+------------------+
         |
         v
+------------------+
| Product delta    |
+------------------+
         |
         v
+------------------+
| Validation proof |
+------------------+
         |
         v
+------------------+
| Harness delta    |
+------------------+
         |
         v
+------------------+
| Next intent      |
+------------------+
```

Every task has two possible outputs:

1. Product delta: app code, tests, API shape, data model, or product docs.
2. Harness delta: docs, templates, validation expectations, backlog items, or
   decision records that make the next task easier.

## Adopted Harness Scope

The adopted harness includes:

- Agent entrypoint.
- Feature intake and risk lanes.
- Story templates.
- Decision log template.
- Validation report template.
- Test matrix tied to existing validation commands.
- Harness growth backlog.

It deliberately does not replace:

- Existing product docs such as `docs/project-overview-pdr.md`,
  `docs/system-architecture.md`, and `docs/codebase-summary.md`.
- Existing root Go CLI code under `cmd/` and `internal/`.
- Existing Cloudflare Worker app under `cloudflare-temp-mail/`.
- Existing operational scripts under `scripts/`.
- Existing validation commands.

New harness files should describe and coordinate the current project; they
should not imply this repository is blank or waiting for a first stack choice.

Harness v0 still excludes:

- A new monolithic `SPEC.md`.
- App source scaffolding.
- Package scripts.
- Test runner config.
- CI workflows.
- Database migrations or infrastructure.

Those should change only when a selected story or maintenance task needs them.

## Source Hierarchy

```text
User-provided spec or prompt
  input material for future changes

README.md
  operator-facing overview and quick start

docs/project-overview-pdr.md
  current product goals, requirements, acceptance criteria

docs/system-architecture.md and docs/codebase-summary.md
  current runtime surfaces, components, execution paths

docs/api-contract.md
  Worker API contract consumed by the root CLI and external clients

docs/stories/* and docs/TEST_MATRIX.md
  selected work packets and behavior-to-proof control panel

docs/decisions/*
  why product, architecture, or harness choices changed
```

Product docs plus executable tests are the living contract for existing
behavior.

## Spec Lifecycle

This repository already has accepted product behavior. Treat new user prompts
as change input, not as a replacement for the current docs.

When a future specification or larger initiative arrives, decompose it into the
current product docs, story packets, validation expectations, and decisions
instead of creating a second permanent specification.

Ongoing work should enter the harness as one of these input types:

- New spec: a project specification that needs to become product docs and
  initial story candidates.
- Spec slice: a selected behavior from the provided spec.
- Change request: a bounded behavior change, bug fix, or product refinement.
- New initiative: a larger product area that needs multiple stories.
- Maintenance request: dependency, architecture, performance, security, or
  operational work.
- Harness improvement: a process, template, proof, or agent-instruction change.

The spec-to-work loop is:

```text
human intent or supplied spec
  -> classify input type
  -> update current product contract
  -> create story packet or initiative notes when needed
  -> define validation proof
  -> implement or document the blocker
  -> update product docs, stories, test matrix, and decisions
  -> capture harness friction
```

Large product areas should use scoped initiative notes instead of a second
monolithic specification. An initiative should explain the goal, affected
product docs, candidate stories, validation shape, open decisions, and exit
criteria. If initiative work becomes a repeated pattern, add a template or
proposal to `docs/HARNESS_BACKLOG.md`.

## Growth Rule

The harness grows from friction.

When an agent is confused, repeats manual reasoning, needs a new validation
command, discovers a missing rule, or sees a recurring failure pattern, it must
either improve the harness directly or add a proposal to `HARNESS_BACKLOG.md`.

## Validation Ladder

Use the smallest command set that proves the touched surface.

```text
validate:quick
  go test ./...
  go vet ./...

test:integration
  go test -race ./...
  npm --prefix cloudflare-temp-mail run build
  npm --prefix cloudflare-temp-mail run test:unit

test:e2e
  npm --prefix cloudflare-temp-mail run test:ui

test:platform
  wrangler/local Worker smoke checks when deployment behavior changes

test:release
  full Go suite, Worker build/test suite, log/redaction checks, deployment smoke
```

Agents must not claim these commands pass until they have been run in the
current task or the final response clearly states the gap.

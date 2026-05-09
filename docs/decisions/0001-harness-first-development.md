# 0001 Harness-First Development

Date: 2026-05-05

Scope: imported upstream Harness v0 decision. This records why the generic
harness was created; it does not describe the current `chatgpt-creator` runtime
state.

## Status

Accepted

## Context

The upstream harness repository contained a product README and a large product
specification. It did not have application implementation yet.

The project will likely involve human direction plus agent implementation over
many evolving stories. A single massive specification is not enough for safe
agent work because it becomes hard to locate current truth, risk, proof, and
change history.

## Decision

Create Harness v0 before scaffolding product code.

Harness v0 defines:

- Agent entrypoint.
- Product doc split.
- Feature intake and risk lanes.
- Story packet templates.
- Decision records.
- Test matrix.
- Harness backlog.

No application code, fake scripts, CI, or tests were created by this upstream
decision.

## Consequences

Positive:

- Agents have a clear operating model before implementation starts.
- Product truth can split away from the massive spec.
- Risky work has a slower lane before code changes.
- Harness growth becomes part of the work.

Tradeoffs:

- Some docs are placeholders until real stories exercise them.
- Validation commands are only contracts until implementation begins.
- The harness must stay small enough to revise from real friction.

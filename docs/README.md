# Documentation Map

This directory holds the current product docs plus the harness files that guide
future agent-assisted work.

## Main Files

- `project-overview-pdr.md`: current product goals, requirements, and
  acceptance criteria.
- `system-architecture.md`: implemented runtime surfaces and execution flow.
- `codebase-summary.md`: quick map of packages, products, and commands.
- `testing-guide.md`: validation commands and test policy.
- `HARNESS.md`: how humans and agents collaborate.
- `FEATURE_INTAKE.md`: how prompts become tiny, normal, or high-risk work.
- `ARCHITECTURE.md`: harness-level architecture boundary rules.
- `TEST_MATRIX.md`: living map of behavior to proof.
- `HARNESS_BACKLOG.md`: improvements discovered while working.
- `GLOSSARY.md`: shared terms.

## Folders

- `product/`: index for product-contract docs. Current product truth still
  lives in the root docs until a selected story splits it further.
- `stories/`: feature packets and backlog.
- `decisions/`: durable decisions and tradeoffs.
- `templates/`: reusable spec-intake, story, plan, decision, and validation
  formats.
- `journals/`: dated technical notes from completed work.
- `proxy/`: proxy provider references and credentials guidance. Redact secrets
  before quoting.

## Current State

This repository already has implementation, tests, scripts, and deployment
docs. Harness v0 is installed as an overlay for future work intake, story
tracking, validation evidence, and decision records.

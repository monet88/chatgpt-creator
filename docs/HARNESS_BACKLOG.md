# Harness Backlog

Use this file when an agent discovers a missing harness capability but should
not change the operating model immediately.

## Template

```md
## Missing Harness Capability

### Title

Short name.

### Discovered While

Task or story that exposed the gap.

### Current Pain

What was hard, repeated, ambiguous, or unsafe?

### Suggested Improvement

What should be added or changed?

### Risk

Tiny, normal, or high-risk.

### Status

proposed | accepted | implemented | rejected
```

## Items

## Missing Harness Capability

### Title

Align installer payload with documentation map.

### Discovered While

Adopting Harness v0 into the existing `chatgpt-creator` repository with
`--merge`.

### Current Pain

The upstream docs map referenced `docs/demo/`, but the installed payload did not
create that folder. Local docs were updated to remove the missing reference.

### Suggested Improvement

Either include the demo folder in the installer payload or keep the installed
documentation map limited to files the installer actually writes.

### Risk

Tiny.

### Status

proposed

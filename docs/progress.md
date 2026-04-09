# Progress

## Purpose

Track what was done, what failed, what decision was made, and how to resume.

## Storage

- `runs/current` = active run id
- `runs/<run-id>/progress.json` = authoritative progress record

## Record shape

- done
- failed
- decision
- resume

## Example

```json
{
  "run_id": "2026-04-09-001",
  "updated_at": "2026-04-09T10:00:00Z",
  "state": "blocked",
  "done": ["parsed config", "loaded inputs"],
  "failed": [
    {"step": "write output", "reason": "permission denied"}
  ],
  "decision": {
    "value": "switch output path to runs/tmp",
    "why": "workspace is read-only"
  },
  "resume": {
    "next_step": "write output to runs/tmp",
    "cursor": "step-3",
    "notes": "retry after path fix"
  }
}
```

## Resume flow

1. Read `runs/current`
2. Load `runs/<run-id>/progress.json`
3. Continue from `resume.cursor` or `resume.next_step`
4. Update the record after each meaningful step
5. When a new run starts, write a new progress file and update `runs/current`

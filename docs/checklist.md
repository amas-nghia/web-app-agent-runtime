# Build Checklist

Use this file to track what is done, what failed, what was chosen, and how it was fixed.

## Milestones

- [x] Repo skeleton created
- [x] Canonical state model added
- [x] Workflow sequencing added
- [x] Mock adapter added
- [x] SQLite run-state store added
- [x] No-op sandbox boundary added
- [x] CLI `run` smoke path added
- [x] Contract and smoke tests added
- [x] Progress artifacts added (`runs/current`, `runs/<run-id>/progress.json`)
- [x] OpenCode adapter shell-out wiring added
- [x] Container sandbox API + noop backend wrapper added
- [x] Docker-backed sandbox runtime added (opt-in via `WABR_SANDBOX=docker`)
- [x] VM sandbox backend scaffold added (opt-in via `WABR_SANDBOX=vm`)
- [x] Structured debate artifacts added
- [x] Bugfix/resume flow with replay
- [x] `status` command added
- [x] `replay` command added
- [x] Replay flow added
- [x] `bugfix` command added
- [x] Run history JSONL added (`runs/<run-id>/history.jsonl`)

## Work log

| Area | Status | Done | Issue / Failure | Decision | Fix |
| --- | --- | --- | --- | --- | --- |
| CLI smoke path | done | `run` executes mock-backed flow | Initial skeleton had no run flow | Default to `mock` for early testing | Added `run` command + coordinator |
| Canonical state | done | Normalized run/task/attempt/approval/checkpoint/artifact model | None | Use versioned engine bindings | Added Go structs + SQLite store |
| Sandbox | done | No-op boundary exists | No real isolation yet | Keep interface stable first | Added deterministic boundary stub |
| Progress artifacts | done | `runs/current` and `progress.json` written | None | Keep file-based resume state | Added `internal/progress` |
| OpenCode adapter | done | shells out to `opencode run` | CLI binary may be missing locally | Default to mock in smoke path | Added external command wiring |
| Sandbox API | done | container-first boundary + noop runtime wrapper | Real backend needed opt-in selection | Keep interface small and swappable | Added `Spec`, `Runtime`, and `NewContainer` |
| Docker sandbox | done | opt-in docker runtime via env | VM backend still pending | Prefer explicit `WABR_SANDBOX=docker` | Added `dockerRuntime` |
| VM sandbox scaffold | done | opt-in `vm` runtime selection exists | real VM execution still not implemented | Keep the VM boundary explicit | Added `vmRuntime` |
| Sandbox env tests | done | default/docker/vm selection covered | Optional host binaries vary | Make selection explicit and deterministic | Added `runtime_test.go` and `vm_runtime_test.go` |
| Debate artifacts | done | per-step debate files written under `runs/<run-id>/debates/` | None | Keep debate records file-based | Added `internal/debate` |
| Resume flow | done | resume continues the live run from checkpoint/start step and keeps the current run id | replay still creates a new run id for inspection | Continue from current progress file and saved checkpoint | Added `resume` command behavior |
| Status command | done | prints current run + progress + checkpoint | None | Keep it read-only | Added `status` |
| Replay command | done | reruns from current run + checkpoint fallback | Uses current run as input | Rebuild using canonical state | Added `replay` |
| Replay flow | done | replay rebuilds a prior run from `runs/current`, `progress.json`, and debate artifacts | resume continues the live run instead of reconstructing history | Re-run the recorded path from checkpoint/start step | Added replay flow notes |
| Bugfix command | done | `bugfix` starts runs in bugfix mode with a required bug report and no mode override | Post-release defect work was only reachable through `run --mode bugfix` | Add a dedicated command for the fix workflow | Added `bugfix` command + smoke coverage + bug-report enforcement |
| SQLite migration | done | DB schema now adds `start_step` safely for old databases | `start_step` was missing in existing `runtime.db` | Auto-migrate on startup | Added `ensureColumn(start_step)` |
| Run history | done | JSONL history records run/step/approval/checkpoint/failure events | Status output was too raw and replay had no first-class event log | Keep a compact operator summary and JSONL history | Added `internal/history` and status summary |

## Current blockers

- OpenCode adapter is wired but still minimal.
- VM runtime is scaffolded; a real VM execution backend still needs implementation.

## Locked choices

- Web-app-only scope for MVP
- Engine-agnostic orchestrator
- PostgreSQL-compatible canonical state (SQLite for early testing)
- One worktree per run
- Configurable approval tiers: `auto`, `ask`, `block`
- Mock engine is the default for smoke tests

## Notes

- Update this file before and after every meaningful change.
- When a task fails, record the failure reason, decision, and fix in the table.

# Web App Agent Runtime

License: MIT

Runtime/orchestrator skeleton for building and maintaining web apps for mobile and desktop.

## What this is

- a workflow runtime that coordinates tasks, approvals, artifacts, and retries
- a backend-agnostic skeleton that starts with a mock adapter
- a testable control plane for web app work

## What this is not

- not a full web app generator end-to-end out of the box
- not a finished product that emits a production app on the mock path
- not a replacement for a real frontend/backend implementation

## Goal

This system helps you orchestrate the work needed to build a web app:

- intake a web app idea
- split it into tasks automatically
- debate implementation choices before coding
- build and test the implementation behind a web app
- fix bugs after release
- keep approval policy configurable
- remain engine-agnostic from day one

## Current design decisions

- Orchestrator is the source of truth.
- Engine adapters are swappable.
- SQLite holds canonical state for early testing; the schema stays PostgreSQL-friendly.
- Repo files hold readable artifacts only.
- One git worktree per run.
- Container-first sandboxing, VM only for risky steps.
- Approval policy is config-driven.
- OpenCode is the first adapter, not the system of record.

## Repo layout

```text
.
├─ cmd/orchestrator/
├─ internal/
│  ├─ orchestrator/
│  │  └─ workflow.go
│  ├─ adapters/
│  │  ├─ opencode/
│  │  └─ mock/
│  ├─ state/
│  │  └─ sqlite/
│  ├─ policy/
│  ├─ sandbox/
│  └─ telemetry/
├─ configs/
├─ docs/
└─ runs/
```

## Working docs

- `docs/architecture.md`
- `docs/contracts.md`
- `docs/testing.md`
- `docs/decision-log.md`
- `docs/roadmap.md`

## MVP definition

The MVP is done when the system can orchestrate a web app workflow, and the control-flow path is verifiable with the mock adapter:

1. create a run from a web app goal
2. split work into tasks
3. debate implementation options before coding
4. build and test the implementation behind a web app
5. resume after a failure or bug report
6. apply approvals from config
7. swap the execution backend through an adapter

## Quick start

Run the mock-backed smoke flow. This verifies the orchestration path only; it does not generate a real app:

```bash
GOTOOLCHAIN=local go run ./cmd/orchestrator run "build a responsive web app"
```

Run against a custom state file:

```bash
GOTOOLCHAIN=local go run ./cmd/orchestrator run --state ./runs/state.db "build a responsive web app"
```

The default backend is `mock`, so early testing stays deterministic.

If you want to generate a real app, you must wire a real backend adapter and sandbox first.

Use `bugfix` when you already have a reported defect and want the run to start in the bugfix path instead of a fresh intake. It requires `--bug-report`:

```bash
GOTOOLCHAIN=local go run ./cmd/orchestrator bugfix --bug-report "checkout fails on mobile" "fix the checkout flow"
```

It records the bug report in the run state and follows the debate → bugfix → build → test → release path.

If you want the sandbox layer to create a real container boundary, set:

```bash
WABR_SANDBOX=docker
```

The docker sandbox is opt-in; the default remains a lightweight no-op boundary for fast local smoke tests.

Check the current run state:

```bash
GOTOOLCHAIN=local go run ./cmd/orchestrator status
```

Replay the current run from its saved checkpoint/state:

```bash
GOTOOLCHAIN=local go run ./cmd/orchestrator replay
```

Useful artifacts:

- `runs/current`
- `runs/<run-id>/progress.json`
- `runs/<run-id>/history.jsonl`
- `runs/<run-id>/debates/<step>.json`

Replay uses the canonical run record plus the saved progress checkpoint; `resume` continues the current run id from its checkpoint/start step, while `replay` creates a new replay run by default.
The `status` command prints a compact progress summary and the last history event instead of raw slices.

## Output contract

The mock smoke path is a control-flow check, not a full app generator.

After a successful mock run, the durable outputs are:

- `runs/current`
- `runs/<run-id>/progress.json`
- `runs/<run-id>/history.jsonl`
- `runs/<run-id>/debates/<step>.json`

The mock path also records placeholder artifact refs such as `mock-artifact` or `release task-artifact`.

### Tarot demo example

Input goal:

```text
build a beautiful MVC tarot reading web app with frontend, backend, and complete seed data
```

Expected result on the mock smoke path:

- the run completes successfully
- workflow metadata is persisted
- placeholder artifacts are written
- no runnable tarot app is produced until a real backend adapter is wired

## Current runtime path

- CLI `run` command is the first executable path.
- SQLite backs the early canonical run store.
- Mock adapter is the default adapter for fast verification.
- OpenCode adapter is present as the first swappable backend stub.
- `status` reads the current run, progress, and canonical state summary.

The default smoke path is intentionally non-production and should be read as a control-flow test, not a real app build.

## Replay flow

Replay rebuilds a prior run from saved artifacts so you can inspect or reproduce it.
It reads `runs/current`, `runs/<run-id>/progress.json`, and `runs/<run-id>/debates/*.json`.
When a checkpoint exists, replay starts from the recorded checkpoint step; otherwise it falls back to the stored run start step.
Unlike resume, replay creates a fresh replay run id by default; resume continues the live run from its checkpoint/start step.
Verify by replaying a run and checking the step order and regenerated artifacts match the saved run.

## Notes to stay on track

- Do not let engine internals become canonical state.
- Do not let the main workspace become the execution target.
- Do not hardcode approval policy.
- Do not widen scope beyond web apps until the MVP is stable.
- Do not block on engine choice for the first smoke path; default to mock.

## Canonical model

- `Run` is the aggregate root.
- `Task` is one unit of planned work.
- `Attempt` is one execution of a task.
- `Approval` gates risky attempts.
- `Checkpoint` is the resume anchor.
- `ArtifactRef` is the durable output record.

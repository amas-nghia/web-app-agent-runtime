# Architecture

## System shape

- `cmd/orchestrator`: entrypoint
- `internal/orchestrator`: workflow engine
- `internal/adapters`: backend implementations
- `internal/state`: canonical run state
- `internal/policy`: approval and safety rules
- `internal/sandbox`: isolated execution boundary
- `internal/telemetry`: logs, traces, metrics

## Canonical state model

Use a normalized graph with `Run` as the aggregate root:

- `Run` owns the plan, current step, and lifecycle
- `Task` models one unit of work in the plan
- `Attempt` models one execution of one task
- `Approval` models a policy gate on an attempt
- `Checkpoint` models a resume point tied to a run and usually a task or attempt
- `ArtifactRef` models durable outputs; repo files are artifacts, not canonical state
- `AuditEntry` records state transitions for traceability

Relationships:

- one run has many tasks, attempts, approvals, checkpoints, and artifacts
- one task has many attempts and at most one current attempt
- one attempt may point to one approval and one checkpoint
- one checkpoint may be reused by a later resume run

## Invariants

- orchestrator owns workflow state
- adapter owns engine-specific execution only
- repo files are artifacts only
- sandbox owns execution isolation
- CI verifies; it does not define truth

## Engine abstraction

Every backend must satisfy a stable adapter contract:

- capability negotiation
- start run
- execute task
- request approval
- apply approval decision
- fetch artifacts
- resume
- cancel

## MVP workflow

The workflow engine is intentionally small and sequence-driven:

- intake mode: intake -> debate -> build -> test -> release
- bugfix mode: debate -> bugfix -> build -> test -> release
- resume mode: resume -> test -> release

Rules:

- intake creates the initial run and task graph
- debate resolves implementation choices before code changes
- build creates code-changing attempts
- test validates the build and can produce a checkpoint
- release publishes the approved build
- bugfix starts from a reported defect instead of a fresh intake
- resume starts from a checkpoint and skips directly to the recovery path

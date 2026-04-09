# Contracts

## Adapter contract

Adapters are execution backends, not sources of truth.
The orchestrator owns workflow state and asks adapters only for engine-specific execution.

Each adapter must provide:

- `Name()` for logging and persisted engine binding
- `Capabilities()` for negotiation
- `StartRun()` to create or attach a backend run
- `ExecuteTask()` to run one task attempt
- `RequestApproval()` / `ApplyApproval()` for approval-gated steps
- `FetchArtifacts()` to collect produced files
- `Resume()` and `Cancel()` for lifecycle control

## Capability negotiation

Use the adapter's offered `CapabilitySet` against the policy-required set.
Required capabilities are computed per workflow step.

Canonical bits today:

- `Read`
- `Write`
- `Shell`
- `Resume`
- `ApprovalMemo`

Negotiation rule:

- missing required bits fail the run before execution starts
- selected capabilities are the intersection of offered and required

## Canonical objects

- `RunSpec`
- `TaskSpec`
- `Observation`
- `Action`
- `ApprovalRequest`
- `ApprovalDecision`
- `ArtifactRef`
- `ExecutionResult`
- `CapabilitySet`

## Event vocabulary

- run started
- task dispatched
- approval requested
- approval granted
- artifact produced
- task failed
- checkpoint stored
- run completed

package adapters

import "web-app-agent-runtime/internal/state"

type Adapter interface {
	Name() string
	Capabilities() state.CapabilitySet
	StartRun(state.RunSpec) (string, error)
	ExecuteTask(state.TaskSpec) (state.ExecutionResult, error)
	RequestApproval(state.ApprovalRequest) (state.ApprovalDecision, error)
	ApplyApproval(state.ApprovalDecision) error
	FetchArtifacts(runID string) ([]state.ArtifactRef, error)
	Resume(runID string) error
	Cancel(runID string) error
}

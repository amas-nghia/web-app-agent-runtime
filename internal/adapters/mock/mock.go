package mock

import (
	"sync"

	"web-app-agent-runtime/internal/state"
)

type Call struct {
	Method string
	RunID  string
	TaskID string
	Step   string
	Reason string
}

type Adapter struct {
	mu    sync.Mutex
	calls []Call
}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "mock" }

func (a *Adapter) Capabilities() state.CapabilitySet {
	return state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}
}

func (a *Adapter) StartRun(spec state.RunSpec) (string, error) {
	a.record(Call{Method: "StartRun", Step: string(spec.Mode), Reason: spec.Goal})
	return "mock-run", nil
}

func (a *Adapter) ExecuteTask(task state.TaskSpec) (state.ExecutionResult, error) {
	a.record(Call{Method: "ExecuteTask", RunID: task.RunID, TaskID: task.Title, Step: string(task.Step), Reason: task.Prompt})
	return state.ExecutionResult{
		RunID:        task.RunID,
		TaskID:       task.Title,
		AttemptID:    task.Title + "-attempt",
		CheckpointID: task.Title + "-checkpoint",
		Summary:      "mock-result",
		Artifacts: []state.ArtifactRef{{
			ID:        task.Title + "-artifact",
			RunID:     task.RunID,
			TaskID:    task.Title,
			AttemptID: task.Title + "-attempt",
			Kind:      state.ArtifactKindTestLog,
			Path:      "mock-artifact",
		}},
	}, nil
}

func (a *Adapter) RequestApproval(req state.ApprovalRequest) (state.ApprovalDecision, error) {
	a.record(Call{Method: "RequestApproval", RunID: req.RunID, TaskID: req.TaskID, Step: string(req.Step), Reason: req.Reason})
	approved := req.Tier != state.ApprovalBlock
	memo := "mock-approved"
	if !approved {
		memo = "mock-blocked"
	}
	return state.ApprovalDecision{
		RequestID: req.TaskID,
		Approved:  approved,
		Memo:      memo,
	}, nil
}

func (a *Adapter) ApplyApproval(decision state.ApprovalDecision) error {
	a.record(Call{Method: "ApplyApproval", RunID: decision.RequestID, Reason: decision.Memo})
	return nil
}

func (a *Adapter) FetchArtifacts(runID string) ([]state.ArtifactRef, error) {
	a.record(Call{Method: "FetchArtifacts", RunID: runID})
	return []state.ArtifactRef{{Kind: state.ArtifactKindSource, Path: "mock-artifact"}}, nil
}

func (a *Adapter) Resume(runID string) error {
	a.record(Call{Method: "Resume", RunID: runID})
	return nil
}

func (a *Adapter) Cancel(runID string) error {
	a.record(Call{Method: "Cancel", RunID: runID})
	return nil
}

func (a *Adapter) Calls() []Call {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]Call, len(a.calls))
	copy(out, a.calls)
	return out
}

func (a *Adapter) record(call Call) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls = append(a.calls, call)
}

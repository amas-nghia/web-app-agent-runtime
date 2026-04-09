package opencode

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"web-app-agent-runtime/internal/state"
)

var ErrNotImplemented = errors.New("opencode adapter: not implemented")

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "opencode" }

func (a *Adapter) Capabilities() state.CapabilitySet {
	return state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}
}

func (a *Adapter) StartRun(spec state.RunSpec) (string, error) {
	if spec.RunID != "" {
		return spec.RunID, nil
	}
	return fmt.Sprintf("opencode-%s", strings.ReplaceAll(spec.Goal, " ", "-")), nil
}

func (a *Adapter) ExecuteTask(task state.TaskSpec) (state.ExecutionResult, error) {
	cmd := exec.Command("opencode", "run", task.Prompt)
	output, err := cmd.CombinedOutput()
	result := state.ExecutionResult{
		RunID:        task.RunID,
		TaskID:       task.Title,
		AttemptID:    task.Title + "-attempt",
		CheckpointID: task.Title + "-checkpoint",
		Summary:      strings.TrimSpace(string(output)),
		Artifacts: []state.ArtifactRef{{
			ID:        task.Title + "-artifact",
			RunID:     task.RunID,
			TaskID:    task.Title,
			AttemptID: task.Title + "-attempt",
			Kind:      state.ArtifactKindSource,
			Path:      "opencode-output",
		}},
	}
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("opencode adapter execute task: %w", err)
	}
	return result, nil
}
func (a *Adapter) RequestApproval(req state.ApprovalRequest) (state.ApprovalDecision, error) {
	approved := req.Tier != state.ApprovalBlock
	return state.ApprovalDecision{RequestID: req.TaskID, Approved: approved, Memo: "opencode-adapter"}, nil
}
func (a *Adapter) ApplyApproval(decision state.ApprovalDecision) error { return nil }
func (a *Adapter) FetchArtifacts(runID string) ([]state.ArtifactRef, error) {
	return []state.ArtifactRef{{ID: runID + "-artifact", RunID: runID, Kind: state.ArtifactKindSource, Path: "opencode-output"}}, nil
}
func (a *Adapter) Resume(runID string) error { return nil }
func (a *Adapter) Cancel(runID string) error { return nil }

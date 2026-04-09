package opencode

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), opencodeTimeout())
	defer cancel()

	cmd := exec.CommandContext(ctx, "opencode", "run", "--non-interactive", task.Prompt)
	cmd.Stdin = strings.NewReader("")
	cmd.Env = append(os.Environ(), "OPENCODE_NON_INTERACTIVE=1")
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
		if ctx.Err() == context.DeadlineExceeded {
			return result, fmt.Errorf("opencode adapter execute task: timeout after %s: %w", opencodeTimeout(), err)
		}
		return result, fmt.Errorf("opencode adapter execute task: %w", err)
	}
	return result, nil
}

func opencodeTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("WABR_OPENCODE_TIMEOUT"))
	if raw == "" {
		return 2 * time.Minute
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 2 * time.Minute
	}
	return d
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

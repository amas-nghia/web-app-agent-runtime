package mock

import (
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestAdapter_DeterministicScript(t *testing.T) {
	a := New()

	if got := a.Name(); got != "mock" {
		t.Fatalf("Name() = %q, want %q", got, "mock")
	}

	if got := a.Capabilities(); !got.Shell {
		t.Fatalf("Capabilities().Shell = false, want true")
	}

	runID, err := a.StartRun(state.RunSpec{Goal: "build a form", Mode: state.RunModeIntake})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if runID != "mock-run" {
		t.Fatalf("StartRun() = %q, want %q", runID, "mock-run")
	}

	result, err := a.ExecuteTask(state.TaskSpec{RunID: runID, Title: "plan", Prompt: "draft the plan", Step: state.WorkflowStepDebate})
	if err != nil {
		t.Fatalf("ExecuteTask() error = %v", err)
	}
	if result.Summary != "mock-result" {
		t.Fatalf("ExecuteTask().Summary = %q, want %q", result.Summary, "mock-result")
	}
	if len(result.Artifacts) != 1 || result.Artifacts[0].Path != "mock-artifact" {
		t.Fatalf("ExecuteTask().Artifacts = %#v, want mock-artifact", result.Artifacts)
	}

	decision, err := a.RequestApproval(state.ApprovalRequest{RunID: runID, TaskID: "plan", Step: state.WorkflowStepBuild, Tier: state.ApprovalAsk, Reason: "needs approval"})
	if err != nil {
		t.Fatalf("RequestApproval() error = %v", err)
	}
	if !decision.Approved {
		t.Fatalf("RequestApproval().Approved = false, want true")
	}

	if err := a.ApplyApproval(decision); err != nil {
		t.Fatalf("ApplyApproval() error = %v", err)
	}

	arts, err := a.FetchArtifacts(runID)
	if err != nil {
		t.Fatalf("FetchArtifacts() error = %v", err)
	}
	if len(arts) != 1 || arts[0].Kind != state.ArtifactKindSource || arts[0].Path != "mock-artifact" {
		t.Fatalf("FetchArtifacts() = %#v, want mock-artifact source", arts)
	}

	if err := a.Resume(runID); err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if err := a.Cancel(runID); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	calls := a.Calls()
	if len(calls) != 7 {
		t.Fatalf("Calls() len = %d, want 7", len(calls))
	}
	if calls[0].Method != "StartRun" || calls[1].Method != "ExecuteTask" || calls[2].Method != "RequestApproval" {
		t.Fatalf("Calls() order = %#v", calls)
	}
}

package orchestrator

import (
	"strings"
	"testing"

	"web-app-agent-runtime/internal/adapters"
	"web-app-agent-runtime/internal/policy"
	"web-app-agent-runtime/internal/state"
)

type fakeAdapter struct {
	name string
	caps state.CapabilitySet
}

func (f fakeAdapter) Name() string                           { return f.name }
func (f fakeAdapter) Capabilities() state.CapabilitySet      { return f.caps }
func (f fakeAdapter) StartRun(state.RunSpec) (string, error) { return "", nil }
func (f fakeAdapter) ExecuteTask(state.TaskSpec) (state.ExecutionResult, error) {
	return state.ExecutionResult{}, nil
}
func (f fakeAdapter) RequestApproval(state.ApprovalRequest) (state.ApprovalDecision, error) {
	return state.ApprovalDecision{}, nil
}
func (f fakeAdapter) ApplyApproval(state.ApprovalDecision) error         { return nil }
func (f fakeAdapter) FetchArtifacts(string) ([]state.ArtifactRef, error) { return nil, nil }
func (f fakeAdapter) Resume(string) error                                { return nil }
func (f fakeAdapter) Cancel(string) error                                { return nil }

func TestOrchestrator_NegotiateSmoke(t *testing.T) {
	tests := []struct {
		name    string
		adapter adapters.Adapter
		step    state.WorkflowStep
		wantErr string
		wantSel state.CapabilitySet
		wantReq state.CapabilitySet
	}{
		{
			name:    "intake selects the exact intersection",
			adapter: fakeAdapter{name: "mock", caps: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}},
			step:    state.WorkflowStepIntake,
			wantReq: state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
			wantSel: state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
		},
		{
			name:    "build fails without shell",
			adapter: fakeAdapter{name: "no-shell", caps: state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true}},
			step:    state.WorkflowStepBuild,
			wantErr: "missing capabilities: [shell]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			o := New(tt.adapter, policy.Default())

			// Act.
			got, err := o.Negotiate(tt.step)

			// Assert.
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Negotiate() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Negotiate() error = %v", err)
			}
			if got.Required != tt.wantReq {
				t.Fatalf("Negotiate().Required = %#v, want %#v", got.Required, tt.wantReq)
			}
			if got.Selected != tt.wantSel {
				t.Fatalf("Negotiate().Selected = %#v, want %#v", got.Selected, tt.wantSel)
			}
		})
	}
}

func TestOrchestrator_NegotiateNilAdapter(t *testing.T) {
	t.Parallel()

	// Arrange.
	var o *Orchestrator

	// Act.
	_, err := o.Negotiate(state.WorkflowStepIntake)

	// Assert.
	if err == nil {
		t.Fatal("Negotiate() error = nil, want failure for nil orchestrator")
	}
}

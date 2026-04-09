package orchestrator

import (
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestNewWorkflow_Next(t *testing.T) {
	tests := []struct {
		name     string
		mode     state.RunMode
		start    state.WorkflowStep
		current  state.WorkflowStep
		wantNext state.WorkflowStep
		wantOK   bool
	}{
		{
			name:     "intake path advances to debate",
			mode:     state.RunModeIntake,
			start:    state.WorkflowStepUnknown,
			current:  state.WorkflowStepIntake,
			wantNext: state.WorkflowStepDebate,
			wantOK:   true,
		},
		{
			name:     "start step trims the workflow",
			mode:     state.RunModeIntake,
			start:    state.WorkflowStepBuild,
			current:  state.WorkflowStepBuild,
			wantNext: state.WorkflowStepTest,
			wantOK:   true,
		},
		{
			name:     "unknown step starts the workflow",
			mode:     state.RunModeIntake,
			start:    state.WorkflowStepUnknown,
			current:  state.WorkflowStepUnknown,
			wantNext: state.WorkflowStepIntake,
			wantOK:   true,
		},
		{
			name:     "bugfix path advances through bugfix",
			mode:     state.RunModeBugfix,
			start:    state.WorkflowStepUnknown,
			current:  state.WorkflowStepDebate,
			wantNext: state.WorkflowStepBugfix,
			wantOK:   true,
		},
		{
			name:     "resume path advances to test",
			mode:     state.RunModeResume,
			start:    state.WorkflowStepUnknown,
			current:  state.WorkflowStepResume,
			wantNext: state.WorkflowStepTest,
			wantOK:   true,
		},
		{
			name:     "terminal release step stops",
			mode:     state.RunModeIntake,
			start:    state.WorkflowStepUnknown,
			current:  state.WorkflowStepRelease,
			wantNext: state.WorkflowStepUnknown,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			workflow := NewWorkflow(tt.mode, tt.start)

			// Act.
			gotNext, gotOK := workflow.Next(tt.current)

			// Assert.
			if gotNext != tt.wantNext || gotOK != tt.wantOK {
				t.Fatalf("Next(%q) = (%q, %v), want (%q, %v)", tt.current, gotNext, gotOK, tt.wantNext, tt.wantOK)
			}
		})
	}
}

func TestWorkflow_NextEmptySequence(t *testing.T) {
	t.Parallel()

	// Arrange.
	workflow := Workflow{}

	// Act.
	gotNext, gotOK := workflow.Next(state.WorkflowStepIntake)

	// Assert.
	if gotNext != state.WorkflowStepUnknown || gotOK {
		t.Fatalf("Next() = (%q, %v), want (unknown, false)", gotNext, gotOK)
	}
}

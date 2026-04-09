package policy

import (
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestDefault_RequiredCapabilities(t *testing.T) {
	tests := []struct {
		name string
		step state.WorkflowStep
		want state.CapabilitySet
	}{
		{
			name: "intake keeps approval memo",
			step: state.WorkflowStepIntake,
			want: state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
		},
		{
			name: "build requires shell",
			step: state.WorkflowStepBuild,
			want: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true},
		},
		{
			name: "test auto-approval still requires memo capability",
			step: state.WorkflowStepTest,
			want: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true},
		},
		{
			name: "release asks for approval memo",
			step: state.WorkflowStepRelease,
			want: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true},
		},
		{
			name: "resume keeps non-shell flow",
			step: state.WorkflowStepResume,
			want: state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			p := Default()

			// Act.
			got := p.RequiredCapabilities(tt.step)

			// Assert.
			if got != tt.want {
				t.Fatalf("RequiredCapabilities(%q) = %#v, want %#v", tt.step, got, tt.want)
			}
		})
	}
}

func TestPolicy_RequiredCapabilitiesHonorsOverrides(t *testing.T) {
	t.Parallel()

	// Arrange.
	p := Policy{DefaultApproval: state.ApprovalAsk, StepApproval: map[state.WorkflowStep]state.ApprovalTier{state.WorkflowStepBuild: state.ApprovalBlock}}

	// Act.
	got := p.RequiredCapabilities(state.WorkflowStepBuild)

	// Assert.
	want := state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true}
	if got != want {
		t.Fatalf("RequiredCapabilities(build) = %#v, want %#v", got, want)
	}
}

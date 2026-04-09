package orchestrator

import (
	"testing"

	"web-app-agent-runtime/internal/adapters/mock"
	"web-app-agent-runtime/internal/policy"
	"web-app-agent-runtime/internal/state"
)

func TestOrchestrator_Negotiate(t *testing.T) {
	o := New(mock.New(), policy.Default())

	tests := []struct {
		name   string
		step   state.WorkflowStep
		wantOK bool
	}{
		{name: "intake passes", step: state.WorkflowStepIntake, wantOK: true},
		{name: "build passes with shell", step: state.WorkflowStepBuild, wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			negotiation, err := o.Negotiate(tt.step)
			if tt.wantOK && err != nil {
				t.Fatalf("Negotiate() error = %v", err)
			}
			if !tt.wantOK && err == nil {
				t.Fatal("Negotiate() error = nil, want failure")
			}
			if tt.wantOK {
				if !negotiation.Offered.Satisfies(negotiation.Required) {
					t.Fatalf("negotiation should satisfy required: %#v", negotiation)
				}
			}
		})
	}
}

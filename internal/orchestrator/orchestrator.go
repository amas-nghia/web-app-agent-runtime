package orchestrator

import (
	"fmt"

	"web-app-agent-runtime/internal/adapters"
	"web-app-agent-runtime/internal/policy"
	"web-app-agent-runtime/internal/state"
)

// Orchestrator owns canonical workflow decisions.
type Orchestrator struct {
	adapter adapters.Adapter
	policy  policy.Policy
}

type CapabilityNegotiation struct {
	Required state.CapabilitySet
	Offered  state.CapabilitySet
	Selected state.CapabilitySet
}

func New(adapter adapters.Adapter, p policy.Policy) *Orchestrator {
	return &Orchestrator{adapter: adapter, policy: p}
}

func (o *Orchestrator) Plan(goal string) state.RunSpec {
	return state.RunSpec{Goal: goal, Mode: state.RunModeIntake}
}

func (o *Orchestrator) Negotiate(step state.WorkflowStep) (CapabilityNegotiation, error) {
	if o == nil || o.adapter == nil {
		return CapabilityNegotiation{}, fmt.Errorf("orchestrator: adapter is nil")
	}

	offered := o.adapter.Capabilities()
	required := o.policy.RequiredCapabilities(step)
	if !offered.Satisfies(required) {
		return CapabilityNegotiation{}, fmt.Errorf("adapter %s missing capabilities: %v", o.adapter.Name(), offered.Missing(required))
	}

	return CapabilityNegotiation{
		Required: required,
		Offered:  offered,
		Selected: offered.Intersect(required),
	}, nil
}

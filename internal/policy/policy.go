package policy

import "web-app-agent-runtime/internal/state"

type Policy struct {
	DefaultApproval state.ApprovalTier
	StepApproval    map[state.WorkflowStep]state.ApprovalTier
}

func Default() Policy {
	return Policy{
		DefaultApproval: state.ApprovalAsk,
		StepApproval: map[state.WorkflowStep]state.ApprovalTier{
			state.WorkflowStepBuild:   state.ApprovalAsk,
			state.WorkflowStepTest:    state.ApprovalAuto,
			state.WorkflowStepRelease: state.ApprovalAsk,
		},
	}
}

func (p Policy) RequiredCapabilities(step state.WorkflowStep) state.CapabilitySet {
	required := state.CapabilitySet{
		Read:   true,
		Write:  true,
		Resume: true,
	}

	switch step {
	case state.WorkflowStepBuild, state.WorkflowStepTest, state.WorkflowStepRelease:
		required.Shell = true
	}

	tier, ok := p.StepApproval[step]
	if !ok {
		tier = p.DefaultApproval
	}
	if tier != state.ApprovalBlock {
		required.ApprovalMemo = true
	}

	return required
}

func (p Policy) StepApprovalOrDefault(step state.WorkflowStep) state.ApprovalTier {
	if tier, ok := p.StepApproval[step]; ok {
		return tier
	}
	return p.DefaultApproval
}

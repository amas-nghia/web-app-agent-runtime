package orchestrator

import "web-app-agent-runtime/internal/state"

type Workflow struct {
	Steps []state.WorkflowStep
}

func NewWorkflow(mode state.RunMode, start state.WorkflowStep) Workflow {
	steps := stepsForMode(mode)
	if start == state.WorkflowStepUnknown {
		return Workflow{Steps: steps}
	}
	for i, step := range steps {
		if step == start {
			return Workflow{Steps: steps[i:]}
		}
	}
	return Workflow{Steps: steps}
}

func (w Workflow) Next(current state.WorkflowStep) (state.WorkflowStep, bool) {
	for i, step := range w.Steps {
		if step != current {
			continue
		}
		if i+1 >= len(w.Steps) {
			return state.WorkflowStepUnknown, false
		}

		return w.Steps[i+1], true
	}

	if len(w.Steps) == 0 {
		return state.WorkflowStepUnknown, false
	}

	return w.Steps[0], true
}

func stepsForMode(mode state.RunMode) []state.WorkflowStep {
	switch mode {
	case state.RunModeBugfix:
		return []state.WorkflowStep{
			state.WorkflowStepDebate,
			state.WorkflowStepBugfix,
			state.WorkflowStepBuild,
			state.WorkflowStepTest,
			state.WorkflowStepRelease,
		}
	case state.RunModeResume:
		return []state.WorkflowStep{
			state.WorkflowStepResume,
			state.WorkflowStepTest,
			state.WorkflowStepRelease,
		}
	default:
		return []state.WorkflowStep{
			state.WorkflowStepIntake,
			state.WorkflowStepDebate,
			state.WorkflowStepBuild,
			state.WorkflowStepTest,
			state.WorkflowStepRelease,
		}
	}
}

package orchestrator

import (
	"context"
	"fmt"
	"time"

	"web-app-agent-runtime/internal/adapters"
	"web-app-agent-runtime/internal/debate"
	"web-app-agent-runtime/internal/history"
	"web-app-agent-runtime/internal/policy"
	"web-app-agent-runtime/internal/progress"
	"web-app-agent-runtime/internal/sandbox"
	"web-app-agent-runtime/internal/state"
)

type Coordinator struct {
	orchestrator *Orchestrator
	store        state.RunStore
	sandbox      sandbox.Boundary
}

func NewCoordinator(adapter adapters.Adapter, p policy.Policy, store state.RunStore, boundary sandbox.Boundary) *Coordinator {
	return &Coordinator{
		orchestrator: New(adapter, p),
		store:        store,
		sandbox:      boundary,
	}
}

func (c *Coordinator) Execute(ctx context.Context, spec state.RunSpec) (state.Run, error) {
	if c == nil || c.orchestrator == nil {
		return state.Run{}, fmt.Errorf("coordinator: orchestrator is nil")
	}

	workflow := NewWorkflow(spec.Mode, spec.StartStep)
	run := state.Run{
		ID:          spec.RunID,
		Spec:        spec,
		Status:      state.RunStatusQueued,
		CurrentStep: state.WorkflowStepUnknown,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if run.ID == "" {
		run.ID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	if c.store != nil {
		if err := c.store.UpsertRun(ctx, run); err != nil {
			return state.Run{}, err
		}
	}
	_ = history.Append(spec.ArtifactsRoot, history.New(run.ID, history.KindRunStarted))

	step, ok := workflow.Next(state.WorkflowStepUnknown)
	if !ok {
		run.Status = state.RunStatusCompleted
		return run, nil
	}

	run.Status = state.RunStatusRunning
	run.CurrentStep = step
	run.Spec.StartStep = step
	if c.store != nil {
		if err := c.store.UpsertRun(ctx, run); err != nil {
			return state.Run{}, err
		}
	}

	for {
		if _, err := c.orchestrator.Negotiate(step); err != nil {
			run.Status = state.RunStatusBlocked
			c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
			return run, err
		}
		if spec.ArtifactsRoot != "" {
			_ = history.Append(spec.ArtifactsRoot, func() history.Event {
				ev := history.New(run.ID, history.KindStepStarted)
				ev.Step = string(step)
				ev.TaskID = fmt.Sprintf("%s task", step)
				return ev
			}())
		}
		if spec.ArtifactsRoot != "" {
			_ = debate.WriteRecord(spec.ArtifactsRoot, debate.New(run.ID, string(step), spec.Goal))
		}

		task := state.TaskSpec{
			RunID:                run.ID,
			Step:                 step,
			Title:                fmt.Sprintf("%s task", step),
			Prompt:               spec.Goal,
			RequiresApprovalTier: c.orchestrator.policy.StepApprovalOrDefault(step),
		}

		approvalTier := c.orchestrator.policy.StepApprovalOrDefault(step)
		if approvalTier != state.ApprovalAuto {
			req := state.ApprovalRequest{
				RunID:  run.ID,
				TaskID: task.Title,
				Step:   step,
				Tier:   approvalTier,
				Reason: fmt.Sprintf("approval required for %s", step),
			}
			decision, err := c.orchestrator.adapter.RequestApproval(req)
			if err != nil {
				run.Status = state.RunStatusBlocked
				c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
				return run, err
			}
			_ = history.Append(spec.ArtifactsRoot, func() history.Event {
				ev := history.New(run.ID, history.KindApprovalRequest)
				ev.Step = string(step)
				ev.TaskID = task.Title
				ev.Message = req.Reason
				return ev
			}())
			if !decision.Approved {
				_ = history.Append(spec.ArtifactsRoot, func() history.Event {
					ev := history.New(run.ID, history.KindApprovalResult)
					ev.Step = string(step)
					ev.TaskID = task.Title
					ev.Message = decision.Memo
					return ev
				}())
				run.Status = state.RunStatusBlocked
				err := fmt.Errorf("approval denied for step %s", step)
				c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
				return run, err
			}
			_ = history.Append(spec.ArtifactsRoot, func() history.Event {
				ev := history.New(run.ID, history.KindApprovalResult)
				ev.Step = string(step)
				ev.TaskID = task.Title
				ev.Message = decision.Memo
				return ev
			}())
			if err := c.orchestrator.adapter.ApplyApproval(decision); err != nil {
				run.Status = state.RunStatusBlocked
				c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
				return run, err
			}
		}

		if c.sandbox != nil {
			session, err := c.sandbox.Open(ctx, spec)
			if err != nil {
				run.Status = state.RunStatusFailed
				c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
				return run, err
			}
			_ = session.Close()
		}

		result, err := c.orchestrator.adapter.ExecuteTask(task)
		if err != nil {
			run.Status = state.RunStatusFailed
			c.recordFailure(spec.ArtifactsRoot, run.ID, step, err)
			return run, err
		}

		run.TaskIDs = append(run.TaskIDs, task.Title)
		run.AttemptIDs = append(run.AttemptIDs, result.AttemptID)
		if result.CheckpointID != "" {
			run.CheckpointIDs = append(run.CheckpointIDs, result.CheckpointID)
		}
		for _, artifact := range result.Artifacts {
			if artifact.ID != "" {
				run.ArtifactIDs = append(run.ArtifactIDs, artifact.ID)
			}
		}
		run.CurrentStep = step
		if c.store != nil {
			if err := c.store.UpsertRun(ctx, run); err != nil {
				return state.Run{}, err
			}
		}

		next, hasNext := workflow.Next(step)
		if hasNext {
			c.recordStep(spec.ArtifactsRoot, run.ID, step, next, result.CheckpointID, false, "")
			_ = history.Append(spec.ArtifactsRoot, func() history.Event {
				ev := history.New(run.ID, history.KindStepCompleted)
				ev.Step = string(step)
				ev.TaskID = task.Title
				ev.Checkpoint = result.CheckpointID
				return ev
			}())
		} else {
			c.recordStep(spec.ArtifactsRoot, run.ID, step, state.WorkflowStepUnknown, result.CheckpointID, true, "")
			_ = history.Append(spec.ArtifactsRoot, func() history.Event {
				ev := history.New(run.ID, history.KindStepCompleted)
				ev.Step = string(step)
				ev.TaskID = task.Title
				ev.Checkpoint = result.CheckpointID
				return ev
			}())
		}
		if !hasNext {
			run.Status = state.RunStatusCompleted
			if c.store != nil {
				if err := c.store.UpsertRun(ctx, run); err != nil {
					return state.Run{}, err
				}
			}
			return run, nil
		}
		step = next
	}
}

func (c *Coordinator) recordStep(root, runID string, step, next state.WorkflowStep, checkpointID string, done bool, failure string) {
	if root == "" || runID == "" {
		return
	}
	record, err := progress.LoadRecord(root, runID)
	if err != nil {
		record = progress.New(runID)
	}
	if failure != "" {
		record.State = progress.StateBlocked
		record.Failed = append(record.Failed, progress.Failure{Step: string(step), Reason: failure})
		_ = history.Append(root, func() history.Event {
			ev := history.New(runID, history.KindFailure)
			ev.Step = string(step)
			ev.Message = failure
			return ev
		}())
	} else if done {
		record.State = progress.StateDone
		record.Done = append(record.Done, string(step))
	} else {
		record.State = progress.StateRunning
		record.Done = append(record.Done, string(step))
	}
	record.Resume.NextStep = string(next)
	if record.Checkpoint == nil {
		record.Checkpoint = &progress.Checkpoint{}
	}
	record.Checkpoint.Step = string(step)
	record.Checkpoint.Cursor = string(next)
	record.Checkpoint.Snapshot = checkpointID
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	_ = progress.WriteRecord(root, record)
	if checkpointID != "" {
		_ = history.Append(root, func() history.Event {
			ev := history.New(runID, history.KindCheckpoint)
			ev.Step = string(step)
			ev.Checkpoint = checkpointID
			return ev
		}())
	}
}

func (c *Coordinator) recordFailure(root, runID string, step state.WorkflowStep, err error) {
	if err == nil {
		return
	}
	c.recordStep(root, runID, step, state.WorkflowStepUnknown, "", false, err.Error())
}

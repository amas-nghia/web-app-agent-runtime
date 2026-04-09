package history

import "time"

type Kind string

const (
	KindRunStarted      Kind = "run_started"
	KindStepStarted     Kind = "step_started"
	KindStepCompleted   Kind = "step_completed"
	KindApprovalRequest Kind = "approval_requested"
	KindApprovalResult  Kind = "approval_result"
	KindFailure         Kind = "failure"
	KindCheckpoint      Kind = "checkpoint"
)

type Event struct {
	RunID      string `json:"run_id"`
	Kind       Kind   `json:"kind"`
	Step       string `json:"step,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	Checkpoint string `json:"checkpoint,omitempty"`
	Message    string `json:"message,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func New(runID string, kind Kind) Event {
	return Event{RunID: runID, Kind: kind, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
}

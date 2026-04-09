package progress

import "time"

type State string

const (
	StateRunning State = "running"
	StateBlocked State = "blocked"
	StateDone    State = "done"
)

type Failure struct {
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Decision struct {
	Value string `json:"value"`
	Why   string `json:"why"`
}

type Resume struct {
	NextStep string `json:"next_step"`
	Cursor   string `json:"cursor"`
	Notes    string `json:"notes"`
}

type Checkpoint struct {
	Step     string `json:"step"`
	Cursor   string `json:"cursor"`
	Snapshot string `json:"snapshot"`
}

type Record struct {
	RunID      string      `json:"run_id"`
	UpdatedAt  string      `json:"updated_at"`
	State      State       `json:"state"`
	Done       []string    `json:"done"`
	Failed     []Failure   `json:"failed"`
	Decision   *Decision   `json:"decision"`
	Resume     Resume      `json:"resume"`
	Checkpoint *Checkpoint `json:"checkpoint,omitempty"`
}

func New(runID string) Record {
	return Record{RunID: runID, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano), State: StateRunning, Done: []string{}, Failed: []Failure{}, Resume: Resume{}, Checkpoint: &Checkpoint{}}
}

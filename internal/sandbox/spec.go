package sandbox

import (
	"time"

	"web-app-agent-runtime/internal/state"
)

// Spec is the minimal policy-shaped input for a sandbox backend.
type Spec struct {
	RunID      string
	RepoPath   string
	BaseBranch string
	Mode       state.RunMode
	ReadOnly   bool
	Network    bool
	Worktree   string
	Timeout    time.Duration
}

func FromRunSpec(run state.RunSpec) Spec {
	return Spec{
		RunID:      run.RunID,
		RepoPath:   run.RepoPath,
		BaseBranch: run.BaseBranch,
		Mode:       run.Mode,
		ReadOnly:   false,
		Network:    false,
		Worktree:   "",
		Timeout:    0,
	}
}

package state

import (
	"context"
	"errors"
)

var ErrRunNotFound = errors.New("state: run not found")

// RunStore persists canonical run state.
//
// The interface stays narrow so SQLite can back early tests now and PostgreSQL
// can replace the implementation later without changing callers.
type RunStore interface {
	UpsertRun(ctx context.Context, run Run) error
	GetRun(ctx context.Context, id string) (Run, error)
	ListRuns(ctx context.Context) ([]Run, error)
}

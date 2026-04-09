package sandbox

import (
	"context"
	"errors"
	"sync"

	"web-app-agent-runtime/internal/state"
)

// ErrClosed is returned when a session is used after Close.
var ErrClosed = errors.New("sandbox: session closed")

// Boundary owns the execution isolation boundary for a run.
//
// The MVP ships with a safe in-process default so tests can stay lightweight.
// Container/VM implementations can swap in later without changing callers.
type Boundary interface {
	Name() string
	Capabilities() state.CapabilitySet
	Open(context.Context, state.RunSpec) (Session, error)
}

// Session represents one opened sandbox boundary for a run.
type Session interface {
	ID() string
	Spec() state.RunSpec
	Execute(context.Context, state.TaskSpec) (state.ExecutionResult, error)
	Close() error
}

type noopBoundary struct{}

// New returns the safe default sandbox boundary.
func New() Boundary { return &noopBoundary{} }

func (noopBoundary) Name() string { return "noop" }

func (noopBoundary) Capabilities() state.CapabilitySet {
	return state.CapabilitySet{
		Read:         true,
		Write:        true,
		Resume:       true,
		ApprovalMemo: true,
	}
}

func (noopBoundary) Open(ctx context.Context, spec state.RunSpec) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &noopSession{spec: spec}, nil
}

type noopSession struct {
	mu     sync.Mutex
	closed bool
	spec   state.RunSpec
}

func (s *noopSession) ID() string { return "noop" }

func (s *noopSession) Spec() state.RunSpec { return s.spec }

func (s *noopSession) Execute(ctx context.Context, task state.TaskSpec) (state.ExecutionResult, error) {
	if err := ctx.Err(); err != nil {
		return state.ExecutionResult{}, err
	}

	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return state.ExecutionResult{}, ErrClosed
	}

	return state.ExecutionResult{
		RunID:   task.RunID,
		TaskID:  task.Title,
		Summary: "sandbox:no-op",
	}, nil
}

func (s *noopSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

var _ Boundary = (*noopBoundary)(nil)
var _ Session = (*noopSession)(nil)

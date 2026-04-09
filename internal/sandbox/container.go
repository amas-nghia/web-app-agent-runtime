package sandbox

import (
	"context"
	"fmt"
	"sync"

	"web-app-agent-runtime/internal/state"
)

type containerBoundary struct {
	runtime Runtime
}

// NewContainer returns a container-first boundary wrapper.
func NewContainer(runtime Runtime) Boundary {
	if runtime == nil {
		runtime = noopRuntime{}
	}
	return &containerBoundary{runtime: runtime}
}

func (b *containerBoundary) Name() string { return "container" }

func (b *containerBoundary) Capabilities() state.CapabilitySet {
	return state.CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true}
}

func (b *containerBoundary) Open(ctx context.Context, spec state.RunSpec) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	handle, err := b.runtime.Start(ctx, FromRunSpec(spec))
	if err != nil {
		return nil, err
	}
	return &containerSession{handle: handle, spec: spec}, nil
}

type containerSession struct {
	handle Handle
	spec   state.RunSpec
}

func (s *containerSession) ID() string { return s.handle.ID() }

func (s *containerSession) Spec() state.RunSpec { return s.spec }

func (s *containerSession) Execute(ctx context.Context, task state.TaskSpec) (state.ExecutionResult, error) {
	return s.handle.Execute(ctx, task)
}

func (s *containerSession) Close() error { return s.handle.Close() }

type noopRuntime struct{}

func (noopRuntime) Start(ctx context.Context, spec Spec) (Handle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &noopHandle{id: fmt.Sprintf("container-%s", spec.RunID), spec: spec}, nil
}

type noopHandle struct {
	id   string
	spec Spec
	mu   sync.Mutex
	open bool
}

func (h *noopHandle) ID() string { return h.id }

func (h *noopHandle) Execute(ctx context.Context, task state.TaskSpec) (state.ExecutionResult, error) {
	if err := ctx.Err(); err != nil {
		return state.ExecutionResult{}, err
	}
	return state.ExecutionResult{RunID: task.RunID, TaskID: task.Title, Summary: "container:no-op"}, nil
}

func (h *noopHandle) Close() error { return nil }

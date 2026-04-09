package sandbox

import (
	"context"
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestNew_NoopBoundary(t *testing.T) {
	boundary := New()

	if got := boundary.Name(); got != "noop" {
		t.Fatalf("Name() = %q, want %q", got, "noop")
	}

	if got := boundary.Capabilities(); got.Shell {
		t.Fatalf("Capabilities().Shell = true, want false")
	}

	session, err := boundary.Open(context.Background(), state.RunSpec{Goal: "build a form"})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if got := session.Spec().Goal; got != "build a form" {
		t.Fatalf("Spec().Goal = %q, want %q", got, "build a form")
	}

	result, err := session.Execute(context.Background(), state.TaskSpec{RunID: "run-1", Title: "plan", Step: state.WorkflowStepDebate})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.RunID != "run-1" || result.TaskID != "plan" {
		t.Fatalf("Execute() result = %#v, want run-1/plan", result)
	}
	if result.Summary != "sandbox:no-op" {
		t.Fatalf("Execute() Summary = %q, want %q", result.Summary, "sandbox:no-op")
	}

	if err := session.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if _, err := session.Execute(context.Background(), state.TaskSpec{RunID: "run-1", Title: "plan"}); err != ErrClosed {
		t.Fatalf("Execute() after Close() error = %v, want %v", err, ErrClosed)
	}
}

func TestOpenHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := New().Open(ctx, state.RunSpec{}); err != context.Canceled {
		t.Fatalf("Open() error = %v, want %v", err, context.Canceled)
	}
}

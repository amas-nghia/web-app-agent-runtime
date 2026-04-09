package sandbox

import (
	"context"
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestNewContainer_NoopFallback(t *testing.T) {
	t.Parallel()

	b := NewContainer(nil)
	session, err := b.Open(context.Background(), state.RunSpec{RunID: "run-1", Goal: "build app"})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	result, err := session.Execute(context.Background(), state.TaskSpec{RunID: "run-1", Title: "plan"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Summary != "container:no-op" {
		t.Fatalf("Execute() Summary = %q, want %q", result.Summary, "container:no-op")
	}
}

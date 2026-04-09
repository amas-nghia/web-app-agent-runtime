package history

import "testing"

func TestAppendLoadLast(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := Append(root, New("run-1", KindRunStarted)); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := Append(root, New("run-1", KindStepCompleted)); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	events, err := Load(root, "run-1")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Load() len = %d, want 2", len(events))
	}
	last, ok := Last(events)
	if !ok || last.Kind != KindStepCompleted {
		t.Fatalf("Last() = %#v, %v", last, ok)
	}
}

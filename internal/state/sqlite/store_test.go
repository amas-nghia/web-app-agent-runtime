package sqlite

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"web-app-agent-runtime/internal/state"
)

func TestStore_RunRoundTrip(t *testing.T) {
	t.Parallel()

	store := openTestStore(t)
	ctx := context.Background()

	original := state.Run{
		ID:            "run-1",
		Spec:          state.RunSpec{Goal: "build a web app", Mode: state.RunModeIntake, RepoPath: "/tmp/repo", BaseBranch: "main", BugReport: "crash on load", ResumeFromCheckpoint: "checkpoint-1"},
		Status:        state.RunStatusRunning,
		CurrentStep:   state.WorkflowStepBuild,
		TaskIDs:       []string{"task-1", "task-2"},
		AttemptIDs:    []string{"attempt-1"},
		ApprovalIDs:   []string{"approval-1"},
		CheckpointIDs: []string{"checkpoint-1"},
		ArtifactIDs:   []string{"artifact-1"},
		CreatedAt:     time.Unix(100, 0).UTC(),
		UpdatedAt:     time.Unix(200, 0).UTC(),
	}

	if err := store.UpsertRun(ctx, original); err != nil {
		t.Fatalf("UpsertRun() error = %v", err)
	}

	got, err := store.GetRun(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}

	if !reflect.DeepEqual(got, original) {
		t.Fatalf("GetRun() = %#v, want %#v", got, original)
	}
}

func TestStore_GetRunNotFound(t *testing.T) {
	t.Parallel()

	store := openTestStore(t)

	_, err := store.GetRun(context.Background(), "missing")
	if err != state.ErrRunNotFound {
		t.Fatalf("GetRun() error = %v, want %v", err, state.ErrRunNotFound)
	}
}

func TestStore_ListRuns(t *testing.T) {
	t.Parallel()

	store := openTestStore(t)
	ctx := context.Background()

	runs := []state.Run{
		{ID: "run-b", Spec: state.RunSpec{Goal: "b"}, Status: state.RunStatusQueued, CurrentStep: state.WorkflowStepIntake, CreatedAt: time.Unix(2, 0).UTC(), UpdatedAt: time.Unix(2, 0).UTC()},
		{ID: "run-a", Spec: state.RunSpec{Goal: "a"}, Status: state.RunStatusQueued, CurrentStep: state.WorkflowStepIntake, CreatedAt: time.Unix(1, 0).UTC(), UpdatedAt: time.Unix(1, 0).UTC()},
	}

	for _, run := range runs {
		if err := store.UpsertRun(ctx, run); err != nil {
			t.Fatalf("UpsertRun(%s) error = %v", run.ID, err)
		}
	}

	got, err := store.ListRuns(ctx)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}

	if len(got) != len(runs) {
		t.Fatalf("ListRuns() len = %d, want %d", len(got), len(runs))
	}
	if got[0].ID != "run-a" || got[1].ID != "run-b" {
		t.Fatalf("ListRuns() order = %#v, want run-a then run-b", got)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()

	path := filepath.Join(t.TempDir(), "state.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return store
}

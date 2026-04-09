package opencode

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestExecuteTask(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		wantErr           string
		wantErrIsNotFound bool
		wantSummary       []string
	}{
		{
			name: "success writes stdout and stderr from PATH executable",
			script: `#!/bin/sh
printf 'stdout:%s\n' "$3"
printf 'stderr:%s\n' "$3" >&2
exit 0
`,
			wantSummary: []string{"stdout:draft the plan", "stderr:draft the plan"},
		},
		{
			name: "failure returns wrapped exit error and preserves output",
			script: `#!/bin/sh
printf 'stdout:%s\n' "$3"
printf 'stderr:%s\n' "$3" >&2
exit 7
`,
			wantErr:     "exit status 7",
			wantSummary: []string{"stdout:draft the plan", "stderr:draft the plan"},
		},
		{
			name:              "missing executable on path",
			wantErrIsNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			adapter := New()
			task := state.TaskSpec{RunID: "run-123", Title: "plan", Prompt: "draft the plan"}
			if tt.script != "" {
				path := writeOpencodeExecutable(t, tt.script)
				setPathForExecutable(t, filepath.Dir(path))
			} else {
				t.Setenv("PATH", t.TempDir())
			}

			// Act.
			got, err := adapter.ExecuteTask(task)

			// Assert.
			if tt.wantErrIsNotFound {
				if !errors.Is(err, exec.ErrNotFound) {
					t.Fatalf("ExecuteTask() error = %v, want exec.ErrNotFound", err)
				}
				if got.Summary != "" {
					t.Fatalf("ExecuteTask().Summary = %q, want empty", got.Summary)
				}
				if got.Error == "" {
					t.Fatal("ExecuteTask().Error = empty, want a captured error string")
				}
				return
			}

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ExecuteTask() error = %v, want contains %q", err, tt.wantErr)
				}
				if got.Error == "" {
					t.Fatal("ExecuteTask().Error = empty, want captured command error")
				}
			} else if err != nil {
				t.Fatalf("ExecuteTask() error = %v, want nil", err)
			}

			for _, want := range tt.wantSummary {
				if !strings.Contains(got.Summary, want) {
					t.Fatalf("ExecuteTask().Summary = %q, want contains %q", got.Summary, want)
				}
			}
			if got.RunID != task.RunID {
				t.Fatalf("ExecuteTask().RunID = %q, want %q", got.RunID, task.RunID)
			}
			if got.TaskID != task.Title {
				t.Fatalf("ExecuteTask().TaskID = %q, want %q", got.TaskID, task.Title)
			}
			if got.AttemptID != task.Title+"-attempt" {
				t.Fatalf("ExecuteTask().AttemptID = %q, want %q", got.AttemptID, task.Title+"-attempt")
			}
			if got.CheckpointID != task.Title+"-checkpoint" {
				t.Fatalf("ExecuteTask().CheckpointID = %q, want %q", got.CheckpointID, task.Title+"-checkpoint")
			}
			if len(got.Artifacts) != 1 {
				t.Fatalf("ExecuteTask().Artifacts len = %d, want 1", len(got.Artifacts))
			}
			if got.Artifacts[0].Path != "opencode-output" {
				t.Fatalf("ExecuteTask().Artifacts[0].Path = %q, want %q", got.Artifacts[0].Path, "opencode-output")
			}
			if tt.wantErr == "" && got.Error != "" {
				t.Fatalf("ExecuteTask().Error = %q, want empty", got.Error)
			}
		})
	}
}

func writeOpencodeExecutable(t *testing.T, script string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "opencode")
	content := []byte(script)
	if runtime.GOOS == "windows" {
		content = []byte(strings.ReplaceAll(script, "\n", "\r\n"))
	}
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("os.Chmod(%q) error = %v", path, err)
	}
	return path
}

func setPathForExecutable(t *testing.T, dir string) {
	t.Helper()

	original := os.Getenv("PATH")
	if original == "" {
		t.Setenv("PATH", dir)
		return
	}
	t.Setenv("PATH", fmt.Sprintf("%s%c%s", dir, os.PathListSeparator, original))
}

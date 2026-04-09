package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"web-app-agent-runtime/internal/state"
	statesqlite "web-app-agent-runtime/internal/state/sqlite"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantRC  int
	}{
		{
			name:    "version command",
			args:    []string{"version"},
			wantOut: "web-app-agent-runtime dev\n",
			wantRC:  0,
		},
		{
			name:    "default smoke output",
			args:    nil,
			wantOut: "web-app-agent-runtime orchestrator skeleton\n",
			wantRC:  0,
		},
		{
			name:    "run command smoke",
			args:    []string{"run", "--state", "__STATE__", "build a responsive web app"},
			wantRC:  0,
			wantOut: "completed release\n",
		},
		{
			name:    "bugfix command smoke",
			args:    []string{"bugfix", "--run-id", "bugfix-smoke", "--state", "__STATE__", "--bug-report", "issue-42", "fix the checkout flow"},
			wantRC:  0,
			wantOut: "completed release\n",
		},
		{
			name:    "bugfix command requires a goal",
			args:    []string{"bugfix", "--state", "__STATE__", "--bug-report", "issue-42"},
			wantRC:  2,
			wantOut: "bugfix requires a goal\n",
		},
		{
			name:    "bugfix command requires a bug report",
			args:    []string{"bugfix", "--state", "__STATE__", "fix the checkout flow"},
			wantRC:  2,
			wantOut: "bugfix requires a bug report\n",
		},
		{
			name:    "bugfix command rejects mode override",
			args:    []string{"bugfix", "--state", "__STATE__", "--mode", "resume", "--bug-report", "issue-42", "fix the checkout flow"},
			wantRC:  2,
			wantOut: "flag provided but not defined: -mode\n",
		},
		{
			name:    "status command with no run",
			args:    []string{"status", "--root", t.TempDir()},
			wantRC:  0,
			wantOut: "no current run\n",
		},
		{
			name:    "resume command when no run exists",
			args:    []string{"resume", "--root", t.TempDir()},
			wantRC:  2,
			wantOut: "no current run to resume\n",
		},
		{
			name:    "replay command when no run exists",
			args:    []string{"replay", "--root", t.TempDir()},
			wantRC:  2,
			wantOut: "no current run to replay\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			args := append([]string(nil), tt.args...)
			var statePath string
			needsStatePath := false
			for _, arg := range args {
				if arg == "__STATE__" {
					needsStatePath = true
					break
				}
			}
			if needsStatePath {
				statePath = filepath.Join(t.TempDir(), "state.db")
				for i, arg := range args {
					if arg == "__STATE__" {
						args[i] = statePath
					}
				}
			}
			var out bytes.Buffer

			// Act.
			gotRC := run(args, &out)

			// Assert.
			if gotRC != tt.wantRC {
				t.Fatalf("run() rc = %d, want %d", gotRC, tt.wantRC)
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				if tt.name == "run command smoke" || tt.name == "bugfix command smoke" {
					if !strings.Contains(gotOut, tt.wantOut) {
						t.Fatalf("run() output = %q, want contains %q", gotOut, tt.wantOut)
					}
				} else {
					t.Fatalf("run() output = %q, want %q", gotOut, tt.wantOut)
				}
			}

			if tt.name == "bugfix command smoke" {
				store, err := statesqlite.Open(statePath)
				if err != nil {
					t.Fatalf("Open() error = %v", err)
				}
				defer store.Close()

				run, err := store.GetRun(context.Background(), "bugfix-smoke")
				if err != nil {
					t.Fatalf("GetRun() error = %v", err)
				}
				if run.Spec.Mode != state.RunModeBugfix {
					t.Fatalf("run.Spec.Mode = %q, want %q", run.Spec.Mode, state.RunModeBugfix)
				}
				if run.Spec.BugReport != "issue-42" {
					t.Fatalf("run.Spec.BugReport = %q, want %q", run.Spec.BugReport, "issue-42")
				}
			}
		})
	}
}

func TestRunStatusShowsCompactHistorySummary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	statePath := filepath.Join(root, "state.db")
	var out bytes.Buffer
	if rc := run([]string{"run", "--state", statePath, "build a responsive web app"}, &out); rc != 0 {
		t.Fatalf("seed run rc = %d, want 0; output=%q", rc, out.String())
	}

	out.Reset()
	if rc := run([]string{"status", "--root", root, "--state", statePath}, &out); rc != 0 {
		t.Fatalf("status rc = %d, want 0; output=%q", rc, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "history: ") {
		t.Fatalf("status output = %q, want compact history summary", got)
	}
	if strings.Contains(got, "done: [") || strings.Contains(got, "failed: [") {
		t.Fatalf("status output = %q, want no raw slice dumps", got)
	}
}

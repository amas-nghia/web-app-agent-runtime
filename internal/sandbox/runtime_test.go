package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestNewRuntimeFromEnv_SelectionAndStartup(t *testing.T) {
	tests := []struct {
		name            string
		sandboxEnv      string
		setupPath       func(*testing.T) string
		wantRuntimeKind string
		wantSessionID   string
		wantSummary     string
		wantErrContains string
	}{
		{
			name:            "default env returns noop runtime and no-op boundary",
			setupPath:       func(t *testing.T) string { return t.TempDir() },
			wantRuntimeKind: "noop",
			wantSessionID:   "container-run-1",
			wantSummary:     "container:no-op",
		},
		{
			name:       "docker env returns docker runtime and starts with stub docker binary",
			sandboxEnv: "docker",
			setupPath: func(t *testing.T) string {
				return writeExecutable(t, "docker", `#!/bin/sh
printf 'fake-container-id\n'
exit 0
`)
			},
			wantRuntimeKind: "docker",
			wantSessionID:   "fake-container-id",
		},
		{
			name:            "docker env fails clearly when docker is missing",
			sandboxEnv:      "docker",
			setupPath:       func(t *testing.T) string { return t.TempDir() },
			wantRuntimeKind: "docker",
			wantErrContains: "docker binary not found",
		},
		{
			name:            "vm env fails clearly when qemu is missing",
			sandboxEnv:      "vm",
			setupPath:       func(t *testing.T) string { return t.TempDir() },
			wantRuntimeKind: "vm",
			wantErrContains: "qemu-system-x86_64 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			if tt.sandboxEnv != "" {
				t.Setenv("WABR_SANDBOX", tt.sandboxEnv)
			} else {
				t.Setenv("WABR_SANDBOX", "")
			}
			if tt.setupPath != nil {
				t.Setenv("PATH", tt.setupPath(t))
			}
			runtime := NewRuntimeFromEnv()
			boundary := NewContainer(runtime)

			// Act.
			session, err := boundary.Open(context.Background(), state.RunSpec{RunID: "run-1", Goal: "build app"})

			// Assert.
			assertRuntimeKind(t, runtime, tt.wantRuntimeKind)

			if tt.wantErrContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("Open() error = %v, want contains %q", err, tt.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("Open() error = %v, want nil", err)
			}

			if got := session.ID(); got != tt.wantSessionID {
				t.Fatalf("Session.ID() = %q, want %q", got, tt.wantSessionID)
			}

			if tt.wantSummary != "" {
				result, err := session.Execute(context.Background(), state.TaskSpec{RunID: "run-1", Title: "plan"})
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				if got := result.Summary; got != tt.wantSummary {
					t.Fatalf("Execute().Summary = %q, want %q", got, tt.wantSummary)
				}
			}
		})
	}
}

func assertRuntimeKind(t *testing.T, runtime Runtime, want string) {
	t.Helper()
	if runtime == nil {
		t.Fatal("runtime = nil, want concrete runtime")
	}

	switch want {
	case "noop":
		if _, ok := runtime.(noopRuntime); !ok {
			t.Fatalf("runtime type = %T, want noopRuntime", runtime)
		}
		if _, ok := runtime.(*dockerRuntime); ok {
			t.Fatalf("runtime type = %T, want not dockerRuntime", runtime)
		}
		if _, ok := runtime.(vmRuntime); ok {
			t.Fatalf("runtime type = %T, want not vmRuntime", runtime)
		}
	case "docker":
		if _, ok := runtime.(*dockerRuntime); !ok {
			t.Fatalf("runtime type = %T, want *dockerRuntime", runtime)
		}
		if _, ok := runtime.(noopRuntime); ok {
			t.Fatalf("runtime type = %T, want not noopRuntime", runtime)
		}
		if _, ok := runtime.(vmRuntime); ok {
			t.Fatalf("runtime type = %T, want not vmRuntime", runtime)
		}
	case "vm":
		if _, ok := runtime.(vmRuntime); !ok {
			t.Fatalf("runtime type = %T, want vmRuntime", runtime)
		}
		if _, ok := runtime.(noopRuntime); ok {
			t.Fatalf("runtime type = %T, want not noopRuntime", runtime)
		}
		if _, ok := runtime.(*dockerRuntime); ok {
			t.Fatalf("runtime type = %T, want not dockerRuntime", runtime)
		}
	default:
		t.Fatalf("unknown want runtime kind %q", want)
	}
}

func writeExecutable(t *testing.T, name, script string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("os.Chmod(%q) error = %v", path, err)
	}
	return dir
}

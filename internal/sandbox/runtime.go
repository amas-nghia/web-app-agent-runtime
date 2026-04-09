package sandbox

import (
	"context"
	"os"

	"web-app-agent-runtime/internal/state"
)

// Runtime starts a backend-specific execution boundary.
type Runtime interface {
	Start(context.Context, Spec) (Handle, error)
}

// Handle executes tasks inside the sandbox backend.
type Handle interface {
	ID() string
	Execute(context.Context, state.TaskSpec) (state.ExecutionResult, error)
	Close() error
}

func NewRuntimeFromEnv() Runtime {
	switch os.Getenv("WABR_SANDBOX") {
	case "docker":
		return newDockerRuntime()
	case "vm":
		return newVMRuntime()
	default:
		return noopRuntime{}
	}
}

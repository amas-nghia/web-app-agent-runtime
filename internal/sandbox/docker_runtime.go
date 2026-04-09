package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"web-app-agent-runtime/internal/state"
)

type dockerRuntime struct {
	image string
}

func newDockerRuntime() Runtime {
	image := os.Getenv("WABR_SANDBOX_IMAGE")
	if image == "" {
		image = "alpine:3.20"
	}
	return &dockerRuntime{image: image}
}

func (r *dockerRuntime) Start(ctx context.Context, spec Spec) (Handle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker runtime: docker binary not found: %w", err)
	}

	name := sanitizeContainerName(spec.RunID)
	worktree := spec.Worktree
	if worktree == "" {
		worktree = spec.RepoPath
	}
	if worktree == "" {
		worktree = os.TempDir()
	}

	args := []string{"run", "-d", "--rm", "--name", name}
	if worktree != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/work", worktree), "-w", "/work")
	}
	if !spec.Network {
		args = append(args, "--network", "none")
	}
	args = append(args, r.image, "sh", "-lc", "trap : TERM INT; sleep infinity & wait")

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("docker runtime start: %s: %w", msg, err)
	}

	id := strings.TrimSpace(stdout.String())
	if id == "" {
		return nil, errors.New("docker runtime start: empty container id")
	}

	return &dockerHandle{id: id, image: r.image}, nil
}

type dockerHandle struct {
	id    string
	image string
}

func (h *dockerHandle) ID() string { return h.id }

func (h *dockerHandle) Execute(ctx context.Context, task state.TaskSpec) (state.ExecutionResult, error) {
	if err := ctx.Err(); err != nil {
		return state.ExecutionResult{}, err
	}
	command := "printf '%s\\n' \"$TASK_PROMPT\""
	cmd := exec.CommandContext(ctx, "docker", "exec", "-e", "TASK_PROMPT="+task.Prompt, h.id, "sh", "-lc", command)
	output, err := cmd.CombinedOutput()
	result := state.ExecutionResult{
		RunID:        task.RunID,
		TaskID:       task.Title,
		AttemptID:    task.Title + "-attempt",
		CheckpointID: task.Title + "-checkpoint",
		Summary:      strings.TrimSpace(string(output)),
		Artifacts: []state.ArtifactRef{{
			ID:        task.Title + "-artifact",
			RunID:     task.RunID,
			TaskID:    task.Title,
			AttemptID: task.Title + "-attempt",
			Kind:      state.ArtifactKindSource,
			Path:      filepath.Join("runs", task.RunID, "artifacts", task.Title+".txt"),
			Digest:    "",
			CreatedAt: time.Now().UTC(),
		}},
	}
	if err != nil {
		result.Error = err.Error()
		return result, fmt.Errorf("docker runtime execute task: %w", err)
	}
	return result, nil
}

func (h *dockerHandle) Close() error {
	if h == nil || h.id == "" {
		return nil
	}
	cmd := exec.Command("docker", "stop", h.id)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker runtime close: %w", err)
	}
	return nil
}

func sanitizeContainerName(raw string) string {
	if raw == "" {
		return fmt.Sprintf("wabr-%d", time.Now().UnixNano())
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		return fmt.Sprintf("wabr-%d", time.Now().UnixNano())
	}
	return "wabr-" + name
}

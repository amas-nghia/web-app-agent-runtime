package issuebot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func fixDocs(root string) (bool, error) {
	changed := false

	readmePath := filepath.Join(root, "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return false, fmt.Errorf("read README: %w", err)
	}
	updated := string(readme)
	if !strings.Contains(updated, "## Output contract") {
		updated += "\n## Output contract\n\nThe mock smoke path is a control-flow check, not a full app generator.\n\nAfter a successful mock run, the durable outputs are:\n\n- `runs/current`\n- `runs/<run-id>/progress.json`\n- `runs/<run-id>/history.jsonl`\n- `runs/<run-id>/debates/<step>.json`\n\nThe mock path also records placeholder artifact refs such as `mock-artifact` or `release task-artifact`.\n\n### Tarot demo example\n\nInput goal:\n\n```text\nbuild a beautiful MVC tarot reading web app with frontend, backend, and complete seed data\n```\n\nExpected result on the mock smoke path:\n\n- the run completes successfully\n- workflow metadata is persisted\n- placeholder artifacts are written\n- no runnable tarot app is produced until a real backend adapter is wired\n"
		changed = true
	}
	if !strings.Contains(updated, "Runtime/orchestrator skeleton") {
		updated = strings.Replace(updated, "Runtime for building and maintaining web apps for mobile and desktop.", "Runtime/orchestrator skeleton for building and maintaining web apps for mobile and desktop.", 1)
		changed = true
	}
	if !strings.Contains(updated, "What this is not") {
		updated = strings.Replace(updated, "## Goal", "## What this is\n\n- a workflow runtime that coordinates tasks, approvals, artifacts, and retries\n- a backend-agnostic skeleton that starts with a mock adapter\n- a testable control plane for web app work\n\n## What this is not\n\n- not a full web app generator end-to-end out of the box\n- not a finished product that emits a production app on the mock path\n- not a replacement for a real frontend/backend implementation\n\n## Goal", 1)
		changed = true
	}
	if changed {
		if err := os.WriteFile(readmePath, []byte(updated), 0o644); err != nil {
			return false, fmt.Errorf("write README: %w", err)
		}
	}

	contractsPath := filepath.Join(root, "docs", "contracts.md")
	contracts, err := os.ReadFile(contractsPath)
	if err != nil {
		return changed, fmt.Errorf("read contracts: %w", err)
	}
	contractsUpdated := string(contracts)
	if !strings.Contains(contractsUpdated, "## Output contract") {
		contractsUpdated += "\n## Output contract\n\nThe mock smoke path is a control-flow check, not a full app generator.\n\nAfter a successful mock run, the durable outputs are:\n\n- `runs/current`\n- `runs/<run-id>/progress.json`\n- `runs/<run-id>/history.jsonl`\n- `runs/<run-id>/debates/<step>.json`\n\nThe mock path also records placeholder artifact refs such as `mock-artifact` or `release task-artifact`.\n\n### Tarot demo example\n\nInput goal:\n\n```text\nbuild a beautiful MVC tarot reading web app with frontend, backend, and complete seed data\n```\n\nExpected result on the mock smoke path:\n\n- the run completes successfully\n- workflow metadata is persisted\n- placeholder artifacts are written\n- no runnable tarot app is produced until a real backend adapter is wired\n"
		changed = true
	}
	if changed {
		if err := os.WriteFile(contractsPath, []byte(contractsUpdated), 0o644); err != nil {
			return false, fmt.Errorf("write contracts: %w", err)
		}
	}

	return changed, nil
}

func fixOpenCodeAdapter(root string) (bool, error) {
	path := filepath.Join(root, "internal", "adapters", "opencode", "opencode.go")
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read opencode adapter: %w", err)
	}
	updated := string(content)
	if strings.Contains(updated, "context.WithTimeout") && strings.Contains(updated, "--non-interactive") && strings.Contains(updated, "OPENCODE_NON_INTERACTIVE=1") {
		return false, nil
	}
	if strings.Contains(updated, "cmd := exec.Command(\"opencode\", \"run\", task.Prompt)") {
		updated = strings.Replace(updated, "import (\n\t\"errors\"\n\t\"fmt\"\n\t\"os/exec\"\n\t\"strings\"\n\n\t\"web-app-agent-runtime/internal/state\"\n)", "import (\n\t\"context\"\n\t\"errors\"\n\t\"fmt\"\n\t\"os\"\n\t\"os/exec\"\n\t\"strings\"\n\t\"time\"\n\n\t\"web-app-agent-runtime/internal/state\"\n)", 1)
		updated = strings.Replace(updated, "func (a *Adapter) ExecuteTask(task state.TaskSpec) (state.ExecutionResult, error) {\n\tcmd := exec.Command(\"opencode\", \"run\", task.Prompt)\n\toutput, err := cmd.CombinedOutput()\n", "func (a *Adapter) ExecuteTask(task state.TaskSpec) (state.ExecutionResult, error) {\n\tctx, cancel := context.WithTimeout(context.Background(), opencodeTimeout())\n\tdefer cancel()\n\n\tcmd := exec.CommandContext(ctx, \"opencode\", \"run\", \"--non-interactive\", task.Prompt)\n\tcmd.Stdin = strings.NewReader(\"\")\n\tcmd.Env = append(os.Environ(), \"OPENCODE_NON_INTERACTIVE=1\")\n\toutput, err := cmd.CombinedOutput()\n", 1)
		updated = strings.Replace(updated, "\tif err != nil {\n\t\tresult.Error = err.Error()\n\t\treturn result, fmt.Errorf(\"opencode adapter execute task: %w\", err)\n\t}\n", "\tif err != nil {\n\t\tresult.Error = err.Error()\n\t\tif ctx.Err() == context.DeadlineExceeded {\n\t\t\treturn result, fmt.Errorf(\"opencode adapter execute task: timeout after %s: %w\", opencodeTimeout(), err)\n\t\t}\n\t\treturn result, fmt.Errorf(\"opencode adapter execute task: %w\", err)\n\t}\n", 1)
		updated = strings.Replace(updated, "func (a *Adapter) Cancel(runID string) error { return nil }\n", "func (a *Adapter) Cancel(runID string) error { return nil }\n\nfunc opencodeTimeout() time.Duration {\n\traw := strings.TrimSpace(os.Getenv(\"WABR_OPENCODE_TIMEOUT\"))\n\tif raw == \"\" {\n\t\treturn 2 * time.Minute\n\t}\n\td, err := time.ParseDuration(raw)\n\tif err != nil || d <= 0 {\n\t\treturn 2 * time.Minute\n\t}\n\treturn d\n}\n", 1)
	} else {
		return false, fmt.Errorf("opencode adapter shape not recognized")
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return false, fmt.Errorf("write opencode adapter: %w", err)
	}
	return true, nil
}

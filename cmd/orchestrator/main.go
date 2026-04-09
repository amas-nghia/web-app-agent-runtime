package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"web-app-agent-runtime/internal/adapters"
	"web-app-agent-runtime/internal/history"
	"web-app-agent-runtime/internal/orchestrator"
	"web-app-agent-runtime/internal/policy"
	"web-app-agent-runtime/internal/progress"
	"web-app-agent-runtime/internal/sandbox"
	"web-app-agent-runtime/internal/state"
	statesqlite "web-app-agent-runtime/internal/state/sqlite"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, out io.Writer) int {
	if len(args) > 0 && args[0] == "version" {
		fmt.Fprintln(out, "web-app-agent-runtime dev")
		return 0
	}
	if len(args) > 0 && args[0] == "status" {
		return runStatus(args[1:], out)
	}
	if len(args) > 0 && args[0] == "resume" {
		return runResume(args[1:], out)
	}
	if len(args) > 0 && args[0] == "bugfix" {
		return runBugfix(args[1:], out)
	}
	if len(args) > 0 && args[0] == "replay" {
		return runReplay(args[1:], out)
	}

	if len(args) > 0 && args[0] == "run" {
		return runGoal(args[1:], out)
	}

	fmt.Fprintln(out, "web-app-agent-runtime orchestrator skeleton")
	return 0
}

func runGoal(args []string, out io.Writer) int {
	return runGoalWithFlags(args, out, "run", true, false)
}

func runStatus(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	root := fs.String("root", "runs", "runtime data root")
	statePath := fs.String("state", filepath.Join("runs", "runtime.db"), "state db path")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		return 2
	}

	currentRunID, err := progress.LoadCurrent(*root)
	if err != nil {
		fmt.Fprintln(out, "no current run")
		return 0
	}

	record, err := progress.LoadRecord(*root, currentRunID)
	if err != nil {
		fmt.Fprintf(out, "current run: %s\nprogress unavailable: %v\n", currentRunID, err)
	} else {
		historyEvents, _ := history.Load(*root, currentRunID)
		lastHistory, ok := history.Last(historyEvents)
		historyLine := "history: 0 events"
		if ok {
			historyLine = fmt.Sprintf("history: %d events (last: %s %s)", len(historyEvents), lastHistory.Kind, lastHistory.Step)
		}
		checkpointStep := ""
		checkpointSnapshot := ""
		if record.Checkpoint != nil {
			checkpointStep = record.Checkpoint.Step
			checkpointSnapshot = record.Checkpoint.Snapshot
		}
		fmt.Fprintf(out, "current run: %s\nstate: %s\nprogress: done=%d failed=%d\n%s\nresume: %s\ncheckpoint: %s/%s\n", record.RunID, record.State, len(record.Done), len(record.Failed), historyLine, record.Resume.NextStep, checkpointStep, checkpointSnapshot)
	}

	store, err := statesqlite.Open(*statePath)
	if err == nil {
		defer store.Close()
		if run, err := store.GetRun(context.Background(), currentRunID); err == nil {
			fmt.Fprintf(out, "canonical: %s %s %s\n", run.ID, run.Status, run.CurrentStep)
			fmt.Fprintf(out, "start-step: %s\n", run.Spec.StartStep)
		}
	}

	if _, err := os.Stat(filepath.Join(*root, currentRunID, "debates")); err == nil {
		fmt.Fprintln(out, "debates: present")
	}
	return 0
}

func runResume(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("resume", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	engine := fs.String("engine", "mock", "execution engine")
	statePath := fs.String("state", filepath.Join("runs", "runtime.db"), "state db path")
	root := fs.String("root", "runs", "runtime data root")
	runID := fs.String("run-id", "", "override run id")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		return 2
	}

	currentRunID, err := progress.LoadCurrent(*root)
	if err != nil {
		fmt.Fprintln(out, "no current run to resume")
		return 2
	}
	if *runID != "" {
		currentRunID = *runID
	}

	store, err := statesqlite.Open(*statePath)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	defer store.Close()

	run, err := store.GetRun(context.Background(), currentRunID)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	progressRecord, _ := progress.LoadRecord(*root, currentRunID)
	resumeStartStep := run.CurrentStep
	if progressRecord.Checkpoint != nil && progressRecord.Checkpoint.Step != "" {
		resumeStartStep = parseWorkflowStep(progressRecord.Checkpoint.Step)
	}
	if resumeStartStep == state.WorkflowStepUnknown {
		resumeStartStep = run.Spec.StartStep
	}

	resumeArgs := []string{
		"--engine", *engine,
		"--state", *statePath,
		"--mode", string(run.Spec.Mode),
		"--start-step", string(resumeStartStep),
	}
	if run.Spec.RepoPath != "" {
		resumeArgs = append(resumeArgs, "--repo-path", run.Spec.RepoPath)
	}
	if run.Spec.BaseBranch != "" {
		resumeArgs = append(resumeArgs, "--base-branch", run.Spec.BaseBranch)
	}
	if run.Spec.BugReport != "" {
		resumeArgs = append(resumeArgs, "--bug-report", run.Spec.BugReport)
	}
	if run.Spec.ResumeFromCheckpoint != "" {
		resumeArgs = append(resumeArgs, "--resume-from-checkpoint", run.Spec.ResumeFromCheckpoint)
	}
	resumeArgs = append(resumeArgs, run.Spec.Goal)
	return runGoal(resumeArgs, out)
}

func runBugfix(args []string, out io.Writer) int {
	return runGoalWithFlags(args, out, "bugfix", false, true)
}

func runGoalWithFlags(args []string, out io.Writer, commandName string, allowMode bool, requireBugReport bool) int {
	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	engine := fs.String("engine", "mock", "execution engine")
	statePath := fs.String("state", filepath.Join("runs", "runtime.db"), "state db path")
	runID := fs.String("run-id", "", "run id")
	startStep := fs.String("start-step", "", "start step")
	repoPath := fs.String("repo-path", "", "repo path")
	baseBranch := fs.String("base-branch", "main", "base branch")
	bugReport := fs.String("bug-report", "", "bug report")
	resumeFrom := fs.String("resume-from-checkpoint", "", "resume checkpoint")
	mode := state.RunModeIntake
	var modeFlag *string
	if allowMode {
		modeFlag = fs.String("mode", string(state.RunModeIntake), "run mode")
	}
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		return 2
	}

	goal := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if goal == "" {
		fmt.Fprintln(out, commandName+" requires a goal")
		return 2
	}
	if requireBugReport && strings.TrimSpace(*bugReport) == "" {
		fmt.Fprintln(out, "bugfix requires a bug report")
		return 2
	}
	if allowMode {
		mode = parseRunMode(*modeFlag)
	}
	if !allowMode {
		mode = state.RunModeBugfix
	}

	return executeRun(out, runInvocation{
		engine:     *engine,
		statePath:  *statePath,
		runID:      *runID,
		startStep:  parseWorkflowStep(*startStep),
		repoPath:   *repoPath,
		baseBranch: *baseBranch,
		bugReport:  *bugReport,
		resumeFrom: *resumeFrom,
		goal:       goal,
		mode:       mode,
	})
}

type runInvocation struct {
	engine     string
	statePath  string
	runID      string
	startStep  state.WorkflowStep
	repoPath   string
	baseBranch string
	bugReport  string
	resumeFrom string
	goal       string
	mode       state.RunMode
}

func executeRun(out io.Writer, inv runInvocation) int {
	adapter, err := adapters.New(inv.engine)
	if err != nil {
		fmt.Fprintln(out, err)
		return 2
	}

	if err := os.MkdirAll(filepath.Dir(inv.statePath), 0o755); err != nil {
		fmt.Fprintln(out, err)
		return 1
	}

	store, err := statesqlite.Open(inv.statePath)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	defer store.Close()

	coord := orchestrator.NewCoordinator(adapter, policy.Default(), store, sandbox.NewContainer(sandbox.NewRuntimeFromEnv()))
	progressRoot := filepath.Dir(inv.statePath)
	activeRunID := inv.runID
	if activeRunID == "" {
		activeRunID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	if inv.mode == state.RunModeResume && inv.runID == "" {
		if currentRunID, err := progress.LoadCurrent(progressRoot); err == nil && currentRunID != "" {
			activeRunID = currentRunID
		}
	}
	progressRecord := progress.New(activeRunID)
	if err := progress.WriteCurrent(progressRoot, progressRecord.RunID); err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	if err := progress.WriteRecord(progressRoot, progressRecord); err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	run, err := coord.Execute(context.Background(), state.RunSpec{
		RunID:                activeRunID,
		Goal:                 inv.goal,
		Mode:                 inv.mode,
		StartStep:            inv.startStep,
		RepoPath:             inv.repoPath,
		BaseBranch:           inv.baseBranch,
		BugReport:            inv.bugReport,
		ResumeFromCheckpoint: inv.resumeFrom,
		ArtifactsRoot:        progressRoot,
	})
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}

	_ = progress.WriteCurrent(progressRoot, run.ID)

	fmt.Fprintf(out, "%s %s %s\n", run.ID, run.Status, run.CurrentStep)
	if len(run.ArtifactIDs) > 0 {
		fmt.Fprintf(out, "artifacts: %s\n", strings.Join(run.ArtifactIDs, ","))
	}
	return 0
}

func runReplay(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	engine := fs.String("engine", "mock", "execution engine")
	statePath := fs.String("state", filepath.Join("runs", "runtime.db"), "state db path")
	root := fs.String("root", "runs", "runtime data root")
	runID := fs.String("run-id", "", "override run id")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(out, err)
		return 2
	}

	currentRunID, err := progress.LoadCurrent(*root)
	if err != nil {
		fmt.Fprintln(out, "no current run to replay")
		return 2
	}

	store, err := statesqlite.Open(*statePath)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	defer store.Close()

	run, err := store.GetRun(context.Background(), currentRunID)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}

	progressRecord, err := progress.LoadRecord(*root, currentRunID)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}

	activeRunID := *runID
	if activeRunID == "" {
		activeRunID = fmt.Sprintf("%s-replay-%d", currentRunID, time.Now().UnixNano())
	}

	replayMode := run.Spec.Mode
	replayStartStep := run.Spec.StartStep
	if progressRecord.Checkpoint != nil && progressRecord.Checkpoint.Step != "" {
		replayStartStep = parseWorkflowStep(progressRecord.Checkpoint.Step)
	}

	return runGoal([]string{
		"--engine", *engine,
		"--state", *statePath,
		"--run-id", activeRunID,
		"--mode", string(replayMode),
		"--start-step", string(replayStartStep),
		"--repo-path", run.Spec.RepoPath,
		"--base-branch", run.Spec.BaseBranch,
		"--bug-report", run.Spec.BugReport,
		"--resume-from-checkpoint", checkpointOrFallback(run, progressRecord),
		strings.TrimSpace(run.Spec.Goal),
	}, out)
}

func checkpointOrFallback(run state.Run, record progress.Record) string {
	if record.Checkpoint != nil && record.Checkpoint.Snapshot != "" {
		return record.Checkpoint.Snapshot
	}
	if len(run.CheckpointIDs) > 0 {
		return run.CheckpointIDs[len(run.CheckpointIDs)-1]
	}
	return run.Spec.ResumeFromCheckpoint
}

func parseWorkflowStep(raw string) state.WorkflowStep {
	switch state.WorkflowStep(raw) {
	case state.WorkflowStepIntake, state.WorkflowStepDebate, state.WorkflowStepBuild, state.WorkflowStepTest, state.WorkflowStepRelease, state.WorkflowStepBugfix, state.WorkflowStepResume:
		return state.WorkflowStep(raw)
	default:
		return state.WorkflowStepUnknown
	}
}

func parseRunMode(raw string) state.RunMode {
	switch raw {
	case string(state.RunModeBugfix):
		return state.RunModeBugfix
	case string(state.RunModeResume):
		return state.RunModeResume
	default:
		return state.RunModeIntake
	}
}

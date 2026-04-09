package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"web-app-agent-runtime/internal/issuebot"
)

func main() {
	if code := run(os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func run(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("issuebot", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("repo-root", ".", "repository root")
	label := fs.String("label", "", "label filter")
	once := fs.Bool("once", false, "run a single cycle and exit")
	dryRun := fs.Bool("dry-run", false, "apply fixes but do not commit/push")
	interval := fs.Duration("interval", 15*time.Minute, "poll interval")
	maxIssues := fs.Int("max-issues", 100, "maximum issues to inspect")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	res, err := issuebot.Run(context.Background(), issuebot.Options{
		RepoRoot:  *root,
		Label:     *label,
		Once:      *once,
		DryRun:    *dryRun,
		Interval:  *interval,
		MaxIssues: *maxIssues,
		AutoClose: true,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "checked=%d fixed=%d closed=%d skipped=%d failed=%d\n", res.Checked, res.Fixed, res.Closed, res.Skipped, res.Failed)
	return 0
}

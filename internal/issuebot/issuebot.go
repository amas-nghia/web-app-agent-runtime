package issuebot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type Issue struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	URL    string  `json:"url"`
	Labels []Label `json:"labels"`
}

type Options struct {
	RepoRoot   string
	Label      string
	Interval   time.Duration
	Once       bool
	DryRun     bool
	MaxIssues  int
	AutoClose  bool
	BranchName string
}

type Result struct {
	Checked int
	Fixed   int
	Closed  int
	Skipped int
	Failed  int
}

func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.RepoRoot == "" {
		return Result{}, fmt.Errorf("issuebot: repo root is required")
	}
	if opts.Interval <= 0 {
		opts.Interval = 15 * time.Minute
	}
	if opts.MaxIssues <= 0 {
		opts.MaxIssues = 100
	}
	if opts.BranchName == "" {
		opts.BranchName = "feat/issuebot-auto-fix"
	}

	var result Result
	for {
		cycle, err := runOnce(ctx, opts)
		result.Checked += cycle.Checked
		result.Fixed += cycle.Fixed
		result.Closed += cycle.Closed
		result.Skipped += cycle.Skipped
		result.Failed += cycle.Failed
		if err != nil {
			return result, err
		}
		if opts.Once {
			return result, nil
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(opts.Interval):
		}
	}
}

func runOnce(ctx context.Context, opts Options) (Result, error) {
	issues, err := listOpenIssues(ctx, opts.RepoRoot, opts.MaxIssues)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for _, issue := range issues {
		if opts.Label != "" && !hasLabel(issue, opts.Label) {
			result.Skipped++
			continue
		}
		result.Checked++
		kind := classify(issue)
		if kind == issueKindUnknown {
			result.Skipped++
			continue
		}

		changed, fixNote, err := applyFix(ctx, opts.RepoRoot, issue, kind)
		if err != nil {
			result.Failed++
			_ = commentIssue(ctx, issue.Number, fmt.Sprintf("Auto-fix failed: %v", err))
			continue
		}

		if err := runTests(ctx, opts.RepoRoot); err != nil {
			result.Failed++
			_ = commentIssue(ctx, issue.Number, fmt.Sprintf("Auto-fix applied but tests failed: %v", err))
			continue
		}

		if !opts.DryRun && changed {
			if err := commitAndPush(ctx, opts.RepoRoot, opts.BranchName, issue, fixNote); err != nil {
				result.Failed++
				_ = commentIssue(ctx, issue.Number, fmt.Sprintf("Auto-fix applied but commit/push failed: %v", err))
				continue
			}
		}

		result.Fixed++
		if opts.AutoClose {
			if err := commentIssue(ctx, issue.Number, fmt.Sprintf("Auto-fixed in branch `%s` (%s)", opts.BranchName, fixNote)); err != nil {
				result.Failed++
				continue
			}
			if err := closeIssue(ctx, issue.Number); err != nil {
				result.Failed++
				continue
			}
			result.Closed++
		}
	}

	return result, nil
}

type issueKind string

const (
	issueKindUnknown  issueKind = ""
	issueKindDocs     issueKind = "docs"
	issueKindOpencode issueKind = "opencode"
)

func classify(issue Issue) issueKind {
	t := strings.ToLower(issue.Title + "\n" + issue.Body)
	switch {
	case containsAny(t, []string{"output contract", "quick start", "mvp docs", "mock run flow", "app generator"}):
		return issueKindDocs
	case containsAny(t, []string{"opencode adapter", "interactive invocation", "hang", "timeout", "non-interactive"}):
		return issueKindOpencode
	default:
		return issueKindUnknown
	}
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

func hasLabel(issue Issue, label string) bool {
	for _, l := range issue.Labels {
		if strings.EqualFold(l.Name, label) {
			return true
		}
	}
	return false
}

func applyFix(ctx context.Context, root string, issue Issue, kind issueKind) (changed bool, note string, err error) {
	switch kind {
	case issueKindDocs:
		changed, err = fixDocs(root)
		return changed, "docs-output-contract", err
	case issueKindOpencode:
		changed, err = fixOpenCodeAdapter(root)
		return changed, "opencode-timeout", err
	default:
		return false, "", nil
	}
}

func runTests(ctx context.Context, root string) error {
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test ./...: %w", err)
	}
	return nil
}

func commitAndPush(ctx context.Context, root, branch string, issue Issue, note string) error {
	msg := fmt.Sprintf("fix issue #%d: %s", issue.Number, slugify(issue.Title))
	if err := runGit(ctx, root, "checkout", "-B", branch); err != nil {
		return err
	}
	if err := runGit(ctx, root, "add", "README.md", "docs/contracts.md", "internal/adapters/opencode/opencode.go", "internal/adapters/opencode/opencode_test.go", "docs/checklist.md", "docs/decision-log.md"); err != nil {
		return err
	}
	if err := runGit(ctx, root, "commit", "-m", msg); err != nil {
		if isNothingToCommit(err) {
			return nil
		}
		return err
	}
	if err := runGit(ctx, root, "push", "-u", "origin", branch); err != nil {
		return err
	}
	return nil
}

func runGit(ctx context.Context, root string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func isNothingToCommit(err error) bool {
	return err != nil && strings.Contains(err.Error(), "nothing to commit")
}

func commentIssue(ctx context.Context, number int, body string) error {
	return gh(ctx, "issue", "comment", fmt.Sprint(number), "--body", body)
}

func closeIssue(ctx context.Context, number int) error {
	return gh(ctx, "issue", "close", fmt.Sprint(number))
}

func gh(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func listOpenIssues(ctx context.Context, root string, limit int) ([]Issue, error) {
	cmd := exec.CommandContext(ctx, "gh", "issue", "list", "--state", "open", "--limit", fmt.Sprint(limit), "--json", "number,title,body,url,labels")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue list: %w", err)
	}
	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("issuebot: decode issues: %w", err)
	}
	return issues, nil
}

func slugify(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '-', r == '_':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

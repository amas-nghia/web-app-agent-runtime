package debate

import "time"

type Option struct {
	Name string `json:"name"`
	Why  string `json:"why"`
}

type Record struct {
	RunID     string   `json:"run_id"`
	Step      string   `json:"step"`
	Goal      string   `json:"goal"`
	Selected  string   `json:"selected"`
	Reason    string   `json:"reason"`
	Options   []Option `json:"options"`
	UpdatedAt string   `json:"updated_at"`
}

func New(runID, step, goal string) Record {
	return Record{
		RunID:     runID,
		Step:      step,
		Goal:      goal,
		Selected:  defaultSelection(step),
		Reason:    defaultReason(step),
		Options:   defaultOptions(step),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func defaultSelection(step string) string {
	switch step {
	case "intake":
		return "mvp-first"
	case "debate":
		return "fast-path"
	case "build":
		return "mock-first"
	case "test":
		return "smoke-plus-contract"
	case "release":
		return "preview-release"
	case "bugfix":
		return "hotfix-branch"
	case "resume":
		return "checkpoint-resume"
	default:
		return "mvp-first"
	}
}

func defaultReason(step string) string {
	switch step {
	case "build":
		return "start with the deterministic path so the system is easy to test"
	case "test":
		return "keep the first gate fast and reliable"
	case "release":
		return "ship preview first before wider rollout"
	case "bugfix":
		return "preserve release safety and isolate the fix"
	case "resume":
		return "continue from the last checkpoint instead of restarting"
	default:
		return "prefer the smallest viable path for the MVP"
	}
}

func defaultOptions(step string) []Option {
	switch step {
	case "intake":
		return []Option{{Name: "mvp-first", Why: "keep the scope narrow"}, {Name: "feature-rich", Why: "maximize optionality but slow the MVP"}}
	case "build":
		return []Option{{Name: "mock-first", Why: "fast, deterministic smoke path"}, {Name: "real-engine-first", Why: "higher fidelity but slower to stabilize"}}
	case "test":
		return []Option{{Name: "smoke-plus-contract", Why: "fast baseline coverage"}, {Name: "full-suite-only", Why: "slower but broader"}}
	case "release":
		return []Option{{Name: "preview-release", Why: "safe verification step"}, {Name: "direct-production", Why: "too risky for MVP"}}
	case "bugfix":
		return []Option{{Name: "hotfix-branch", Why: "contain the change"}, {Name: "patch-in-place", Why: "faster but riskier"}}
	case "resume":
		return []Option{{Name: "checkpoint-resume", Why: "continue from evidence"}, {Name: "full-rerun", Why: "slower and duplicates work"}}
	default:
		return []Option{{Name: "mvp-first", Why: "narrow, testable default"}, {Name: "expand-later", Why: "greater flexibility later"}}
	}
}

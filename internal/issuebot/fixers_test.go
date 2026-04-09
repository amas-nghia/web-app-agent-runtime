package issuebot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	if got := classify(Issue{Title: "Clarify output contract", Body: "quick start"}); got != issueKindDocs {
		t.Fatalf("classify docs = %q, want docs", got)
	}
	if got := classify(Issue{Title: "OpenCode adapter can hang", Body: "interactive invocation"}); got != issueKindOpencode {
		t.Fatalf("classify opencode = %q, want opencode", got)
	}
	if got := classify(Issue{Title: "something else", Body: "unrelated"}); got != issueKindUnknown {
		t.Fatalf("classify unknown = %q, want unknown", got)
	}
}

func TestFixDocs_NoopWhenAlreadyUpdated(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	copyRepoFiles(t, root)
	changed, err := fixDocs(root)
	if err != nil {
		t.Fatalf("fixDocs() error = %v", err)
	}
	if changed {
		t.Fatalf("fixDocs() changed = true, want false on already-updated repo")
	}
}

func copyRepoFiles(t *testing.T, root string) {
	t.Helper()

	files := []string{"README.md", filepath.Join("docs", "contracts.md"), filepath.Join("internal", "adapters", "opencode", "opencode.go")}
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", rel))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", rel, err)
		}
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", path, err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", path, err)
		}
	}
}

package debate

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	r := New("run-1", "build", "make a web app")
	if r.Selected != "mock-first" {
		t.Fatalf("Selected = %q, want %q", r.Selected, "mock-first")
	}
	if len(r.Options) != 2 {
		t.Fatalf("Options len = %d, want 2", len(r.Options))
	}
}

package adapters

import (
	"strings"
	"testing"

	"web-app-agent-runtime/internal/state"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		engine  string
		wantNil bool
		wantErr string
		want    state.CapabilitySet
	}{
		{name: "mock engine", engine: "mock", want: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}},
		{name: "opencode engine", engine: "opencode", want: state.CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}},
		{name: "unknown engine", engine: "does-not-exist", wantNil: true, wantErr: "adapters: unknown engine"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.

			// Act.
			got, err := New(tt.engine)

			// Assert.
			if tt.wantNil {
				if err == nil {
					t.Fatal("New() error = nil, want failure")
				}
				if got != nil {
					t.Fatalf("New() adapter = %#v, want nil", got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("New() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got == nil {
				t.Fatal("New() adapter = nil, want adapter")
			}
			if got.Name() != tt.engine {
				t.Fatalf("Name() = %q, want %q", got.Name(), tt.engine)
			}
			if got.Capabilities() != tt.want {
				t.Fatalf("Capabilities() = %#v, want %#v", got.Capabilities(), tt.want)
			}
		})
	}
}

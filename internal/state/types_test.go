package state

import "testing"

func TestCapabilitySet_MissingAndSatisfies(t *testing.T) {
	tests := []struct {
		name     string
		offered  CapabilitySet
		required CapabilitySet
		missing  []string
		wantOK   bool
	}{
		{
			name:     "all required capabilities present",
			offered:  CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
			required: CapabilitySet{Read: true, Write: true, Resume: true, ApprovalMemo: true},
			wantOK:   true,
		},
		{
			name:     "missing shell and approval memo",
			offered:  CapabilitySet{Read: true, Write: true, Resume: true},
			required: CapabilitySet{Read: true, Write: true, Shell: true, ApprovalMemo: true},
			missing:  []string{"shell", "approval_memo"},
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMissing := tt.offered.Missing(tt.required)
			if len(gotMissing) != len(tt.missing) {
				t.Fatalf("Missing() len = %d, want %d; got %v", len(gotMissing), len(tt.missing), gotMissing)
			}
			for i := range tt.missing {
				if gotMissing[i] != tt.missing[i] {
					t.Fatalf("Missing()[%d] = %q, want %q", i, gotMissing[i], tt.missing[i])
				}
			}
			if got := tt.offered.Satisfies(tt.required); got != tt.wantOK {
				t.Fatalf("Satisfies() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}

func TestCapabilitySet_Intersect(t *testing.T) {
	got := CapabilitySet{Read: true, Write: true, Shell: true, Resume: true, ApprovalMemo: true}.
		Intersect(CapabilitySet{Read: true, Write: false, Shell: true, Resume: false, ApprovalMemo: true})

	want := CapabilitySet{Read: true, Shell: true, ApprovalMemo: true}
	if got != want {
		t.Fatalf("Intersect() = %#v, want %#v", got, want)
	}
}

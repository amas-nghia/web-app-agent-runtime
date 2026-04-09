package progress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRecordAndCurrent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	record := New("run-1")
	record.Done = append(record.Done, "parsed config")
	record.State = StateBlocked
	record.Decision = &Decision{Value: "switch path", Why: "temp dir"}

	if err := WriteCurrent(root, record.RunID); err != nil {
		t.Fatalf("WriteCurrent() error = %v", err)
	}
	if err := WriteRecord(root, record); err != nil {
		t.Fatalf("WriteRecord() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "current"))
	if err != nil {
		t.Fatalf("ReadFile(current) error = %v", err)
	}
	if got := string(data); got != "run-1\n" {
		t.Fatalf("current = %q, want %q", got, "run-1\n")
	}
}

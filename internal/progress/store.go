package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CurrentPath(root string) string {
	return filepath.Join(root, "current")
}

func RecordPath(root, runID string) string {
	return filepath.Join(root, runID, "progress.json")
}

func WriteCurrent(root, runID string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("progress: mkdir root: %w", err)
	}
	if err := os.WriteFile(CurrentPath(root), []byte(runID+"\n"), 0o644); err != nil {
		return fmt.Errorf("progress: write current: %w", err)
	}
	return nil
}

func WriteRecord(root string, record Record) error {
	path := RecordPath(root, record.RunID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("progress: mkdir run dir: %w", err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("progress: marshal: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("progress: write record: %w", err)
	}
	return nil
}

func LoadCurrent(root string) (string, error) {
	data, err := os.ReadFile(CurrentPath(root))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func LoadRecord(root, runID string) (Record, error) {
	data, err := os.ReadFile(RecordPath(root, runID))
	if err != nil {
		return Record{}, err
	}
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("progress: unmarshal: %w", err)
	}
	return record, nil
}

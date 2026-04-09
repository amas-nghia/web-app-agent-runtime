package debate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func RecordPath(root, runID, step string) string {
	return filepath.Join(root, runID, "debates", step+".json")
}

func WriteRecord(root string, record Record) error {
	path := RecordPath(root, record.RunID, record.Step)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("debate: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("debate: marshal: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("debate: write: %w", err)
	}
	return nil
}

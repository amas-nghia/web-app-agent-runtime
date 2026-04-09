package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Path(root, runID string) string {
	return filepath.Join(root, runID, "history.jsonl")
}

func Append(root string, event Event) error {
	if root == "" || event.RunID == "" {
		return nil
	}
	path := Path(root, event.RunID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("history: mkdir: %w", err)
	}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("history: open: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("history: write: %w", err)
	}
	return nil
}

func Load(root, runID string) ([]Event, error) {
	path := Path(root, runID)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("history: unmarshal: %w", err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("history: scan: %w", err)
	}
	return events, nil
}

func Last(events []Event) (Event, bool) {
	if len(events) == 0 {
		return Event{}, false
	}
	return events[len(events)-1], true
}

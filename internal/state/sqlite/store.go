package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"web-app-agent-runtime/internal/state"
)

const tableName = "canonical_runs"

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite store: path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sqlite store: create parent dir: %w", err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite store: open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.Init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store: db is nil")
	}

	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS canonical_runs (
    id TEXT PRIMARY KEY,
    goal TEXT NOT NULL,
    mode TEXT NOT NULL,
    start_step TEXT NOT NULL DEFAULT '',
    repo_path TEXT NOT NULL DEFAULT '',
    base_branch TEXT NOT NULL DEFAULT '',
    bug_report TEXT NOT NULL DEFAULT '',
    resume_from_checkpoint TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    current_step TEXT NOT NULL,
    task_ids TEXT NOT NULL,
    attempt_ids TEXT NOT NULL,
    approval_ids TEXT NOT NULL,
    checkpoint_ids TEXT NOT NULL,
    artifact_ids TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
`)

	if err != nil {
		return fmt.Errorf("sqlite store: migrate: %w", err)
	}

	if err := s.ensureColumn(ctx, "start_step", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}

	return nil
}

func (s *Store) ensureColumn(ctx context.Context, name, decl string) error {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, tableName))
	if err != nil {
		return fmt.Errorf("sqlite store: inspect columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var colName, colType string
		var notnull int
		var dflt any
		var pk int
		if err := rows.Scan(&cid, &colName, &colType, &notnull, &dflt, &pk); err != nil {
			return fmt.Errorf("sqlite store: scan columns: %w", err)
		}
		if colName == name {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("sqlite store: inspect columns rows: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, tableName, name, decl)); err != nil {
		return fmt.Errorf("sqlite store: add column %s: %w", name, err)
	}
	return nil
}

func (s *Store) UpsertRun(ctx context.Context, run state.Run) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store: db is nil")
	}
	if run.ID == "" {
		return fmt.Errorf("sqlite store: run id is empty")
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now().UTC()
	}
	if run.UpdatedAt.IsZero() {
		run.UpdatedAt = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`
INSERT INTO %s (
    id, goal, mode, start_step, repo_path, base_branch, bug_report, resume_from_checkpoint,
    status, current_step, task_ids, attempt_ids, approval_ids, checkpoint_ids,
    artifact_ids, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    goal = excluded.goal,
    mode = excluded.mode,
    start_step = excluded.start_step,
    repo_path = excluded.repo_path,
    base_branch = excluded.base_branch,
    bug_report = excluded.bug_report,
    resume_from_checkpoint = excluded.resume_from_checkpoint,
    status = excluded.status,
    current_step = excluded.current_step,
    task_ids = excluded.task_ids,
    attempt_ids = excluded.attempt_ids,
    approval_ids = excluded.approval_ids,
    checkpoint_ids = excluded.checkpoint_ids,
    artifact_ids = excluded.artifact_ids,
    updated_at = excluded.updated_at
`, tableName),
		run.ID,
		run.Spec.Goal,
		string(run.Spec.Mode),
		string(run.Spec.StartStep),
		run.Spec.RepoPath,
		run.Spec.BaseBranch,
		run.Spec.BugReport,
		run.Spec.ResumeFromCheckpoint,
		string(run.Status),
		string(run.CurrentStep),
		mustJSON(run.TaskIDs),
		mustJSON(run.AttemptIDs),
		mustJSON(run.ApprovalIDs),
		mustJSON(run.CheckpointIDs),
		mustJSON(run.ArtifactIDs),
		run.CreatedAt.UTC().Format(time.RFC3339Nano),
		run.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("sqlite store: upsert run %q: %w", run.ID, err)
	}

	return nil
}

func (s *Store) GetRun(ctx context.Context, id string) (state.Run, error) {
	if s == nil || s.db == nil {
		return state.Run{}, fmt.Errorf("sqlite store: db is nil")
	}

	row := s.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT id, goal, mode, start_step, repo_path, base_branch, bug_report, resume_from_checkpoint,
       status, current_step, task_ids, attempt_ids, approval_ids, checkpoint_ids,
       artifact_ids, created_at, updated_at
FROM %s
WHERE id = ?
`, tableName), id)

	return scanRun(row)
}

func (s *Store) ListRuns(ctx context.Context) ([]state.Run, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("sqlite store: db is nil")
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
SELECT id, goal, mode, start_step, repo_path, base_branch, bug_report, resume_from_checkpoint,
       status, current_step, task_ids, attempt_ids, approval_ids, checkpoint_ids,
       artifact_ids, created_at, updated_at
FROM %s
ORDER BY created_at ASC, id ASC
`, tableName))
	if err != nil {
		return nil, fmt.Errorf("sqlite store: list runs: %w", err)
	}
	defer rows.Close()

	var runs []state.Run
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite store: list runs rows: %w", err)
	}

	return runs, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRun(row rowScanner) (state.Run, error) {
	var run state.Run
	var goal, mode, startStep, repoPath, baseBranch, bugReport, resumeFromCheckpoint string
	var status, currentStep string
	var taskIDsJSON, attemptIDsJSON, approvalIDsJSON, checkpointIDsJSON, artifactIDsJSON string
	var createdAt, updatedAt string

	if err := row.Scan(
		&run.ID,
		&goal,
		&mode,
		&startStep,
		&repoPath,
		&baseBranch,
		&bugReport,
		&resumeFromCheckpoint,
		&status,
		&currentStep,
		&taskIDsJSON,
		&attemptIDsJSON,
		&approvalIDsJSON,
		&checkpointIDsJSON,
		&artifactIDsJSON,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return state.Run{}, state.ErrRunNotFound
		}
		return state.Run{}, fmt.Errorf("sqlite store: scan run: %w", err)
	}

	taskIDs, err := decodeStrings(taskIDsJSON)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: decode task ids: %w", err)
	}
	attemptIDs, err := decodeStrings(attemptIDsJSON)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: decode attempt ids: %w", err)
	}
	approvalIDs, err := decodeStrings(approvalIDsJSON)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: decode approval ids: %w", err)
	}
	checkpointIDs, err := decodeStrings(checkpointIDsJSON)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: decode checkpoint ids: %w", err)
	}
	artifactIDs, err := decodeStrings(artifactIDsJSON)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: decode artifact ids: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: parse created_at: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return state.Run{}, fmt.Errorf("sqlite store: parse updated_at: %w", err)
	}

	run.Spec = state.RunSpec{
		Goal:                 goal,
		Mode:                 state.RunMode(mode),
		StartStep:            state.WorkflowStep(startStep),
		RepoPath:             repoPath,
		BaseBranch:           baseBranch,
		BugReport:            bugReport,
		ResumeFromCheckpoint: resumeFromCheckpoint,
	}
	run.Status = state.RunStatus(status)
	run.CurrentStep = state.WorkflowStep(currentStep)
	run.TaskIDs = taskIDs
	run.AttemptIDs = attemptIDs
	run.ApprovalIDs = approvalIDs
	run.CheckpointIDs = checkpointIDs
	run.ArtifactIDs = artifactIDs
	run.CreatedAt = created
	run.UpdatedAt = updated

	return run, nil
}

func mustJSON[T any](v T) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func decodeStrings(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

var _ state.RunStore = (*Store)(nil)

package state

import "time"

type RunMode string

const (
	RunModeUnknown RunMode = ""
	RunModeIntake  RunMode = "intake"
	RunModeBugfix  RunMode = "bugfix"
	RunModeResume  RunMode = "resume"
)

type WorkflowStep string

const (
	WorkflowStepUnknown WorkflowStep = ""
	WorkflowStepIntake  WorkflowStep = "intake"
	WorkflowStepDebate  WorkflowStep = "debate"
	WorkflowStepBuild   WorkflowStep = "build"
	WorkflowStepTest    WorkflowStep = "test"
	WorkflowStepRelease WorkflowStep = "release"
	WorkflowStepBugfix  WorkflowStep = "bugfix"
	WorkflowStepResume  WorkflowStep = "resume"
)

type RunStatus string

const (
	RunStatusUnknown         RunStatus = ""
	RunStatusQueued          RunStatus = "queued"
	RunStatusRunning         RunStatus = "running"
	RunStatusWaitingApproval RunStatus = "waiting_approval"
	RunStatusBlocked         RunStatus = "blocked"
	RunStatusFailed          RunStatus = "failed"
	RunStatusCompleted       RunStatus = "completed"
	RunStatusCancelled       RunStatus = "cancelled"
)

type TaskStatus string

const (
	TaskStatusUnknown TaskStatus = ""
	TaskStatusPlanned TaskStatus = "planned"
	TaskStatusReady   TaskStatus = "ready"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusBlocked TaskStatus = "blocked"
	TaskStatusPassed  TaskStatus = "passed"
	TaskStatusFailed  TaskStatus = "failed"
	TaskStatusSkipped TaskStatus = "skipped"
)

type AttemptStatus string

const (
	AttemptStatusUnknown AttemptStatus = ""
	AttemptStatusRunning AttemptStatus = "running"
	AttemptStatusPassed  AttemptStatus = "passed"
	AttemptStatusFailed  AttemptStatus = "failed"
	AttemptStatusBlocked AttemptStatus = "blocked"
	AttemptStatusResumed AttemptStatus = "resumed"
)

type ApprovalStatus string

const (
	ApprovalStatusUnknown   ApprovalStatus = ""
	ApprovalStatusRequested ApprovalStatus = "requested"
	ApprovalStatusApproved  ApprovalStatus = "approved"
	ApprovalStatusRejected  ApprovalStatus = "rejected"
	ApprovalStatusSkipped   ApprovalStatus = "skipped"
)

type CheckpointKind string

const (
	CheckpointKindUnknown CheckpointKind = ""
	CheckpointKindState   CheckpointKind = "state"
	CheckpointKindResume  CheckpointKind = "resume"
	CheckpointKindRelease CheckpointKind = "release"
)

type ArtifactKind string

const (
	ArtifactKindUnknown ArtifactKind = ""
	ArtifactKindSource  ArtifactKind = "source"
	ArtifactKindPatch   ArtifactKind = "patch"
	ArtifactKindTestLog ArtifactKind = "test_log"
	ArtifactKindRelease ArtifactKind = "release_bundle"
)

type ApprovalTier string

const (
	ApprovalAuto  ApprovalTier = "auto"
	ApprovalAsk   ApprovalTier = "ask"
	ApprovalBlock ApprovalTier = "block"
)

type CapabilitySet struct {
	Read         bool
	Write        bool
	Shell        bool
	Resume       bool
	ApprovalMemo bool
}

type RunSpec struct {
	RunID                string
	Goal                 string
	Mode                 RunMode
	StartStep            WorkflowStep
	RepoPath             string
	BaseBranch           string
	BugReport            string
	ResumeFromCheckpoint string
	ArtifactsRoot        string
}

type Run struct {
	ID            string
	Spec          RunSpec
	Status        RunStatus
	CurrentStep   WorkflowStep
	TaskIDs       []string
	AttemptIDs    []string
	ApprovalIDs   []string
	CheckpointIDs []string
	ArtifactIDs   []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Task struct {
	ID               string
	RunID            string
	ParentTaskID     string
	DependsOnTaskIDs []string
	Step             WorkflowStep
	Status           TaskStatus
	CurrentAttemptID string
	AttemptIDs       []string
	ApprovalID       string
	CheckpointID     string
	Title            string
	Prompt           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Attempt struct {
	ID           string
	RunID        string
	TaskID       string
	Step         WorkflowStep
	Sequence     int
	Status       AttemptStatus
	Engine       string
	ApprovalID   string
	CheckpointID string
	Observation  string
	Result       string
	Error        string
	StartedAt    time.Time
	FinishedAt   time.Time
}

type Approval struct {
	ID          string
	RunID       string
	TaskID      string
	AttemptID   string
	Tier        ApprovalTier
	Status      ApprovalStatus
	Reason      string
	Decision    string
	RequestedAt time.Time
	DecidedAt   time.Time
}

type Checkpoint struct {
	ID        string
	RunID     string
	TaskID    string
	AttemptID string
	Kind      CheckpointKind
	Ref       string
	CreatedAt time.Time
}

type ArtifactRef struct {
	ID        string
	RunID     string
	TaskID    string
	AttemptID string
	Kind      ArtifactKind
	Path      string
	Digest    string
	CreatedAt time.Time
}

type AuditEntry struct {
	ID        string
	RunID     string
	Entity    string
	EntityID  string
	Message   string
	CreatedAt time.Time
}

type State struct {
	Run         Run
	Tasks       []Task
	Attempts    []Attempt
	Approvals   []Approval
	Checkpoints []Checkpoint
	Artifacts   []ArtifactRef
	Audit       []AuditEntry
}

type TaskSpec struct {
	RunID                string
	ParentTaskID         string
	DependsOnTaskIDs     []string
	Step                 WorkflowStep
	Title                string
	Prompt               string
	RequiresApprovalTier ApprovalTier
	ResumeFromCheckpoint string
}

type ApprovalRequest struct {
	RunID     string
	TaskID    string
	AttemptID string
	Step      WorkflowStep
	Tier      ApprovalTier
	Reason    string
}

type ApprovalDecision struct {
	RequestID string
	Approved  bool
	Memo      string
	DecidedBy string
}

type ExecutionResult struct {
	RunID        string
	TaskID       string
	AttemptID    string
	Summary      string
	Artifacts    []ArtifactRef
	Checkpoint   Checkpoint
	CheckpointID string
	Error        string
}

type Observation struct {
	Message   string
	CreatedAt time.Time
}

type Action struct {
	Name string
	Args map[string]string
}

func (c CapabilitySet) Satisfies(required CapabilitySet) bool {
	return len(c.Missing(required)) == 0
}

func (c CapabilitySet) Missing(required CapabilitySet) []string {
	missing := make([]string, 0, 5)
	if required.Read && !c.Read {
		missing = append(missing, "read")
	}
	if required.Write && !c.Write {
		missing = append(missing, "write")
	}
	if required.Shell && !c.Shell {
		missing = append(missing, "shell")
	}
	if required.Resume && !c.Resume {
		missing = append(missing, "resume")
	}
	if required.ApprovalMemo && !c.ApprovalMemo {
		missing = append(missing, "approval_memo")
	}
	return missing
}

func (c CapabilitySet) Intersect(other CapabilitySet) CapabilitySet {
	return CapabilitySet{
		Read:         c.Read && other.Read,
		Write:        c.Write && other.Write,
		Shell:        c.Shell && other.Shell,
		Resume:       c.Resume && other.Resume,
		ApprovalMemo: c.ApprovalMemo && other.ApprovalMemo,
	}
}

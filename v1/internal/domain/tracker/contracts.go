package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"fmt"
	"strings"
	"time"
)

type BoardState string

const (
	BoardStatePending   BoardState = "pending"
	BoardStateActive    BoardState = "active"
	BoardStateCompleted BoardState = "completed"
	BoardStateFailed    BoardState = "failed"
)

func (state BoardState) Validate() error {
	switch state {
	case BoardStatePending, BoardStateActive, BoardStateCompleted, BoardStateFailed:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported board state %q", state))
	}
}

type EpicState string

const (
	EpicStatePlanned    EpicState = "planned"
	EpicStateInProgress EpicState = "in_progress"
	EpicStateCompleted  EpicState = "completed"
	EpicStateBlocked    EpicState = "blocked"
	EpicStateFailed     EpicState = "failed"
)

func (state EpicState) Validate() error {
	switch state {
	case EpicStatePlanned, EpicStateInProgress, EpicStateCompleted, EpicStateBlocked, EpicStateFailed:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported epic state %q", state))
	}
}

type TaskState string

const (
	TaskStatePlanned      TaskState = "planned"
	TaskStateInProgress   TaskState = "in_progress"
	TaskStateCompleted    TaskState = "completed"
	TaskStateFailed       TaskState = "failed"
	TaskStateNoWorkNeeded TaskState = "no_work_needed"
)

func (state TaskState) Validate() error {
	switch state {
	case TaskStatePlanned, TaskStateInProgress, TaskStateCompleted, TaskStateFailed, TaskStateNoWorkNeeded:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported task state %q", state))
	}
}

type OutcomeStatus string

const (
	OutcomeStatusSuccess OutcomeStatus = "success"
	OutcomeStatusPartial OutcomeStatus = "partial"
	OutcomeStatusFailed  OutcomeStatus = "failed"
)

func (status OutcomeStatus) Validate() error {
	switch status {
	case OutcomeStatusSuccess, OutcomeStatusPartial, OutcomeStatusFailed:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported outcome status %q", status))
	}
}

type WorkItemID string

func (id WorkItemID) Validate(name string) error {
	if strings.TrimSpace(string(id)) == "" {
		return failures.WrapTerminal(fmt.Errorf("%s is required", name))
	}
	return nil
}

type TaskModelAudit struct {
	ModelProvider     string     `json:"model_provider"`
	ModelName         string     `json:"model_name"`
	ModelVersion      string     `json:"model_version,omitempty"`
	ModelRunID        string     `json:"model_run_id,omitempty"`
	PromptFingerprint string     `json:"prompt_fingerprint,omitempty"`
	InputTokens       *int       `json:"input_tokens,omitempty"`
	OutputTokens      *int       `json:"output_tokens,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type TaskOutcome struct {
	Status       OutcomeStatus `json:"status"`
	Summary      string        `json:"summary"`
	ErrorCode    string        `json:"error_code,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

func (outcome TaskOutcome) Validate() error {
	if err := outcome.Status.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(outcome.Summary) == "" {
		return failures.WrapTerminal(errors.New("summary is required"))
	}
	if outcome.Status == OutcomeStatusFailed && strings.TrimSpace(outcome.ErrorCode) == "" {
		return failures.WrapTerminal(errors.New("error_code is required for failed outcome"))
	}
	return nil
}

type Task struct {
	ID              WorkItemID      `json:"id"`
	BoardID         string          `json:"board_id"`
	EpicID          WorkItemID      `json:"epic_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	TaskType        string          `json:"task_type"`
	State           TaskState       `json:"state"`
	Rank            int             `json:"rank"`
	DependsOnTaskIDs []WorkItemID   `json:"depends_on_task_ids,omitempty"`
	Audit           TaskModelAudit  `json:"audit,omitempty"`
	Outcome         *TaskOutcome    `json:"outcome,omitempty"`
	ClaimedByAgentID string         `json:"claimed_by_agent_id,omitempty"`
	ClaimedAt       *time.Time      `json:"claimed_at,omitempty"`
	ClaimExpiresAt  *time.Time      `json:"claim_expires_at,omitempty"`
	ClaimToken      string          `json:"claim_token,omitempty"`
	AttemptCount    int             `json:"attempt_count,omitempty"`
	CreatedAt       time.Time       `json:"created_at,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at,omitempty"`
}

func (task Task) Validate() error {
	if err := task.ID.Validate("task id"); err != nil {
		return err
	}
	if strings.TrimSpace(task.BoardID) == "" {
		return failures.WrapTerminal(errors.New("task board_id is required"))
	}
	if err := task.EpicID.Validate("task epic_id"); err != nil {
		return err
	}
	if strings.TrimSpace(task.Title) == "" {
		return failures.WrapTerminal(errors.New("task title is required"))
	}
	if strings.TrimSpace(task.TaskType) == "" {
		return failures.WrapTerminal(errors.New("task_type is required"))
	}
	if err := task.State.Validate(); err != nil {
		return err
	}
	if task.Rank < 0 {
		return failures.WrapTerminal(errors.New("task rank cannot be negative"))
	}
	for _, dependencyID := range task.DependsOnTaskIDs {
		if err := dependencyID.Validate("task dependency id"); err != nil {
			return err
		}
		if dependencyID == task.ID {
			return failures.WrapTerminal(fmt.Errorf("task %s cannot depend on itself", task.ID))
		}
	}
	if task.Outcome != nil {
		if err := task.Outcome.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Epic struct {
	ID               WorkItemID    `json:"id"`
	BoardID          string        `json:"board_id"`
	Title            string        `json:"title"`
	Objective        string        `json:"objective,omitempty"`
	State            EpicState     `json:"state"`
	Rank             int           `json:"rank"`
	DependsOnEpicIDs []WorkItemID  `json:"depends_on_epic_ids,omitempty"`
	Tasks            []Task        `json:"tasks"`
	CreatedAt        time.Time     `json:"created_at,omitempty"`
	UpdatedAt        time.Time     `json:"updated_at,omitempty"`
}

func (epic Epic) Validate() error {
	if err := epic.ID.Validate("epic id"); err != nil {
		return err
	}
	if strings.TrimSpace(epic.BoardID) == "" {
		return failures.WrapTerminal(errors.New("epic board_id is required"))
	}
	if strings.TrimSpace(epic.Title) == "" {
		return failures.WrapTerminal(errors.New("epic title is required"))
	}
	if err := epic.State.Validate(); err != nil {
		return err
	}
	if epic.Rank < 0 {
		return failures.WrapTerminal(errors.New("epic rank cannot be negative"))
	}
	for _, dependencyID := range epic.DependsOnEpicIDs {
		if err := dependencyID.Validate("epic dependency id"); err != nil {
			return err
		}
		if dependencyID == epic.ID {
			return failures.WrapTerminal(fmt.Errorf("epic %s cannot depend on itself", epic.ID))
		}
	}
	for _, task := range epic.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		if task.BoardID != epic.BoardID {
			return failures.WrapTerminal(fmt.Errorf("task %s board_id must match epic board_id", task.ID))
		}
		if task.EpicID != epic.ID {
			return failures.WrapTerminal(fmt.Errorf("task %s epic_id must match parent epic", task.ID))
		}
	}
	return nil
}

type Board struct {
	BoardID   string     `json:"board_id"`
	RunID     string     `json:"run_id"`
	ProjectID string     `json:"project_id,omitempty"`
	Name      string     `json:"name,omitempty"`
	State     BoardState `json:"state"`
	Epics     []Epic     `json:"epics"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (board Board) Validate() error {
	if strings.TrimSpace(board.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(board.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if err := board.State.Validate(); err != nil {
		return err
	}
	epicIDs := make(map[WorkItemID]struct{}, len(board.Epics))
	taskIDs := map[WorkItemID]struct{}{}
	for _, epic := range board.Epics {
		if err := epic.Validate(); err != nil {
			return err
		}
		if epic.BoardID != board.BoardID {
			return failures.WrapTerminal(fmt.Errorf("epic %s board_id must equal board board_id", epic.ID))
		}
		if _, exists := epicIDs[epic.ID]; exists {
			return failures.WrapTerminal(fmt.Errorf("duplicate epic id %q", epic.ID))
		}
		epicIDs[epic.ID] = struct{}{}
		for _, task := range epic.Tasks {
			if _, exists := taskIDs[task.ID]; exists {
				return failures.WrapTerminal(fmt.Errorf("duplicate task id %q", task.ID))
			}
			taskIDs[task.ID] = struct{}{}
		}
	}
	for _, epic := range board.Epics {
		for _, dependencyEpicID := range epic.DependsOnEpicIDs {
			if _, exists := epicIDs[dependencyEpicID]; !exists {
				return failures.WrapTerminal(fmt.Errorf("epic %s depends on missing epic %s", epic.ID, dependencyEpicID))
			}
		}
		for _, task := range epic.Tasks {
			for _, dependencyTaskID := range task.DependsOnTaskIDs {
				if _, exists := taskIDs[dependencyTaskID]; !exists {
					return failures.WrapTerminal(fmt.Errorf("task %s depends on missing task %s", task.ID, dependencyTaskID))
				}
			}
		}
	}
	return nil
}

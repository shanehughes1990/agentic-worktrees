package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusNotStarted Status = "not-started"
	StatusInProgress Status = "in-progress"
	StatusCompleted  Status = "completed"
	StatusBlocked    Status = "blocked"
)

func (status Status) Validate() error {
	switch status {
	case StatusNotStarted, StatusInProgress, StatusCompleted, StatusBlocked:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported status %q", status))
	}
}

type SourceKind string

const (
	SourceKindInternal     SourceKind = "internal"
	SourceKindGitHubIssues SourceKind = "github_issues"
)

func (kind SourceKind) Validate() error {
	switch kind {
	case SourceKindInternal, SourceKindGitHubIssues:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported source kind %q", kind))
	}
}

type SourceRef struct {
	Kind     SourceKind     `json:"kind"`
	Location string         `json:"location,omitempty"`
	BoardID  string         `json:"board_id,omitempty"`
	Config   map[string]any `json:"config,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (source SourceRef) Validate() error {
	if err := source.Kind.Validate(); err != nil {
		return err
	}
	switch source.Kind {
	case SourceKindInternal:
		if strings.TrimSpace(source.Location) == "" {
			return failures.WrapTerminal(errors.New("location is required for internal source"))
		}
	case SourceKindGitHubIssues:
		if strings.TrimSpace(source.Location) == "" {
			return failures.WrapTerminal(errors.New("location is required for github_issues source (owner/repo)"))
		}
	}
	return nil
}

type WorkItemID string

func (id WorkItemID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return failures.WrapTerminal(errors.New("id is required"))
	}
	return nil
}

type Priority string

const (
	PriorityP0 Priority = "p0"
	PriorityP1 Priority = "p1"
	PriorityP2 Priority = "p2"
	PriorityP3 Priority = "p3"
)

func (priority Priority) Validate() error {
	if strings.TrimSpace(string(priority)) == "" {
		return nil
	}
	switch priority {
	case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported priority %q", priority))
	}
}

type WorkItem struct {
	ID          WorkItemID     `json:"id"`
	BoardID     string         `json:"board_id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      Status         `json:"status"`
	Priority    Priority       `json:"priority,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
}

func (item WorkItem) Validate() error {
	if err := item.ID.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(item.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(item.Title) == "" {
		return failures.WrapTerminal(errors.New("title is required"))
	}
	if err := item.Status.Validate(); err != nil {
		return err
	}
	if err := item.Priority.Validate(); err != nil {
		return err
	}
	return nil
}

type TaskOutcome struct {
	Status          string    `json:"status"`
	Reason          string    `json:"reason,omitempty"`
	TaskBranch      string    `json:"task_branch,omitempty"`
	Worktree        string    `json:"worktree,omitempty"`
	ResumeSessionID string    `json:"resume_session_id,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

func (outcome TaskOutcome) Validate() error {
	if strings.TrimSpace(outcome.Status) == "" {
		return failures.WrapTerminal(errors.New("status is required"))
	}
	return nil
}

type Task struct {
	WorkItem
	DependsOn []WorkItemID `json:"depends_on,omitempty"`
	Outcome   *TaskOutcome `json:"outcome,omitempty"`
}

func (task Task) Validate() error {
	if err := task.WorkItem.Validate(); err != nil {
		return err
	}
	if task.Outcome != nil {
		if err := task.Outcome.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Epic struct {
	WorkItem
	DependsOn []WorkItemID `json:"depends_on,omitempty"`
	Tasks     []Task       `json:"tasks"`
}

func (epic Epic) Validate() error {
	if err := epic.WorkItem.Validate(); err != nil {
		return err
	}
	if len(epic.Tasks) == 0 {
		return failures.WrapTerminal(errors.New("tasks are required"))
	}
	for _, task := range epic.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Board struct {
	BoardID   string         `json:"board_id"`
	RunID     string         `json:"run_id"`
	Title     string         `json:"title,omitempty"`
	Goal      string         `json:"goal,omitempty"`
	Source    SourceRef      `json:"source"`
	Status    Status         `json:"status"`
	Epics     []Epic         `json:"epics"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (board Board) Validate() error {
	if strings.TrimSpace(board.BoardID) == "" {
		return failures.WrapTerminal(errors.New("board_id is required"))
	}
	if strings.TrimSpace(board.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if err := board.Source.Validate(); err != nil {
		return err
	}
	if err := board.Status.Validate(); err != nil {
		return err
	}
	if len(board.Epics) == 0 {
		return failures.WrapTerminal(errors.New("epics are required"))
	}

	epicIDs := make(map[WorkItemID]struct{}, len(board.Epics))
	taskIDs := map[WorkItemID]struct{}{}
	for _, epic := range board.Epics {
		if err := epic.Validate(); err != nil {
			return err
		}
		if strings.TrimSpace(epic.BoardID) != board.BoardID {
			return failures.WrapTerminal(fmt.Errorf("epic %s board_id must equal board board_id", epic.ID))
		}
		if _, exists := epicIDs[epic.ID]; exists {
			return failures.WrapTerminal(fmt.Errorf("duplicate epic id %q", epic.ID))
		}
		epicIDs[epic.ID] = struct{}{}
		for _, task := range epic.Tasks {
			if strings.TrimSpace(task.BoardID) != board.BoardID {
				return failures.WrapTerminal(fmt.Errorf("task %s board_id must equal board board_id", task.ID))
			}
			if _, exists := taskIDs[task.ID]; exists {
				return failures.WrapTerminal(fmt.Errorf("duplicate task id %q", task.ID))
			}
			taskIDs[task.ID] = struct{}{}
		}
	}

	for _, epic := range board.Epics {
		for _, dependencyEpicID := range epic.DependsOn {
			if _, exists := epicIDs[WorkItemID(strings.TrimSpace(string(dependencyEpicID)))]; !exists {
				return failures.WrapTerminal(fmt.Errorf("epic %s depends on missing epic %s", epic.ID, dependencyEpicID))
			}
		}
		for _, task := range epic.Tasks {
			for _, dependencyTaskID := range task.DependsOn {
				if _, exists := taskIDs[WorkItemID(strings.TrimSpace(string(dependencyTaskID)))]; !exists {
					return failures.WrapTerminal(fmt.Errorf("task %s depends on missing task %s", task.ID, dependencyTaskID))
				}
			}
		}
	}

	return nil
}

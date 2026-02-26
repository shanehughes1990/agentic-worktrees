package taskboard

import (
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

type WorkItem struct {
	ID          string         `json:"id"`
	BoardID     string         `json:"board_id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      Status         `json:"status"`
	Priority    string         `json:"priority,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at,omitempty"`
}

type Task struct {
	WorkItem
	DependsOn []string     `json:"depends_on,omitempty"`
	Outcome   *TaskOutcome `json:"outcome,omitempty"`
}

type TaskOutcome struct {
	Status          string    `json:"status"`
	Reason          string    `json:"reason,omitempty"`
	TaskBranch      string    `json:"task_branch,omitempty"`
	Worktree        string    `json:"worktree,omitempty"`
	ResumeSessionID string    `json:"resume_session_id,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type Epic struct {
	WorkItem
	DependsOn []string `json:"depends_on,omitempty"`
	Tasks     []Task   `json:"tasks"`
}

type Board struct {
	BoardID   string          `json:"board_id"`
	RunID     string          `json:"run_id"`
	Title     string          `json:"title,omitempty"`
	Goal      string          `json:"goal,omitempty"`
	Source    *SourceMetadata `json:"source,omitempty"`
	Status    Status          `json:"status"`
	Epics     []Epic          `json:"epics"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (board *Board) ValidateBasics() error {
	if strings.TrimSpace(board.BoardID) == "" {
		return fmt.Errorf("board_id is required")
	}
	if strings.TrimSpace(board.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if board.Source != nil {
		if err := board.Source.ValidateBasics(); err != nil {
			return fmt.Errorf("source metadata invalid: %w", err)
		}
	}

	epicByID := make(map[string]struct{}, len(board.Epics))
	taskByID := map[string]struct{}{}
	for _, epic := range board.Epics {
		epicID := strings.TrimSpace(epic.ID)
		if epicID == "" {
			return fmt.Errorf("epic id is required")
		}
		if _, exists := epicByID[epicID]; exists {
			return fmt.Errorf("duplicate epic id: %s", epicID)
		}
		epicByID[epicID] = struct{}{}

		for _, task := range epic.Tasks {
			taskID := strings.TrimSpace(task.ID)
			if taskID == "" {
				return fmt.Errorf("task id is required")
			}
			if _, exists := taskByID[taskID]; exists {
				return fmt.Errorf("duplicate task id: %s", taskID)
			}
			taskByID[taskID] = struct{}{}
		}
	}

	for _, epic := range board.Epics {
		for _, dependencyEpicID := range epic.DependsOn {
			cleanDependencyEpicID := strings.TrimSpace(dependencyEpicID)
			if cleanDependencyEpicID == "" {
				continue
			}
			if _, exists := epicByID[cleanDependencyEpicID]; !exists {
				return fmt.Errorf("epic %s depends_on missing epic %s", epic.ID, cleanDependencyEpicID)
			}
		}

		for _, task := range epic.Tasks {
			for _, dependencyTaskID := range task.DependsOn {
				cleanDependencyTaskID := strings.TrimSpace(dependencyTaskID)
				if cleanDependencyTaskID == "" {
					continue
				}
				if _, exists := taskByID[cleanDependencyTaskID]; !exists {
					return fmt.Errorf("task %s depends_on missing task %s", task.ID, cleanDependencyTaskID)
				}
			}
		}
	}

	return nil
}

func (board *Board) ValidateComplete() error {
	if len(board.Epics) == 0 {
		return fmt.Errorf("taskboard must include epics")
	}

	for _, epic := range board.Epics {
		if strings.TrimSpace(epic.Title) == "" {
			return fmt.Errorf("epic %s must have title", epic.ID)
		}
		if len(epic.Tasks) == 0 {
			return fmt.Errorf("epic %s must include at least one task", epic.ID)
		}
		for _, task := range epic.Tasks {
			if strings.TrimSpace(task.Title) == "" {
				return fmt.Errorf("task %s must have title", task.ID)
			}
		}
	}

	return nil
}

func (board *Board) IsCompleted(workItemID string) bool {
	for _, epic := range board.Epics {
		if epic.ID == workItemID {
			return epic.Status == StatusCompleted
		}
		for _, task := range epic.Tasks {
			if task.ID == workItemID {
				return task.Status == StatusCompleted
			}
		}
	}
	return false
}

func (board *Board) SetTaskStatus(taskID string, status Status) error {
	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			if board.Epics[epicIndex].Tasks[taskIndex].ID == taskID {
				board.Epics[epicIndex].Tasks[taskIndex].Status = status
				now := time.Now().UTC()
				board.Epics[epicIndex].Tasks[taskIndex].UpdatedAt = now
				board.UpdatedAt = now
				return nil
			}
		}
	}
	return fmt.Errorf("task not found: %s", taskID)
}

func (board *Board) SetTaskOutcome(taskID string, outcome TaskOutcome) error {
	for epicIndex := range board.Epics {
		for taskIndex := range board.Epics[epicIndex].Tasks {
			if board.Epics[epicIndex].Tasks[taskIndex].ID != taskID {
				continue
			}
			now := time.Now().UTC()
			if outcome.UpdatedAt.IsZero() {
				outcome.UpdatedAt = now
			}
			board.Epics[epicIndex].Tasks[taskIndex].Outcome = &outcome
			board.Epics[epicIndex].Tasks[taskIndex].UpdatedAt = now
			board.UpdatedAt = now
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", taskID)
}

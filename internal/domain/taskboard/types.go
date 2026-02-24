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

type Epic struct {
	WorkItem
	ItemType string `json:"item_type"`
}

type Task struct {
	WorkItem
	ItemType string `json:"item_type"`
	EpicID   string `json:"epic_id"`
}

type MicroTask struct {
	WorkItem
	ItemType string `json:"item_type"`
	TaskID   string `json:"task_id"`
}

type Dependency struct {
	EdgeID         string `json:"edge_id"`
	BoardID        string `json:"board_id"`
	FromID         string `json:"from_id"`
	ToID           string `json:"to_id"`
	DependencyType string `json:"dependency_type,omitempty"`
}

type Board struct {
	BoardID      string         `json:"board_id"`
	RunID        string         `json:"run_id"`
	Status       Status         `json:"status"`
	Epics        []Epic         `json:"epics"`
	Tasks        []Task         `json:"tasks"`
	MicroTasks   []MicroTask    `json:"micro_tasks"`
	Dependencies []Dependency   `json:"dependencies"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (board *Board) ValidateBasics() error {
	if strings.TrimSpace(board.BoardID) == "" {
		return fmt.Errorf("board_id is required")
	}
	if strings.TrimSpace(board.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}

	seen := make(map[string]struct{})
	for _, epic := range board.Epics {
		if strings.TrimSpace(epic.ID) == "" {
			return fmt.Errorf("epic id is required")
		}
		if _, exists := seen[epic.ID]; exists {
			return fmt.Errorf("duplicate work item id: %s", epic.ID)
		}
		seen[epic.ID] = struct{}{}
	}
	for _, task := range board.Tasks {
		if strings.TrimSpace(task.ID) == "" {
			return fmt.Errorf("task id is required")
		}
		if _, exists := seen[task.ID]; exists {
			return fmt.Errorf("duplicate work item id: %s", task.ID)
		}
		seen[task.ID] = struct{}{}
	}
	for _, microTask := range board.MicroTasks {
		if strings.TrimSpace(microTask.ID) == "" {
			return fmt.Errorf("micro_task id is required")
		}
		if _, exists := seen[microTask.ID]; exists {
			return fmt.Errorf("duplicate work item id: %s", microTask.ID)
		}
		seen[microTask.ID] = struct{}{}
	}
	for _, dependency := range board.Dependencies {
		if strings.TrimSpace(dependency.FromID) == "" || strings.TrimSpace(dependency.ToID) == "" {
			return fmt.Errorf("dependency from_id and to_id are required")
		}
		if _, exists := seen[dependency.FromID]; !exists {
			return fmt.Errorf("dependency references unknown from_id: %s", dependency.FromID)
		}
		if _, exists := seen[dependency.ToID]; !exists {
			return fmt.Errorf("dependency references unknown to_id: %s", dependency.ToID)
		}
	}

	return nil
}

func (board *Board) IsCompleted(workItemID string) bool {
	for _, epic := range board.Epics {
		if epic.ID == workItemID {
			return epic.Status == StatusCompleted
		}
	}
	for _, task := range board.Tasks {
		if task.ID == workItemID {
			return task.Status == StatusCompleted
		}
	}
	for _, microTask := range board.MicroTasks {
		if microTask.ID == workItemID {
			return microTask.Status == StatusCompleted
		}
	}
	return false
}

func (board *Board) SetMicroTaskStatus(microTaskID string, status Status) error {
	for index := range board.MicroTasks {
		if board.MicroTasks[index].ID == microTaskID {
			board.MicroTasks[index].Status = status
			now := time.Now().UTC()
			board.MicroTasks[index].UpdatedAt = now
			board.UpdatedAt = now
			return nil
		}
	}
	return fmt.Errorf("micro_task not found: %s", microTaskID)
}

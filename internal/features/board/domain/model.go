package domain

import (
	"fmt"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
)

type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Dependencies []string   `json:"dependencies"`
	Lane         string     `json:"lane"`
	Status       TaskStatus `json:"status"`
}

type Epic struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Dependencies []string `json:"dependencies"`
	Tasks        []Task   `json:"tasks"`
}

type Board struct {
	SchemaVersion int       `json:"schema_version"`
	SourceScope   string    `json:"source_scope"`
	GeneratedAt   time.Time `json:"generated_at"`
	Epics         []Epic    `json:"epics"`
}

func (b Board) Validate() error {
	if b.SchemaVersion < 1 {
		return fmt.Errorf("schema_version must be >= 1")
	}
	if strings.TrimSpace(b.SourceScope) == "" {
		return fmt.Errorf("source_scope cannot be empty")
	}
	if len(b.Epics) == 0 {
		return fmt.Errorf("epics cannot be empty")
	}
	for _, epic := range b.Epics {
		if strings.TrimSpace(epic.ID) == "" {
			return fmt.Errorf("epic id cannot be empty")
		}
		if len(epic.Tasks) == 0 {
			return fmt.Errorf("epic %s must include tasks", epic.ID)
		}
	}
	return nil
}

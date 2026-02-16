package domain

import "time"

const CurrentSchemaVersion = 1

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusDone      TaskStatus = "done"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusBlocked   TaskStatus = "blocked"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Prompt       string     `json:"prompt,omitempty"`
	Dependencies []string   `json:"dependencies,omitempty"`
	Lane         string     `json:"lane"`
	Status       TaskStatus `json:"status"`
}

type Board struct {
	SchemaVersion int       `json:"schema_version"`
	SourceScope   string    `json:"source_scope"`
	GeneratedAt   time.Time `json:"generated_at"`
	Tasks         []Task    `json:"tasks"`
}

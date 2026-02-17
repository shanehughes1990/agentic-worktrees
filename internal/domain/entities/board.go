package entities

import "time"

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID           string
	Title        string
	Description  string
	Status       TaskStatus
	Dependencies []string
}

type Epic struct {
	ID           string
	Title        string
	Description  string
	Dependencies []string
	Tasks        []Task
}

type Board struct {
	ID        string
	Title     string
	Epics     []Epic
	CreatedAt time.Time
	UpdatedAt time.Time
}

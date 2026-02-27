package taskboard

import (
	"context"
	"strings"
	"time"
)

type RunCheckpoint struct {
	Name      string    `json:"name"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type RunState struct {
	RunID       string          `json:"run_id"`
	BoardID     string          `json:"board_id,omitempty"`
	Status      WorkflowStatus  `json:"status"`
	Message     string          `json:"message,omitempty"`
	Checkpoints []RunCheckpoint `json:"checkpoints,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (state *RunState) Normalize(runID string) {
	now := time.Now().UTC()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	state.RunID = strings.TrimSpace(runID)
}

type RunRepository interface {
	GetRunState(ctx context.Context, runID string) (*RunState, error)
	ListRunStates(ctx context.Context) ([]RunState, error)
	SaveRunState(ctx context.Context, runState *RunState) error
}

type JobState struct {
	JobID        string    `json:"job_id"`
	RunID        string    `json:"run_id"`
	TaskID       string    `json:"task_id,omitempty"`
	Attempt      int       `json:"attempt,omitempty"`
	Status       string    `json:"status"`
	FailureClass string    `json:"failure_class,omitempty"`
	ResultRef    string    `json:"result_ref,omitempty"`
	OutputRef    string    `json:"output_ref,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (state *JobState) Normalize(runID string, jobID string) {
	now := time.Now().UTC()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	state.RunID = strings.TrimSpace(runID)
	state.JobID = strings.TrimSpace(jobID)
}

type JobRepository interface {
	GetJobState(ctx context.Context, runID string, jobID string) (*JobState, error)
	ListJobStatesByRunID(ctx context.Context, runID string) ([]JobState, error)
	SaveJobState(ctx context.Context, jobState *JobState) error
}

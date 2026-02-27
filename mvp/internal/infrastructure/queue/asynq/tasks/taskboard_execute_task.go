package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TaskTypeTaskboardExecute = "taskboard.execute"

type TaskboardExecutePayload struct {
	BoardID        string `json:"board_id"`
	SourceBranch   string `json:"source_branch"`
	RepositoryRoot string `json:"repository_root"`
	MaxTasks       int    `json:"max_tasks,omitempty"`
	IdempotencyKey string `json:"idempotency_key"`
}

func NewTaskboardExecuteTask(payload TaskboardExecutePayload, options ...asynq.Option) (*asynq.Task, []asynq.Option, error) {
	if strings.TrimSpace(payload.BoardID) == "" {
		return nil, nil, fmt.Errorf("board_id is required")
	}
	if strings.TrimSpace(payload.SourceBranch) == "" {
		return nil, nil, fmt.Errorf("source_branch is required")
	}
	if strings.TrimSpace(payload.RepositoryRoot) == "" {
		return nil, nil, fmt.Errorf("repository_root is required")
	}
	if payload.MaxTasks < 0 {
		return nil, nil, fmt.Errorf("max_tasks cannot be negative")
	}
	if strings.TrimSpace(payload.IdempotencyKey) == "" {
		payload.IdempotencyKey = fmt.Sprintf("%s:%s", strings.TrimSpace(payload.BoardID), strings.TrimSpace(payload.SourceBranch))
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal taskboard execute payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeTaskboardExecute, body)
	opts := []asynq.Option{asynq.Queue(queueAgent), asynq.Retention(24 * time.Hour), asynq.TaskID(payload.IdempotencyKey), asynq.Unique(2 * time.Hour)}
	opts = append(opts, options...)
	return task, opts, nil
}

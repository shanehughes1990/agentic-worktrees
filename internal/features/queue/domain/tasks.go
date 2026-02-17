package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TypePlanBoard = "task.adk.plan_board"

type PlanBoardPayload struct {
	RunID          string `json:"run_id"`
	TaskID         string `json:"task_id"`
	IdempotencyKey string `json:"idempotency_key"`
	ScopePath      string `json:"scope_path"`
	OutPath        string `json:"out_path"`
}

func NewPlanBoardTask(payload PlanBoardPayload, queueName string) (*asynq.Task, error) {
	payload.RunID = strings.TrimSpace(payload.RunID)
	payload.TaskID = strings.TrimSpace(payload.TaskID)
	payload.ScopePath = strings.TrimSpace(payload.ScopePath)
	payload.OutPath = strings.TrimSpace(payload.OutPath)
	payload.IdempotencyKey = strings.TrimSpace(payload.IdempotencyKey)

	if payload.RunID == "" {
		return nil, fmt.Errorf("run_id cannot be empty")
	}
	if payload.TaskID == "" {
		return nil, fmt.Errorf("task_id cannot be empty")
	}
	if payload.ScopePath == "" {
		return nil, fmt.Errorf("scope_path cannot be empty")
	}
	if payload.OutPath == "" {
		return nil, fmt.Errorf("out_path cannot be empty")
	}
	if payload.IdempotencyKey == "" {
		payload.IdempotencyKey = fmt.Sprintf("%s:%s:%s", payload.RunID, payload.TaskID, payload.ScopePath)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	opts := []asynq.Option{asynq.Unique(10 * time.Minute)}
	if strings.TrimSpace(queueName) != "" {
		opts = append(opts, asynq.Queue(queueName))
	}

	return asynq.NewTask(TypePlanBoard, body, opts...), nil
}

func ParsePlanBoardPayload(task *asynq.Task) (PlanBoardPayload, error) {
	var payload PlanBoardPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return PlanBoardPayload{}, fmt.Errorf("invalid payload: %w", err)
	}

	payload.RunID = strings.TrimSpace(payload.RunID)
	payload.TaskID = strings.TrimSpace(payload.TaskID)
	payload.ScopePath = strings.TrimSpace(payload.ScopePath)
	payload.OutPath = strings.TrimSpace(payload.OutPath)
	payload.IdempotencyKey = strings.TrimSpace(payload.IdempotencyKey)

	if payload.RunID == "" || payload.TaskID == "" || payload.ScopePath == "" || payload.OutPath == "" {
		return PlanBoardPayload{}, fmt.Errorf("payload missing required fields")
	}
	if payload.IdempotencyKey == "" {
		payload.IdempotencyKey = fmt.Sprintf("%s:%s:%s", payload.RunID, payload.TaskID, payload.ScopePath)
	}

	return payload, nil
}

package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskTypeCopilotDecompose = "copilot.decompose"
	queueIngestion           = "ingestion"
	queueAgent               = "agent"
)

type CopilotDecomposePayload struct {
	RunID            string   `json:"run_id"`
	IdempotencyKey   string   `json:"idempotency_key,omitempty"`
	Prompt           string   `json:"prompt"`
	Model            string   `json:"model"`
	WorkingDirectory string   `json:"working_directory,omitempty"`
	SkillDirectories []string `json:"skill_directories,omitempty"`
	GithubToken      string   `json:"github_token,omitempty"`
	CLIPath          string   `json:"cli_path,omitempty"`
	CLIURL           string   `json:"cli_url,omitempty"`
}

func NewCopilotDecomposeTask(payload CopilotDecomposePayload, options ...asynq.Option) (*asynq.Task, []asynq.Option, error) {
	if strings.TrimSpace(payload.RunID) == "" {
		return nil, nil, fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(payload.Prompt) == "" {
		return nil, nil, fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(payload.IdempotencyKey) == "" {
		payload.IdempotencyKey = strings.TrimSpace(payload.RunID)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal copilot decompose payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeCopilotDecompose, body)
	opts := []asynq.Option{asynq.Queue(queueIngestion), asynq.Retention(24 * time.Hour), asynq.TaskID(payload.IdempotencyKey), asynq.Unique(2 * time.Hour)}
	opts = append(opts, options...)
	return task, opts, nil
}

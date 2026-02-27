package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type IngestionAgentPayload struct {
	RunID  string `json:"run_id"`
	Prompt string `json:"prompt"`
}

type IngestionAgentHandler struct{}

func NewIngestionAgentHandler() *IngestionAgentHandler {
	return &IngestionAgentHandler{}
}

func (handler *IngestionAgentHandler) Handle(ctx context.Context, job taskengine.Job) error {
	_ = ctx
	var payload IngestionAgentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode ingestion agent payload: %w", err)
	}
	if strings.TrimSpace(payload.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(payload.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	return nil
}

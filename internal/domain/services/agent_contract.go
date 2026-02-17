package services

import (
	"context"
	"fmt"
	"strings"
)

type AgentRequestMetadata struct {
	RunID string `json:"run_id"`
	JobID string `json:"job_id"`
	Model string `json:"model"`
}

func (m AgentRequestMetadata) Validate() error {
	if strings.TrimSpace(m.JobID) == "" {
		return fmt.Errorf("job_id is required")
	}
	return nil
}

type GenerateTaskBoardRequest struct {
	Metadata  AgentRequestMetadata    `json:"metadata"`
	Prompt    string                  `json:"prompt"`
	Documents []DocumentationSourceFile `json:"documents"`
}

func (r GenerateTaskBoardRequest) Validate() error {
	if err := r.Metadata.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(r.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if len(r.Documents) == 0 {
		return fmt.Errorf("documents are required")
	}
	for _, document := range r.Documents {
		if err := document.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type GenerateTaskBoardResult struct {
	BoardJSON string `json:"board_json"`
}

type AgentRunner interface {
	GenerateTaskBoard(ctx context.Context, request GenerateTaskBoardRequest) (GenerateTaskBoardResult, error)
}

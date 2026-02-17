package application

import (
	"context"
	"fmt"
	"strings"

	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
)

const (
	AsynqTaskTypeGenerateTaskBoard       = "task.generate_task_board"
	AsynqTaskTypeGenerateTaskBoardResult = "task.generate_task_board_result"
	DefaultCreateBoardModel              = "gpt-5.3-codex"
)

type GenerateTaskBoardInput struct {
	JobID         string `json:"job_id"`
	RunID         string `json:"run_id"`
	RootDirectory string `json:"root_directory"`
	MaxDepth      int    `json:"max_depth"`
	Prompt        string `json:"prompt"`
	Model         string `json:"model"`
}

func (i GenerateTaskBoardInput) Validate() error {
	if strings.TrimSpace(i.JobID) == "" {
		return fmt.Errorf("job_id is required")
	}
	if strings.TrimSpace(i.RootDirectory) == "" {
		return fmt.Errorf("root_directory is required")
	}
	if i.MaxDepth < 0 {
		return fmt.Errorf("max_depth must be zero or greater")
	}
	if strings.TrimSpace(i.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	return nil
}

type GenerateTaskBoardPayload struct {
	Metadata      domainservices.AgentRequestMetadata      `json:"metadata"`
	Prompt        string                                   `json:"prompt"`
	RootDirectory string                                   `json:"root_directory"`
	MaxDepth      int                                      `json:"max_depth"`
	Documents     []domainservices.DocumentationSourceFile `json:"documents"`
}

func (p GenerateTaskBoardPayload) Validate() error {
	request := domainservices.GenerateTaskBoardRequest{Metadata: p.Metadata, Prompt: p.Prompt, Documents: p.Documents}
	if err := request.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(p.RootDirectory) == "" {
		return fmt.Errorf("root_directory is required")
	}
	if p.MaxDepth < 0 {
		return fmt.Errorf("max_depth must be zero or greater")
	}
	return nil
}

type GenerateTaskBoardResultMessage struct {
	Metadata  domainservices.AgentRequestMetadata `json:"metadata"`
	BoardJSON string                              `json:"board_json"`
	Error     string                              `json:"error"`
}

func (m GenerateTaskBoardResultMessage) Validate() error {
	return m.Metadata.Validate()
}

type GenerateTaskBoardEnqueuer interface {
	EnqueueGenerateTaskBoard(ctx context.Context, payload GenerateTaskBoardPayload) (string, error)
}

type GenerateTaskBoardResultPublisher interface {
	EnqueueGenerateTaskBoardResult(ctx context.Context, result GenerateTaskBoardResultMessage) (string, error)
}

type PrepareGenerateTaskBoardCommand struct {
	loader domainservices.DocumentationFileLoader
}

func NewPrepareGenerateTaskBoardCommand(loader domainservices.DocumentationFileLoader) (*PrepareGenerateTaskBoardCommand, error) {
	if loader == nil {
		return nil, fmt.Errorf("documentation loader cannot be nil")
	}
	return &PrepareGenerateTaskBoardCommand{loader: loader}, nil
}

func (c *PrepareGenerateTaskBoardCommand) Execute(ctx context.Context, input GenerateTaskBoardInput) (GenerateTaskBoardPayload, error) {
	if c == nil {
		return GenerateTaskBoardPayload{}, fmt.Errorf("command cannot be nil")
	}
	if err := input.Validate(); err != nil {
		return GenerateTaskBoardPayload{}, err
	}

	documents, err := c.loader.LoadDocumentationFiles(ctx, input.RootDirectory, input.MaxDepth)
	if err != nil {
		return GenerateTaskBoardPayload{}, fmt.Errorf("load documentation files: %w", err)
	}
	if len(documents) == 0 {
		return GenerateTaskBoardPayload{}, fmt.Errorf("no documentation files found")
	}

	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = DefaultCreateBoardModel
	}

	payload := GenerateTaskBoardPayload{
		Metadata: domainservices.AgentRequestMetadata{
			RunID: input.RunID,
			JobID: input.JobID,
			Model: model,
		},
		Prompt:        input.Prompt,
		RootDirectory: input.RootDirectory,
		MaxDepth:      input.MaxDepth,
		Documents:     documents,
	}
	if err := payload.Validate(); err != nil {
		return GenerateTaskBoardPayload{}, err
	}
	return payload, nil
}

type ExecuteGenerateTaskBoardCommand struct {
	agentRunner domainservices.AgentRunner
}

func NewExecuteGenerateTaskBoardCommand(agentRunner domainservices.AgentRunner) (*ExecuteGenerateTaskBoardCommand, error) {
	if agentRunner == nil {
		return nil, fmt.Errorf("agent runner cannot be nil")
	}
	return &ExecuteGenerateTaskBoardCommand{agentRunner: agentRunner}, nil
}

func (c *ExecuteGenerateTaskBoardCommand) Execute(ctx context.Context, payload GenerateTaskBoardPayload) (domainservices.GenerateTaskBoardResult, error) {
	if c == nil {
		return domainservices.GenerateTaskBoardResult{}, fmt.Errorf("command cannot be nil")
	}
	if err := payload.Validate(); err != nil {
		return domainservices.GenerateTaskBoardResult{}, err
	}

	return c.agentRunner.GenerateTaskBoard(ctx, domainservices.GenerateTaskBoardRequest{
		Metadata:  payload.Metadata,
		Prompt:    payload.Prompt,
		Documents: payload.Documents,
	})
}

package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/application"
)

type GenerateTaskBoardHandler struct {
	executeCommand  *application.ExecuteGenerateTaskBoardCommand
	resultPublisher application.GenerateTaskBoardResultPublisher
}

func NewGenerateTaskBoardHandler(executeCommand *application.ExecuteGenerateTaskBoardCommand, resultPublisher application.GenerateTaskBoardResultPublisher) (*GenerateTaskBoardHandler, error) {
	if executeCommand == nil {
		return nil, fmt.Errorf("execute command cannot be nil")
	}
	if resultPublisher == nil {
		return nil, fmt.Errorf("result publisher cannot be nil")
	}
	return &GenerateTaskBoardHandler{executeCommand: executeCommand, resultPublisher: resultPublisher}, nil
}

func (h *GenerateTaskBoardHandler) Handle(ctx context.Context, task *asynq.Task) error {
	if h == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	var payload application.GenerateTaskBoardPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal task payload: %w", err)
	}

	result, executeErr := h.executeCommand.Execute(ctx, payload)
	message := application.GenerateTaskBoardResultMessage{Metadata: payload.Metadata}
	if executeErr != nil {
		message.Error = executeErr.Error()
	} else {
		message.BoardJSON = result.BoardJSON
	}

	if _, err := h.resultPublisher.EnqueueGenerateTaskBoardResult(ctx, message); err != nil {
		return fmt.Errorf("enqueue result message: %w", err)
	}
	return nil
}

package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/application"
)

type PersistGeneratedBoardResultHandler struct {
	persistCommand *application.PersistGenerateTaskBoardResultCommand
}

func NewPersistGeneratedBoardResultHandler(persistCommand *application.PersistGenerateTaskBoardResultCommand) (*PersistGeneratedBoardResultHandler, error) {
	if persistCommand == nil {
		return nil, fmt.Errorf("persist command cannot be nil")
	}
	return &PersistGeneratedBoardResultHandler{persistCommand: persistCommand}, nil
}

func (h *PersistGeneratedBoardResultHandler) Handle(ctx context.Context, task *asynq.Task) error {
	if h == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	var message application.GenerateTaskBoardResultMessage
	if err := json.Unmarshal(task.Payload(), &message); err != nil {
		return fmt.Errorf("unmarshal result payload: %w", err)
	}
	if err := h.persistCommand.Execute(ctx, message); err != nil {
		return fmt.Errorf("persist generated board result: %w", err)
	}
	return nil
}

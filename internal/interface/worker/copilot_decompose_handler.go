package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

type CopilotDecomposeHandler struct {
	decomposer appcopilot.Decomposer
	logger     *logrus.Logger
}

func NewCopilotDecomposeHandler(decomposer appcopilot.Decomposer, logger *logrus.Logger) *CopilotDecomposeHandler {
	return &CopilotDecomposeHandler{decomposer: decomposer, logger: logger}
}

func (handler *CopilotDecomposeHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != tasks.TaskTypeCopilotDecompose {
		return fmt.Errorf("unsupported task type: %s", task.Type())
	}

	var payload tasks.CopilotDecomposePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode copilot decompose payload: %w", err)
	}

	result, err := handler.decomposer.Decompose(ctx, appcopilot.DecomposeRequest{
		RunID:            payload.RunID,
		Prompt:           payload.Prompt,
		Model:            payload.Model,
		WorkingDirectory: payload.WorkingDirectory,
		SkillDirectories: payload.SkillDirectories,
		GitHubToken:      payload.GithubToken,
		CLIPath:          payload.CLIPath,
		CLIURL:           payload.CLIURL,
	})
	if err != nil {
		return fmt.Errorf("copilot decomposition failed: %w", err)
	}

	handler.logger.WithFields(logrus.Fields{
		"run_id":       payload.RunID,
		"session_id":   result.SessionID,
		"prompt_hash":  result.PromptHash,
		"response_len": len(strings.TrimSpace(result.Response)),
	}).Info("copilot decomposition completed")

	return nil
}

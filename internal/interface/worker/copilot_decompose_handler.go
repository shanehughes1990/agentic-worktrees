package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type CopilotDecomposeHandler struct {
	decomposer   appcopilot.Decomposer
	repository   apptaskboard.Repository
	workflowRepo apptaskboard.WorkflowRepository
	logger       *logrus.Logger
}

func NewCopilotDecomposeHandler(decomposer appcopilot.Decomposer, repository apptaskboard.Repository, workflowRepo apptaskboard.WorkflowRepository, logger *logrus.Logger) *CopilotDecomposeHandler {
	return &CopilotDecomposeHandler{
		decomposer:   decomposer,
		repository:   repository,
		workflowRepo: workflowRepo,
		logger:       logger,
	}
}

func (handler *CopilotDecomposeHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload tasks.CopilotDecomposePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode copilot payload: %w", err)
	}
	entry := handler.entry().WithFields(logrus.Fields{"event": "worker.copilot_decompose", "run_id": strings.TrimSpace(payload.RunID), "task_type": task.Type()})
	entry.Info("processing copilot decomposition task")

	runID := strings.TrimSpace(payload.RunID)
	if runID != "" {
		workflow := &apptaskboard.IngestionWorkflow{RunID: runID, Status: apptaskboard.WorkflowStatusRunning, Message: "copilot decomposition running"}
		workflow.Normalize(runID)
		_ = handler.workflowRepo.SaveWorkflow(ctx, workflow)
		entry.Info("workflow updated to running")
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
		userMessage := formatUserFacingDecomposeError(err)
		handler.saveFailureWorkflow(ctx, payload.RunID, userMessage)
		entry.WithError(err).WithField("user_message", userMessage).Error("copilot decomposition failed")
		if isTerminalCopilotFailure(err) {
			entry.WithError(err).Warn("terminal copilot failure detected; skipping retry")
			return fmt.Errorf("%w: copilot decompose: %v", asynq.SkipRetry, err)
		}
		return fmt.Errorf("copilot decompose: %w", err)
	}
	entry.WithFields(logrus.Fields{"session_id": result.SessionID, "model": result.Model}).Info("copilot decomposition completed")

	if task.ResultWriter() != nil {
		if _, writeErr := task.ResultWriter().Write([]byte(result.Response)); writeErr != nil {
			entry.WithError(writeErr).Warn("failed to write copilot stream to asynq task result")
		}
	}

	board, err := apptaskboard.BuildBoardFromResponse(payload.RunID, result.Response)
	if err != nil {
		handler.saveFailureWorkflow(ctx, payload.RunID, fmt.Sprintf("taskboard generation failed: %v", err))
		entry.WithError(err).Error("taskboard build failed")
		return fmt.Errorf("build taskboard from copilot response: %w", err)
	}

	if err := handler.repository.Save(ctx, board); err != nil {
		handler.saveFailureWorkflow(ctx, payload.RunID, fmt.Sprintf("taskboard persistence failed: %v", err))
		entry.WithError(err).Error("taskboard save failed")
		return fmt.Errorf("save taskboard: %w", err)
	}
	entry.WithField("board_id", board.BoardID).Info("taskboard saved")

	if runID != "" {
		workflow := &apptaskboard.IngestionWorkflow{
			RunID:   runID,
			Status:  apptaskboard.WorkflowStatusCompleted,
			Message: "taskboard created",
			Stream:  result.Response,
			BoardID: board.BoardID,
		}
		workflow.Normalize(runID)
		_ = handler.workflowRepo.SaveWorkflow(ctx, workflow)
		entry.Info("workflow updated to completed")
	}

	return nil
}

func (handler *CopilotDecomposeHandler) saveFailureWorkflow(ctx context.Context, runID string, message string) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return
	}
	workflow := &apptaskboard.IngestionWorkflow{
		RunID:   cleanRunID,
		Status:  apptaskboard.WorkflowStatusFailed,
		Message: message,
	}
	workflow.Normalize(cleanRunID)
	_ = handler.workflowRepo.SaveWorkflow(ctx, workflow)
}

func (handler *CopilotDecomposeHandler) entry() *logrus.Entry {
	if handler.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(handler.logger)
}

func formatUserFacingDecomposeError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "copilot decomposition failed"
	}
	lower := strings.ToLower(message)
	if strings.Contains(lower, "start copilot client") {
		return "Copilot failed to start. Run the dashboard auth command, verify Copilot entitlement, and check logs for startup diagnostics (version/auth status output)."
	}
	if strings.Contains(lower, "create copilot session") {
		return "Copilot session creation failed. Verify model/access settings and retry."
	}
	if strings.Contains(lower, "send decomposition prompt") {
		return "Copilot request failed while waiting for response. Retry and check network/auth state."
	}
	return fmt.Sprintf("copilot decomposition failed: %s", message)
}

func isTerminalCopilotFailure(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	terminalIndicators := []string{
		"start copilot client",
		"authentication failed",
		"auth status",
		"entitlement",
		"executable file not found",
		"cli process exited",
	}
	for _, indicator := range terminalIndicators {
		if strings.Contains(message, indicator) {
			return true
		}
	}
	return false
}

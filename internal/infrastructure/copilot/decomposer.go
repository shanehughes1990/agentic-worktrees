package copilot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
	"github.com/sirupsen/logrus"
)

type Decomposer struct {
	client *Client
	logger *logrus.Logger
}

func NewDecomposer(config ClientConfig, logger *logrus.Logger) *Decomposer {
	return &Decomposer{client: NewClient(config, logger), logger: logger}
}

func (decomposer *Decomposer) Decompose(ctx context.Context, request appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	prompt := strings.TrimSpace(request.Prompt)
	if prompt == "" {
		return appcopilot.DecomposeResult{}, fmt.Errorf("prompt is required")
	}

	entry := decomposer.entry().WithFields(logrus.Fields{
		"event":             "copilot.decompose",
		"run_id":            strings.TrimSpace(request.RunID),
		"task_id":           strings.TrimSpace(request.TaskID),
		"queue_task_id":     strings.TrimSpace(request.QueueTaskID),
		"correlation_id":    strings.TrimSpace(request.CorrelationID),
		"working_directory": strings.TrimSpace(request.WorkingDirectory),
	})
	entry.Info("copilot decomposition request received")

	sessionID, response, usedModel, err := decomposer.client.RunPrompt(ctx, request.RunID, request.TaskID, request.QueueTaskID, request.CorrelationID, request.Model, request.ResumeSessionID, request.WorkingDirectory, request.SkillDirectories, prompt)
	if err != nil {
		entry.WithError(err).Error("copilot decomposition failed")
		return appcopilot.DecomposeResult{RunID: request.RunID, SessionID: sessionID, Model: usedModel}, err
	}

	hash := sha256.Sum256([]byte(prompt))
	result := appcopilot.DecomposeResult{
		RunID:      request.RunID,
		SessionID:  sessionID,
		Response:   response,
		Model:      usedModel,
		PromptHash: hex.EncodeToString(hash[:]),
	}
	entry.WithFields(logrus.Fields{"session_id": sessionID, "model": usedModel, "response_bytes": len(response)}).Info("copilot decomposition completed")
	return result, nil
}

func (decomposer *Decomposer) entry() *logrus.Entry {
	if decomposer.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(decomposer.logger)
}

package asynq

import (
	"context"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	infracopilot "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/copilot"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/sirupsen/logrus"
)

type TaskboardIngestionDispatcher struct {
	client *Client
	config infracopilot.ClientConfig
	logger *logrus.Logger
}

func NewTaskboardIngestionDispatcher(client *Client, config infracopilot.ClientConfig, logger *logrus.Logger) *TaskboardIngestionDispatcher {
	return &TaskboardIngestionDispatcher{
		client: client,
		config: config.Normalized(),
		logger: logger,
	}
}

func (dispatcher *TaskboardIngestionDispatcher) EnqueueIngestion(ctx context.Context, job apptaskboard.IngestionJob) (string, error) {
	entry := dispatcher.entry().WithFields(logrus.Fields{
		"event":             "ingestion.enqueue",
		"run_id":            job.RunID,
		"working_directory": job.WorkingDirectory,
		"model":             job.Model,
	})
	entry.Info("enqueueing copilot decomposition task")

	taskInfo, err := dispatcher.client.EnqueueCopilotDecompose(ctx, tasks.CopilotDecomposePayload{
		RunID:            job.RunID,
		Prompt:           job.Prompt,
		Model:            job.Model,
		WorkingDirectory: job.WorkingDirectory,
		SkillDirectories: dispatcher.config.SkillDirectories,
		GithubToken:      dispatcher.config.GitHubToken,
		CLIPath:          dispatcher.config.CLIPath,
		CLIURL:           dispatcher.config.CLIURL,
	})
	if err != nil {
		entry.WithError(err).Error("failed to enqueue copilot decomposition task")
		return "", err
	}
	if taskInfo == nil {
		entry.Warn("enqueue returned nil task info")
		return "", nil
	}
	entry.WithField("task_id", taskInfo.ID).Info("copilot decomposition task enqueued")
	return taskInfo.ID, nil
}

func (dispatcher *TaskboardIngestionDispatcher) entry() *logrus.Entry {
	if dispatcher.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(dispatcher.logger)
}

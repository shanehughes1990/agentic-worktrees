package gitflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type SourceBranchReader interface {
	CurrentBranch(ctx context.Context, repositoryRoot string) (string, error)
}

type SourceBranchService struct {
	reader SourceBranchReader
	logger *logrus.Logger
}

func NewSourceBranchService(reader SourceBranchReader, loggers ...*logrus.Logger) *SourceBranchService {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &SourceBranchService{reader: reader, logger: logger}
}

func (service *SourceBranchService) Resolve(ctx context.Context, repositoryRoot string) (string, error) {
	entry := service.entry().WithFields(logrus.Fields{"event": "gitflow.source_branch.resolve", "repository_root": strings.TrimSpace(repositoryRoot)})
	if service == nil || service.reader == nil {
		entry.Error("source branch reader is required")
		return "", fmt.Errorf("source branch reader is required")
	}
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	if cleanRepositoryRoot == "" {
		entry.Error("repository_root is required")
		return "", fmt.Errorf("repository_root is required")
	}

	branch, err := service.reader.CurrentBranch(ctx, cleanRepositoryRoot)
	if err != nil {
		entry.WithError(err).Error("failed to resolve current branch")
		return "", fmt.Errorf("resolve current branch: %w", err)
	}
	cleanBranch := strings.TrimSpace(branch)
	if cleanBranch == "" {
		entry.Error("current branch is empty")
		return "", fmt.Errorf("current branch is empty")
	}
	entry.WithField("source_branch", cleanBranch).Info("resolved current source branch")
	return cleanBranch, nil
}

func (service *SourceBranchService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}

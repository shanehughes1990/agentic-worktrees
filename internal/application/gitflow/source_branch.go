package gitflow

import (
	"context"
	"fmt"
	"strings"
)

type SourceBranchReader interface {
	CurrentBranch(ctx context.Context, repositoryRoot string) (string, error)
}

type SourceBranchService struct {
	reader SourceBranchReader
}

func NewSourceBranchService(reader SourceBranchReader) *SourceBranchService {
	return &SourceBranchService{reader: reader}
}

func (service *SourceBranchService) Resolve(ctx context.Context, repositoryRoot string) (string, error) {
	if service == nil || service.reader == nil {
		return "", fmt.Errorf("source branch reader is required")
	}
	cleanRepositoryRoot := strings.TrimSpace(repositoryRoot)
	if cleanRepositoryRoot == "" {
		return "", fmt.Errorf("repository_root is required")
	}

	branch, err := service.reader.CurrentBranch(ctx, cleanRepositoryRoot)
	if err != nil {
		return "", fmt.Errorf("resolve current branch: %w", err)
	}
	cleanBranch := strings.TrimSpace(branch)
	if cleanBranch == "" {
		return "", fmt.Errorf("current branch is empty")
	}
	return cleanBranch, nil
}

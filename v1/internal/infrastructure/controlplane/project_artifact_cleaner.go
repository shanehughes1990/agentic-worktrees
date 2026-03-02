package controlplane

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ProjectArtifactCleanerConfig struct {
	RepositorySourcePath string
	WorktreesPath        string
	TrackerPath          string
}

type ProjectArtifactCleaner struct {
	repositorySourcePath string
	worktreesPath        string
	trackerPath          string
}

func NewProjectArtifactCleaner(config ProjectArtifactCleanerConfig) (*ProjectArtifactCleaner, error) {
	repositorySourcePath := strings.TrimSpace(config.RepositorySourcePath)
	worktreesPath := strings.TrimSpace(config.WorktreesPath)
	trackerPath := strings.TrimSpace(config.TrackerPath)
	if repositorySourcePath == "" {
		return nil, fmt.Errorf("repository source path is required")
	}
	if worktreesPath == "" {
		return nil, fmt.Errorf("worktrees path is required")
	}
	if trackerPath == "" {
		return nil, fmt.Errorf("tracker path is required")
	}
	return &ProjectArtifactCleaner{
		repositorySourcePath: filepath.Clean(repositorySourcePath),
		worktreesPath:        filepath.Clean(worktreesPath),
		trackerPath:          filepath.Clean(trackerPath),
	}, nil
}

func (cleaner *ProjectArtifactCleaner) CleanupProjectArtifacts(ctx context.Context, setup applicationcontrolplane.ProjectSetup) error {
	if cleaner == nil {
		return fmt.Errorf("project artifact cleaner is not configured")
	}
	projectID := strings.TrimSpace(setup.ProjectID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	for _, artifactPath := range cleaner.projectCandidatePaths(projectID, setup.RepositoryURL) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.RemoveAll(artifactPath); err != nil {
			return fmt.Errorf("remove project artifact path %q: %w", artifactPath, err)
		}
	}
	if err := cleaner.removeTaggedWorktreeArtifacts(ctx, projectID); err != nil {
		return err
	}
	return nil
}

func (cleaner *ProjectArtifactCleaner) projectCandidatePaths(projectID string, repositoryURL string) []string {
	paths := []string{
		filepath.Join(cleaner.worktreesPath, projectID),
		filepath.Join(cleaner.trackerPath, projectID),
		filepath.Join(cleaner.trackerPath, projectID+".json"),
	}
	if owner, repository, ok := repositoryOwnerAndName(repositoryURL); ok {
		paths = append(paths,
			filepath.Join(cleaner.repositorySourcePath, owner, repository),
			filepath.Join(cleaner.repositorySourcePath, repository),
		)
	}
	return uniquePaths(paths)
}

func (cleaner *ProjectArtifactCleaner) removeTaggedWorktreeArtifacts(ctx context.Context, projectID string) error {
	entries, err := os.ReadDir(cleaner.worktreesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read worktrees path: %w", err)
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(name), strings.ToLower(projectID)) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(cleaner.worktreesPath, name)); err != nil {
			return fmt.Errorf("remove tagged worktree artifact %q: %w", name, err)
		}
	}
	return nil
}

func repositoryOwnerAndName(repositoryURL string) (string, string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(repositoryURL))
	if err != nil {
		return "", "", false
	}
	segments := strings.Split(strings.Trim(strings.TrimSpace(parsed.Path), "/"), "/")
	if len(segments) < 2 {
		return "", "", false
	}
	owner := strings.TrimSpace(segments[0])
	repository := strings.TrimSpace(strings.TrimSuffix(segments[1], ".git"))
	if owner == "" || repository == "" {
		return "", "", false
	}
	return owner, repository, true
}

func uniquePaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, candidate := range paths {
		cleaned := filepath.Clean(strings.TrimSpace(candidate))
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		result = append(result, cleaned)
	}
	return result
}

var _ applicationcontrolplane.ProjectCleanupManager = (*ProjectArtifactCleaner)(nil)

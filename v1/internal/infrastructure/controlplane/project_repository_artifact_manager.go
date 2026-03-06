package controlplane

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProjectRepositoryArtifactManagerConfig struct {
	ProjectsPath string
	GitBinary    string
}

type ProjectRepositoryArtifactManager struct {
	projectsPath string
	gitBinary    string
}

func NewProjectRepositoryArtifactManager(config ProjectRepositoryArtifactManagerConfig) (*ProjectRepositoryArtifactManager, error) {
	projectsPath := strings.TrimSpace(config.ProjectsPath)
	if projectsPath == "" {
		return nil, fmt.Errorf("projects path is required")
	}
	gitBinary := strings.TrimSpace(config.GitBinary)
	if gitBinary == "" {
		gitBinary = "git"
	}
	return &ProjectRepositoryArtifactManager{projectsPath: filepath.Clean(projectsPath), gitBinary: gitBinary}, nil
}

func (manager *ProjectRepositoryArtifactManager) ReconcileProjectRepositories(ctx context.Context, previous *applicationcontrolplane.ProjectSetup, current applicationcontrolplane.ProjectSetup) error {
	if manager == nil {
		return fmt.Errorf("project repository artifact manager is not configured")
	}
	projectID := strings.TrimSpace(current.ProjectID)
	if projectID == "" && previous != nil {
		projectID = strings.TrimSpace(previous.ProjectID)
	}
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	projectRepositoriesDir := filepath.Join(manager.projectsPath, projectID, "repositories")
	if err := os.MkdirAll(projectRepositoriesDir, 0o755); err != nil {
		return fmt.Errorf("create project repositories dir %q: %w", projectRepositoriesDir, err)
	}

	scmByID := map[string]applicationcontrolplane.ProjectSCM{}
	for _, scm := range current.SCMs {
		scmID := strings.TrimSpace(scm.SCMID)
		if scmID == "" {
			continue
		}
		scmByID[scmID] = scm
	}
	if previous != nil {
		for _, scm := range previous.SCMs {
			scmID := strings.TrimSpace(scm.SCMID)
			if scmID == "" {
				continue
			}
			if _, exists := scmByID[scmID]; !exists {
				scmByID[scmID] = scm
			}
		}
	}

	desiredDirs := make(map[string]applicationcontrolplane.ProjectRepository, len(current.Repositories))
	for _, repository := range current.Repositories {
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryDir := repositoryDirectoryName(repositoryID, repositoryURL)
		if repositoryDir == "" {
			continue
		}
		desiredDirs[repositoryDir] = repository
	}

	for repositoryDir, repository := range desiredDirs {
		if err := ctx.Err(); err != nil {
			return err
		}
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		scm := scmByID[strings.TrimSpace(repository.SCMID)]
		remoteURL := resolveRepositoryRemoteURL(repositoryURL, scm)
		targetPath := filepath.Join(projectRepositoriesDir, repositoryDir)
		if stat, statErr := os.Stat(targetPath); statErr == nil && stat.IsDir() {
			if err := manager.runGit(ctx, targetPath, "remote", "set-url", "origin", remoteURL); err != nil {
				return fmt.Errorf("set origin for repository %q: %w", strings.TrimSpace(repository.RepositoryID), err)
			}
			if err := manager.runGit(ctx, targetPath, "fetch", "--all", "--prune"); err != nil {
				return fmt.Errorf("fetch repository %q: %w", strings.TrimSpace(repository.RepositoryID), err)
			}
			defaultBranch, branchErr := manager.originDefaultBranch(ctx, targetPath)
			if branchErr == nil && strings.TrimSpace(defaultBranch) != "" {
				if err := manager.runGit(ctx, targetPath, "checkout", "-B", defaultBranch, "origin/"+defaultBranch); err != nil {
					return fmt.Errorf("checkout default branch %q for repository %q: %w", defaultBranch, strings.TrimSpace(repository.RepositoryID), err)
				}
				if err := manager.runGit(ctx, targetPath, "reset", "--hard", "origin/"+defaultBranch); err != nil {
					return fmt.Errorf("reset repository %q to origin/%s: %w", strings.TrimSpace(repository.RepositoryID), defaultBranch, err)
				}
			}
			continue
		}
		if err := manager.runGit(ctx, projectRepositoriesDir, "clone", remoteURL, repositoryDir); err != nil {
			return fmt.Errorf("clone repository %q: %w", strings.TrimSpace(repository.RepositoryID), err)
		}
	}

	entries, err := os.ReadDir(projectRepositoriesDir)
	if err != nil {
		return fmt.Errorf("list project repositories dir %q: %w", projectRepositoriesDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		if _, keep := desiredDirs[name]; keep {
			continue
		}
		stalePath := filepath.Join(projectRepositoriesDir, name)
		if err := os.RemoveAll(stalePath); err != nil {
			return fmt.Errorf("remove stale project repository %q: %w", stalePath, err)
		}
	}
	return nil
}

func (manager *ProjectRepositoryArtifactManager) originDefaultBranch(ctx context.Context, repositoryPath string) (string, error) {
	output, err := manager.runGitOutput(ctx, repositoryPath, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "", fmt.Errorf("origin HEAD is empty")
	}
	return strings.TrimPrefix(trimmed, "origin/"), nil
}

func (manager *ProjectRepositoryArtifactManager) runGit(ctx context.Context, workingDirectory string, args ...string) error {
	_, err := manager.runGitOutput(ctx, workingDirectory, args...)
	return err
}

func (manager *ProjectRepositoryArtifactManager) runGitOutput(ctx context.Context, workingDirectory string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, manager.gitBinary, args...)
	command.Dir = workingDirectory
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w (stdout=%s stderr=%s)", strings.Join(args, " "), err, strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String()))
	}
	return stdoutBuffer.String(), nil
}

func resolveRepositoryRemoteURL(repositoryURL string, scm applicationcontrolplane.ProjectSCM) string {
	trimmedURL := strings.TrimSpace(repositoryURL)
	if trimmedURL == "" {
		return trimmedURL
	}
	if strings.TrimSpace(scm.SCMProvider) != "github" || strings.TrimSpace(scm.SCMToken) == "" {
		return trimmedURL
	}
	parsed, err := url.Parse(trimmedURL)
	if err != nil {
		return trimmedURL
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return trimmedURL
	}
	parsed.User = url.UserPassword("x-access-token", strings.TrimSpace(scm.SCMToken))
	return parsed.String()
}

func repositoryDirectoryName(repositoryID string, repositoryURL string) string {
	trimmedURL := strings.TrimSpace(repositoryURL)
	if parsedURL, err := url.Parse(trimmedURL); err == nil {
		pathParts := strings.Split(strings.Trim(strings.TrimSpace(parsedURL.Path), "/"), "/")
		if len(pathParts) >= 2 {
			repoName := strings.TrimSpace(strings.TrimSuffix(pathParts[len(pathParts)-1], ".git"))
			if repoName != "" {
				return sanitizeDirectoryName(repoName)
			}
		}
	}
	if strings.TrimSpace(repositoryID) != "" {
		return sanitizeDirectoryName(repositoryID)
	}
	return "repository"
}

func sanitizeDirectoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "repository"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	sanitized := strings.Trim(replacer.Replace(trimmed), "-.")
	if sanitized == "" {
		return "repository"
	}
	return sanitized
}

var _ applicationcontrolplane.ProjectRepositoryArtifactManager = (*ProjectRepositoryArtifactManager)(nil)

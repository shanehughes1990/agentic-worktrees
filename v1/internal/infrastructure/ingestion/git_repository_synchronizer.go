package ingestion

import (
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"agentic-orchestrator/internal/domain/failures"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitRepositorySynchronizer struct {
	gitBinaryPath   string
	projectsRootDir string
}

func NewGitRepositorySynchronizer(gitBinaryPath string, projectsRootDir string) (*GitRepositorySynchronizer, error) {
	path := strings.TrimSpace(gitBinaryPath)
	if path == "" {
		path = "git"
	}
	rootDir := strings.TrimSpace(projectsRootDir)
	if rootDir == "" {
		return nil, failures.WrapTerminal(fmt.Errorf("projects_root_dir is required"))
	}
	return &GitRepositorySynchronizer{gitBinaryPath: path, projectsRootDir: rootDir}, nil
}

func (synchronizer *GitRepositorySynchronizer) Sync(ctx context.Context, projectID string, sandboxDir string, sourceBranch string, sourceRepositories []applicationingestion.SourceRepository) error {
	if synchronizer == nil {
		return failures.WrapTerminal(fmt.Errorf("git repository synchronizer is not initialized"))
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return failures.WrapTerminal(fmt.Errorf("project_id is required"))
	}
	cleanSandboxDir := strings.TrimSpace(sandboxDir)
	if cleanSandboxDir == "" {
		return failures.WrapTerminal(fmt.Errorf("sandbox_dir is required"))
	}
	cleanBranch := strings.TrimSpace(sourceBranch)
	if cleanBranch == "" {
		cleanBranch = "main"
	}
	repositoriesDir := filepath.Join(cleanSandboxDir, "repos")
	if err := os.MkdirAll(repositoriesDir, 0o755); err != nil {
		return failures.WrapTransient(fmt.Errorf("create repositories directory: %w", err))
	}

	for _, repository := range sourceRepositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		repositoryBranch := strings.TrimSpace(repository.SourceBranch)
		if repositoryBranch == "" {
			repositoryBranch = cleanBranch
		}
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		repositoryDirName := repositoryDirectoryName(repositoryID, repositoryURL)
		targetDirectory := filepath.Join(repositoriesDir, repositoryDirName)
		localRepositoryPath, resolveErr := synchronizer.resolveLocalRepositoryPath(cleanProjectID, repositoryDirName)
		if resolveErr != nil {
			if err := synchronizer.cloneFromOrigin(ctx, repositoriesDir, repositoryURL, repositoryBranch, targetDirectory); err != nil {
				return failures.WrapTransient(fmt.Errorf("sync repository %q from origin fallback: %w", repositoryID, err))
			}
			continue
		}
		if err := os.RemoveAll(targetDirectory); err != nil {
			return failures.WrapTransient(fmt.Errorf("reset sandbox repository directory %s: %w", targetDirectory, err))
		}
		if err := copyDirectory(localRepositoryPath, targetDirectory); err != nil {
			return failures.WrapTransient(fmt.Errorf("copy repository %q from local source cache: %w", repositoryID, err))
		}
		if err := synchronizer.runGit(ctx, targetDirectory, "remote", "set-url", "origin", repositoryURL); err != nil {
			return failures.WrapTransient(fmt.Errorf("configure origin for repository %q: %w", repositoryID, err))
		}
		if err := synchronizer.runGit(ctx, targetDirectory, "fetch", "--all", "--prune"); err != nil {
			return failures.WrapTransient(fmt.Errorf("sync all branches with origin for repository %q: %w", repositoryID, err))
		}
		if err := synchronizer.runGit(ctx, targetDirectory, "rev-parse", "--verify", "origin/"+repositoryBranch); err != nil {
			return failures.WrapTransient(fmt.Errorf("origin branch %q not found for repository %q: %w", repositoryBranch, repositoryID, err))
		}
		if err := synchronizer.runGit(ctx, targetDirectory, "checkout", "-B", repositoryBranch, "origin/"+repositoryBranch); err != nil {
			return failures.WrapTransient(fmt.Errorf("checkout branch %q for repository %q: %w", repositoryBranch, repositoryID, err))
		}
		if err := synchronizer.runGit(ctx, targetDirectory, "reset", "--hard", "origin/"+repositoryBranch); err != nil {
			return failures.WrapTransient(fmt.Errorf("reset repository %q to origin/%s: %w", repositoryID, repositoryBranch, err))
		}
	}

	return nil
}

func (synchronizer *GitRepositorySynchronizer) cloneFromOrigin(ctx context.Context, repositoriesDir string, repositoryURL string, repositoryBranch string, targetDirectory string) error {
	if err := os.RemoveAll(targetDirectory); err != nil {
		return fmt.Errorf("reset sandbox repository directory %s: %w", targetDirectory, err)
	}
	if err := synchronizer.runGit(ctx, repositoriesDir, "clone", "--branch", repositoryBranch, "--single-branch", repositoryURL, targetDirectory); err == nil {
		return nil
	}
	if err := synchronizer.runGit(ctx, repositoriesDir, "clone", repositoryURL, targetDirectory); err != nil {
		return fmt.Errorf("clone repository %q into %s: %w", repositoryURL, targetDirectory, err)
	}
	return nil
}

func (synchronizer *GitRepositorySynchronizer) resolveLocalRepositoryPath(projectID string, repositoryDirName string) (string, error) {
	if synchronizer == nil {
		return "", fmt.Errorf("git repository synchronizer is not initialized")
	}
	projectPath := filepath.Join(synchronizer.projectsRootDir, strings.TrimSpace(projectID), "repositories", strings.TrimSpace(repositoryDirName))
	if _, err := os.Stat(projectPath); err == nil {
		return projectPath, nil
	}
	unscopedPath := filepath.Join(synchronizer.projectsRootDir, "unscoped", "repositories", strings.TrimSpace(repositoryDirName))
	if _, err := os.Stat(unscopedPath); err == nil {
		return unscopedPath, nil
	}
	return "", fmt.Errorf("checked %s and %s", projectPath, unscopedPath)
}

func repositoryDirectoryName(repositoryID string, repositoryURL string) string {
	trimmedID := strings.TrimSpace(repositoryID)
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
	if trimmedID != "" {
		return sanitizeDirectoryName(trimmedID)
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

func (synchronizer *GitRepositorySynchronizer) runGit(ctx context.Context, workingDirectory string, args ...string) error {
	command := exec.CommandContext(ctx, synchronizer.gitBinaryPath, args...)
	command.Dir = workingDirectory
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer
	if err := command.Run(); err != nil {
		return fmt.Errorf("git %s: %w (stdout=%s stderr=%s)", strings.Join(args, " "), err, strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String()))
	}
	return nil
}

func copyDirectory(sourcePath string, destinationPath string) error {
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(sourcePath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		targetPath := filepath.Join(destinationPath, relativePath)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkDestination, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkDestination, targetPath)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()
		targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()
		_, err = io.Copy(targetFile, sourceFile)
		return err
	})
}

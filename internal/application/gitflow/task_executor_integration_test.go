package gitflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

type writingDecomposer struct{}

type integrationGitPort struct{}

func (decomposer *writingDecomposer) Decompose(_ context.Context, request appcopilot.DecomposeRequest) (appcopilot.DecomposeResult, error) {
	filePath := filepath.Join(request.WorkingDirectory, "feature.txt")
	if err := os.WriteFile(filePath, []byte("merged from task worktree\n"), 0o644); err != nil {
		return appcopilot.DecomposeResult{}, err
	}
	return appcopilot.DecomposeResult{RunID: request.RunID, SessionID: "session-e2e"}, nil
}

func (port *integrationGitPort) CreateTaskWorktree(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) error {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))
	if err := os.MkdirAll(filepath.Dir(absoluteWorktreePath), 0o755); err != nil {
		return err
	}
	_, _ = runGit(ctx, repositoryRoot, "worktree", "remove", "--force", absoluteWorktreePath)
	_, err := runGit(ctx, repositoryRoot, "worktree", "add", "-B", taskBranch, absoluteWorktreePath, sourceBranch)
	return err
}

func (port *integrationGitPort) MergeTaskBranch(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string) (MergeAttempt, error) {
	if _, err := runGit(ctx, repositoryRoot, "checkout", sourceBranch); err != nil {
		return MergeAttempt{}, err
	}
	if _, err := runGit(ctx, repositoryRoot, "merge", "--no-ff", "--no-commit", taskBranch); err != nil {
		return MergeAttempt{}, err
	}
	out, err := runGit(ctx, repositoryRoot, "diff", "--cached", "--name-only")
	if err != nil {
		return MergeAttempt{}, err
	}
	if strings.TrimSpace(out) == "" {
		return MergeAttempt{NoChanges: true}, nil
	}
	if _, err := runGit(ctx, repositoryRoot, "commit", "-m", fmt.Sprintf("Merge %s into %s", taskBranch, sourceBranch)); err != nil {
		return MergeAttempt{}, err
	}
	return MergeAttempt{}, nil
}

func (port *integrationGitPort) ResolveConflicts(context.Context, string, []string, string) error {
	return nil
}

func (port *integrationGitPort) StageAll(ctx context.Context, repositoryRoot string) error {
	_, err := runGit(ctx, repositoryRoot, "add", "-A")
	return err
}

func (port *integrationGitPort) Commit(ctx context.Context, repositoryRoot string, message string) error {
	out, err := runGit(ctx, repositoryRoot, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "" {
		return nil
	}
	_, err = runGit(ctx, repositoryRoot, "commit", "-m", strings.TrimSpace(message))
	return err
}

func (port *integrationGitPort) CleanupTaskWorktree(ctx context.Context, repositoryRoot string, worktreePath string, taskBranch string) error {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))
	if _, err := runGit(ctx, repositoryRoot, "worktree", "remove", "--force", absoluteWorktreePath); err != nil {
		return err
	}
	_, _ = runGit(ctx, repositoryRoot, "branch", "-D", taskBranch)
	return nil
}

func (port *integrationGitPort) CleanupRunArtifacts(context.Context, string, string) error {
	return nil
}

func TestTaskExecutorExecuteTaskMergesTaskWorktreeChangesBackToSourceBranch(t *testing.T) {
	tempDirectory := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	runGitCommand(t, ctx, tempDirectory, "init")
	runGitCommand(t, ctx, tempDirectory, "config", "user.email", "test@example.com")
	runGitCommand(t, ctx, tempDirectory, "config", "user.name", "tester")

	seedFilePath := filepath.Join(tempDirectory, "README.md")
	if err := os.WriteFile(seedFilePath, []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCommand(t, ctx, tempDirectory, "add", "README.md")
	runGitCommand(t, ctx, tempDirectory, "commit", "-m", "seed")
	defaultBranch := strings.TrimSpace(runGitCommandOutput(t, ctx, tempDirectory, "branch", "--show-current"))
	if defaultBranch == "" {
		defaultBranch = "master"
	}

	executor := NewTaskExecutor(&integrationGitPort{}, &writingDecomposer{})
	result, err := executor.ExecuteTask(ctx, TaskExecutionRequest{
		BoardID:        "board-1",
		RunID:          "run-1",
		TaskID:         "task-1",
		TaskTitle:      "Add feature file",
		TaskDetail:     "Create feature.txt",
		SourceBranch:   defaultBranch,
		RepositoryRoot: tempDirectory,
	})
	if err != nil {
		t.Fatalf("execute task: %v", err)
	}
	if result.Status != "merged" {
		t.Fatalf("expected merged status, got %s (reason=%s)", result.Status, result.Reason)
	}

	featureFilePath := filepath.Join(tempDirectory, "feature.txt")
	content, readErr := os.ReadFile(featureFilePath)
	if readErr != nil {
		t.Fatalf("read merged feature file: %v", readErr)
	}
	if !strings.Contains(string(content), "merged from task worktree") {
		t.Fatalf("expected merged feature content, got %q", string(content))
	}

	worktreePath := filepath.Join(tempDirectory, ".worktree", "run-1-task-1")
	if _, statErr := os.Stat(worktreePath); statErr == nil {
		t.Fatalf("expected task worktree to be cleaned up: %s", worktreePath)
	}

	logText := runGitCommandOutput(t, ctx, tempDirectory, "log", "--oneline", "-1")
	if !strings.Contains(strings.ToLower(logText), "merge task/run-1/task-1") {
		t.Fatalf("expected merge commit in source branch log, got %q", strings.TrimSpace(logText))
	}
}

func runGitCommand(t *testing.T, ctx context.Context, repositoryRoot string, args ...string) {
	t.Helper()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", repositoryRoot}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
}

func runGitCommandOutput(t *testing.T, ctx context.Context, repositoryRoot string, args ...string) string {
	t.Helper()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", repositoryRoot}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return string(output)
}

func runGit(ctx context.Context, repositoryRoot string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", repositoryRoot}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

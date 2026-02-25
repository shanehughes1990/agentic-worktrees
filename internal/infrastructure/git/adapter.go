package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	"github.com/sirupsen/logrus"
)

type Adapter struct {
	logger *logrus.Logger
}

func NewAdapter(logger *logrus.Logger) *Adapter {
	return &Adapter{logger: logger}
}

func (adapter *Adapter) CreateTaskWorktree(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) error {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))
	if err := os.MkdirAll(filepath.Dir(absoluteWorktreePath), 0o755); err != nil {
		return appgitflow.WrapTerminal(fmt.Errorf("create worktree parent directory: %w", err))
	}

	_, _ = adapter.runGit(ctx, repositoryRoot, "worktree", "remove", "--force", absoluteWorktreePath)

	_, err := adapter.runGit(ctx, repositoryRoot, "worktree", "add", "-B", taskBranch, absoluteWorktreePath, sourceBranch)
	if err != nil {
		return err
	}
	return nil
}

func (adapter *Adapter) MergeTaskBranch(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string) (appgitflow.MergeAttempt, error) {
	if _, err := adapter.runGit(ctx, repositoryRoot, "checkout", sourceBranch); err != nil {
		return appgitflow.MergeAttempt{}, err
	}

	_, err := adapter.runGit(ctx, repositoryRoot, "merge", "--no-ff", "--no-commit", taskBranch)
	if err != nil {
		conflictFiles, conflictErr := adapter.conflictFiles(ctx, repositoryRoot)
		if conflictErr != nil {
			return appgitflow.MergeAttempt{}, conflictErr
		}
		if len(conflictFiles) > 0 {
			return appgitflow.MergeAttempt{ConflictFiles: conflictFiles}, nil
		}
		return appgitflow.MergeAttempt{}, err
	}

	hasStagedChanges, err := adapter.hasStagedChanges(ctx, repositoryRoot)
	if err != nil {
		return appgitflow.MergeAttempt{}, err
	}
	if hasStagedChanges {
		if _, commitErr := adapter.runGit(ctx, repositoryRoot, "commit", "-m", fmt.Sprintf("Merge %s into %s", taskBranch, sourceBranch)); commitErr != nil {
			return appgitflow.MergeAttempt{}, commitErr
		}
		return appgitflow.MergeAttempt{}, nil
	}

	return appgitflow.MergeAttempt{NoChanges: true}, nil
}

func (adapter *Adapter) ResolveConflicts(ctx context.Context, repositoryRoot string, conflictFiles []string, copilotAdvice string) error {
	_ = strings.TrimSpace(copilotAdvice)
	if len(conflictFiles) == 0 {
		return appgitflow.WrapTerminal(fmt.Errorf("conflict_files is required"))
	}
	remainingConflictFiles, err := adapter.conflictFiles(ctx, repositoryRoot)
	if err != nil {
		return err
	}
	if len(remainingConflictFiles) > 0 {
		return appgitflow.WrapTerminal(fmt.Errorf("merge conflicts remain unresolved: %s", strings.Join(remainingConflictFiles, ", ")))
	}
	return nil
}

func (adapter *Adapter) Commit(ctx context.Context, repositoryRoot string, message string) error {
	hasStagedChanges, err := adapter.hasStagedChanges(ctx, repositoryRoot)
	if err != nil {
		return err
	}
	if !hasStagedChanges {
		return nil
	}
	_, err = adapter.runGit(ctx, repositoryRoot, "commit", "-m", strings.TrimSpace(message))
	return err
}

func (adapter *Adapter) StageAll(ctx context.Context, repositoryRoot string) error {
	_, err := adapter.runGit(ctx, repositoryRoot, "add", "-A")
	return err
}

func (adapter *Adapter) CleanupTaskWorktree(ctx context.Context, repositoryRoot string, worktreePath string, taskBranch string) error {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))
	if _, err := adapter.runGit(ctx, repositoryRoot, "worktree", "remove", "--force", absoluteWorktreePath); err != nil {
		return err
	}
	_, _ = adapter.runGit(ctx, repositoryRoot, "branch", "-D", taskBranch)
	return nil
}

func (adapter *Adapter) CleanupRunArtifacts(ctx context.Context, repositoryRoot string, runPrefix string) error {
	cleanRunPrefix := strings.TrimSpace(runPrefix)
	if cleanRunPrefix == "" {
		return appgitflow.WrapTerminal(fmt.Errorf("run_prefix is required"))
	}

	worktreeListOutput, err := adapter.runGit(ctx, repositoryRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return err
	}

	branchPrefix := fmt.Sprintf("task/%s/", cleanRunPrefix)
	entries := strings.Split(worktreeListOutput, "\n\n")
	for _, entry := range entries {
		lines := strings.Split(strings.TrimSpace(entry), "\n")
		if len(lines) == 0 {
			continue
		}
		worktreePath := ""
		branchName := ""
		for _, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if strings.HasPrefix(cleanLine, "worktree ") {
				worktreePath = strings.TrimSpace(strings.TrimPrefix(cleanLine, "worktree "))
				continue
			}
			if strings.HasPrefix(cleanLine, "branch refs/heads/") {
				branchName = strings.TrimSpace(strings.TrimPrefix(cleanLine, "branch refs/heads/"))
			}
		}
		if worktreePath == "" || branchName == "" {
			continue
		}
		if !strings.HasPrefix(branchName, branchPrefix) {
			continue
		}
		if filepath.Clean(worktreePath) == filepath.Clean(repositoryRoot) {
			continue
		}
		_, _ = adapter.runGit(ctx, repositoryRoot, "worktree", "remove", "--force", worktreePath)
	}

	branchesOutput, branchErr := adapter.runGit(ctx, repositoryRoot, "branch", "--list", branchPrefix+"*", "--format=%(refname:short)")
	if branchErr != nil {
		return branchErr
	}
	for _, line := range strings.Split(branchesOutput, "\n") {
		branchName := strings.TrimSpace(line)
		if branchName == "" {
			continue
		}
		_, _ = adapter.runGit(ctx, repositoryRoot, "branch", "-D", branchName)
	}

	return nil
}

func (adapter *Adapter) CurrentBranch(ctx context.Context, repositoryRoot string) (string, error) {
	output, err := adapter.runGit(ctx, repositoryRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (adapter *Adapter) conflictFiles(ctx context.Context, repositoryRoot string) ([]string, error) {
	output, err := adapter.runGit(ctx, repositoryRoot, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			files = append(files, cleanLine)
		}
	}
	return files, nil
}

func (adapter *Adapter) hasStagedChanges(ctx context.Context, repositoryRoot string) (bool, error) {
	output, err := adapter.runGit(ctx, repositoryRoot, "diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

func (adapter *Adapter) runGit(ctx context.Context, repositoryRoot string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", repositoryRoot}, args...)...)
	output, err := command.CombinedOutput()
	outputText := string(output)
	if adapter.logger != nil {
		adapter.logger.WithFields(logrus.Fields{
			"event":           "git.command",
			"repository_root": repositoryRoot,
			"args":            args,
			"output":          strings.TrimSpace(outputText),
		}).Debug("executed git command")
	}
	if err != nil {
		if ctx.Err() != nil {
			return "", appgitflow.WrapTransient(fmt.Errorf("git %s: %w", strings.Join(args, " "), ctx.Err()))
		}
		return "", appgitflow.WrapTerminal(fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(outputText)))
	}
	return outputText, nil
}

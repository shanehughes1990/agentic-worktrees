package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	appgitflow "github.com/shanehughes1990/agentic-worktrees/internal/application/gitflow"
	"github.com/sirupsen/logrus"
)

type Adapter struct {
	logger *logrus.Logger
	mergeMu sync.Mutex
}

func NewAdapter(logger *logrus.Logger) *Adapter {
	return &Adapter{logger: logger}
}

func (adapter *Adapter) CreateTaskWorktree(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) error {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))
	if err := os.MkdirAll(filepath.Dir(absoluteWorktreePath), 0o755); err != nil {
		return appgitflow.WrapTerminal(fmt.Errorf("create worktree parent directory: %w", err))
	}

	hasWorktreeMetadata := false
	if statInfo, statErr := os.Stat(absoluteWorktreePath); statErr == nil && statInfo.IsDir() {
		worktreeGitPath := filepath.Join(absoluteWorktreePath, ".git")
		if _, gitStatErr := os.Stat(worktreeGitPath); gitStatErr == nil {
			hasWorktreeMetadata = true
		} else {
			if removeErr := os.RemoveAll(absoluteWorktreePath); removeErr != nil {
				return appgitflow.WrapTerminal(fmt.Errorf("remove invalid worktree path: %w", removeErr))
			}
		}
	}

	if hasWorktreeMetadata {
		return nil
	}

	_, _ = adapter.runGit(ctx, repositoryRoot, "worktree", "prune")

	if _, err := adapter.runGit(ctx, repositoryRoot, "worktree", "add", absoluteWorktreePath, taskBranch); err == nil {
		return nil
	}

	if _, err := adapter.runGit(ctx, repositoryRoot, "worktree", "add", "-B", taskBranch, absoluteWorktreePath, sourceBranch); err != nil {
		return err
	}
	return nil
}

func (adapter *Adapter) MergeTaskBranch(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string) (appgitflow.MergeAttempt, error) {
	if err := adapter.recoverMergeStateIfNeeded(ctx, repositoryRoot); err != nil {
		return appgitflow.MergeAttempt{}, err
	}
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

func (adapter *Adapter) SyncTaskBranchWithSource(ctx context.Context, repositoryRoot string, sourceBranch string, taskBranch string, worktreePath string) (appgitflow.MergeAttempt, error) {
	absoluteWorktreePath := filepath.Join(repositoryRoot, filepath.FromSlash(worktreePath))

	if err := adapter.recoverMergeStateIfNeeded(ctx, repositoryRoot); err != nil {
		return appgitflow.MergeAttempt{}, err
	}
	if err := adapter.recoverMergeStateIfNeeded(ctx, absoluteWorktreePath); err != nil {
		return appgitflow.MergeAttempt{}, err
	}

	if _, err := adapter.runGit(ctx, repositoryRoot, "checkout", sourceBranch); err != nil {
		return appgitflow.MergeAttempt{}, err
	}
	if _, err := adapter.runGit(ctx, absoluteWorktreePath, "checkout", taskBranch); err != nil {
		return appgitflow.MergeAttempt{}, err
	}

	_, err := adapter.runGit(ctx, absoluteWorktreePath, "merge", "--no-ff", "--no-commit", sourceBranch)
	if err != nil {
		conflictFiles, conflictErr := adapter.conflictFiles(ctx, absoluteWorktreePath)
		if conflictErr != nil {
			return appgitflow.MergeAttempt{}, conflictErr
		}
		if len(conflictFiles) > 0 {
			return appgitflow.MergeAttempt{ConflictFiles: conflictFiles}, nil
		}
		return appgitflow.MergeAttempt{}, err
	}

	hasStagedChanges, stagedErr := adapter.hasStagedChanges(ctx, absoluteWorktreePath)
	if stagedErr != nil {
		return appgitflow.MergeAttempt{}, stagedErr
	}
	if hasStagedChanges {
		if _, commitErr := adapter.runGit(ctx, absoluteWorktreePath, "commit", "-m", fmt.Sprintf("Sync %s with latest %s before final merge", taskBranch, sourceBranch)); commitErr != nil {
			return appgitflow.MergeAttempt{}, commitErr
		}
		return appgitflow.MergeAttempt{}, nil
	}

	return appgitflow.MergeAttempt{NoChanges: true}, nil
}

func (adapter *Adapter) ValidateWorktree(ctx context.Context, repositoryRoot string) error {
	command := exec.CommandContext(ctx, "go", "test", "./...")
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	outputText := strings.TrimSpace(string(output))
	if adapter.logger != nil {
		adapter.logger.WithFields(logrus.Fields{
			"event": "git.validation",
			"repository_root": repositoryRoot,
			"output": outputText,
		}).Debug("validated worktree with go test ./...")
	}
	if err != nil {
		if ctx.Err() != nil {
			return appgitflow.WrapTransient(fmt.Errorf("validate worktree tests: %w", ctx.Err()))
		}
		return appgitflow.WrapTerminal(fmt.Errorf("validate worktree tests failed: %s", outputText))
	}
	return nil
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
	if adapter.shouldSerializeMutation(repositoryRoot, args) {
		adapter.mergeMu.Lock()
		defer adapter.mergeMu.Unlock()
	}

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
		cleanOutput := strings.TrimSpace(outputText)
		if isMissingWorktreePathError(cleanOutput) || isRetryableIndexConflictError(cleanOutput) {
			return "", appgitflow.WrapTransient(fmt.Errorf("git %s failed: %s", strings.Join(args, " "), cleanOutput))
		}
		return "", appgitflow.WrapTerminal(fmt.Errorf("git %s failed: %s", strings.Join(args, " "), cleanOutput))
	}
	return outputText, nil
}

func isMissingWorktreePathError(outputText string) bool {
	cleanOutput := strings.ToLower(strings.TrimSpace(outputText))
	if cleanOutput == "" {
		return false
	}
	if !strings.Contains(cleanOutput, "no such file or directory") {
		return false
	}
	if !strings.Contains(cleanOutput, "cannot change to") {
		return false
	}
	return strings.Contains(cleanOutput, "/.worktree/worktrees/")
}

func isRetryableIndexConflictError(outputText string) bool {
	cleanOutput := strings.ToLower(strings.TrimSpace(outputText))
	if cleanOutput == "" {
		return false
	}
	if strings.Contains(cleanOutput, "you need to resolve your current index first") {
		return true
	}
	return strings.Contains(cleanOutput, "needs merge")
}

func (adapter *Adapter) recoverMergeStateIfNeeded(ctx context.Context, repositoryRoot string) error {
	hasUnmerged, unmergedFiles, err := adapter.hasUnmergedFiles(ctx, repositoryRoot)
	if err != nil {
		return err
	}
	if !hasUnmerged {
		return nil
	}

	if _, abortErr := adapter.runGit(ctx, repositoryRoot, "merge", "--abort"); abortErr != nil {
		abortMessage := strings.ToLower(strings.TrimSpace(abortErr.Error()))
		if !strings.Contains(abortMessage, "there is no merge to abort") && !strings.Contains(abortMessage, "no merge to abort") {
			return appgitflow.WrapTransient(fmt.Errorf("recover merge state with git merge --abort: %w", abortErr))
		}
	}

	hasUnmerged, unmergedFiles, err = adapter.hasUnmergedFiles(ctx, repositoryRoot)
	if err != nil {
		return err
	}
	if !hasUnmerged {
		return nil
	}

	if _, resetErr := adapter.runGit(ctx, repositoryRoot, "reset", "--merge"); resetErr != nil {
		return appgitflow.WrapTransient(fmt.Errorf("recover merge state with git reset --merge: %w", resetErr))
	}

	hasUnmerged, unmergedFiles, err = adapter.hasUnmergedFiles(ctx, repositoryRoot)
	if err != nil {
		return err
	}
	if hasUnmerged {
		return appgitflow.WrapTransient(fmt.Errorf("unmerged index persists after merge-state recovery: %s", strings.Join(unmergedFiles, ", ")))
	}
	return nil
}

func (adapter *Adapter) hasUnmergedFiles(ctx context.Context, repositoryRoot string) (bool, []string, error) {
	output, err := adapter.runGit(ctx, repositoryRoot, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return false, nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			files = append(files, cleanLine)
		}
	}
	return len(files) > 0, files, nil
}

func (adapter *Adapter) shouldSerializeMutation(repositoryRoot string, args []string) bool {
	if len(args) == 0 {
		return false
	}

	cleanRepositoryRoot := filepath.Clean(strings.TrimSpace(repositoryRoot))
	worktreeToken := string(filepath.Separator) + ".worktree" + string(filepath.Separator)
	if strings.Contains(cleanRepositoryRoot, worktreeToken) {
		return false
	}

	switch strings.TrimSpace(args[0]) {
	case "checkout", "merge", "commit", "add", "reset", "restore", "cherry-pick", "rebase", "amend":
		return true
	default:
		return false
	}
}

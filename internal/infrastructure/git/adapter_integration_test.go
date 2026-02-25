package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAdapterMergeConflictFlow(t *testing.T) {
	tempDirectory := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runGitCommand(t, ctx, tempDirectory, "init")
	runGitCommand(t, ctx, tempDirectory, "config", "user.email", "test@example.com")
	runGitCommand(t, ctx, tempDirectory, "config", "user.name", "tester")

	filePath := filepath.Join(tempDirectory, "main.txt")
	if err := os.WriteFile(filePath, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCommand(t, ctx, tempDirectory, "add", "main.txt")
	runGitCommand(t, ctx, tempDirectory, "commit", "-m", "seed")
	defaultBranch := strings.TrimSpace(runGitCommandOutput(t, ctx, tempDirectory, "branch", "--show-current"))
	if defaultBranch == "" {
		defaultBranch = "master"
	}

	adapter := NewAdapter(nil)
	if err := adapter.CreateTaskWorktree(ctx, tempDirectory, defaultBranch, "task/run-1/task-1", ".worktree/run-1-task-1"); err != nil {
		t.Fatalf("create task worktree: %v", err)
	}

	worktreeFilePath := filepath.Join(tempDirectory, ".worktree", "run-1-task-1", "main.txt")
	if err := os.WriteFile(worktreeFilePath, []byte("line1\ntask-branch\n"), 0o644); err != nil {
		t.Fatalf("write worktree file: %v", err)
	}
	runGitCommand(t, ctx, filepath.Join(tempDirectory, ".worktree", "run-1-task-1"), "add", "main.txt")
	runGitCommand(t, ctx, filepath.Join(tempDirectory, ".worktree", "run-1-task-1"), "commit", "-m", "task change")

	runGitCommand(t, ctx, tempDirectory, "checkout", defaultBranch)
	if err := os.WriteFile(filePath, []byte("line1\nsource-branch\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	runGitCommand(t, ctx, tempDirectory, "add", "main.txt")
	runGitCommand(t, ctx, tempDirectory, "commit", "-m", "source change")

	mergeAttempt, err := adapter.MergeTaskBranch(ctx, tempDirectory, defaultBranch, "task/run-1/task-1")
	if err != nil {
		t.Fatalf("merge task branch: %v", err)
	}
	if len(mergeAttempt.ConflictFiles) == 0 {
		t.Fatalf("expected conflict files")
	}

	if err := os.WriteFile(filePath, []byte("line1\nresolved-by-agent\n"), 0o644); err != nil {
		t.Fatalf("write resolved file: %v", err)
	}
	runGitCommand(t, ctx, tempDirectory, "add", "main.txt")

	if err := adapter.ResolveConflicts(ctx, tempDirectory, mergeAttempt.ConflictFiles, ""); err != nil {
		t.Fatalf("resolve conflicts: %v", err)
	}
	if err := adapter.Commit(ctx, tempDirectory, "resolve conflicts"); err != nil {
		t.Fatalf("commit resolution: %v", err)
	}
	if err := adapter.CleanupTaskWorktree(ctx, tempDirectory, ".worktree/run-1-task-1", "task/run-1/task-1"); err != nil {
		t.Fatalf("cleanup worktree: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}
	if strings.Contains(string(content), "<<<<<<<") {
		t.Fatalf("expected conflict markers to be removed")
	}
}

func TestAdapterCreateTaskWorktreeRecreatesFromExistingBranchWhenFilesystemMissing(t *testing.T) {
	tempDirectory := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runGitCommand(t, ctx, tempDirectory, "init")
	runGitCommand(t, ctx, tempDirectory, "config", "user.email", "test@example.com")
	runGitCommand(t, ctx, tempDirectory, "config", "user.name", "tester")

	seedFilePath := filepath.Join(tempDirectory, "main.txt")
	if err := os.WriteFile(seedFilePath, []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitCommand(t, ctx, tempDirectory, "add", "main.txt")
	runGitCommand(t, ctx, tempDirectory, "commit", "-m", "seed")
	defaultBranch := strings.TrimSpace(runGitCommandOutput(t, ctx, tempDirectory, "branch", "--show-current"))
	if defaultBranch == "" {
		defaultBranch = "master"
	}

	adapter := NewAdapter(nil)
	worktreeRelPath := ".worktree/run-1-task-recreate"
	taskBranch := "task/run-1/task-recreate"
	if err := adapter.CreateTaskWorktree(ctx, tempDirectory, defaultBranch, taskBranch, worktreeRelPath); err != nil {
		t.Fatalf("initial create task worktree: %v", err)
	}

	worktreeAbsPath := filepath.Join(tempDirectory, filepath.FromSlash(worktreeRelPath))
	resumeFilePath := filepath.Join(worktreeAbsPath, "resume.txt")
	if err := os.WriteFile(resumeFilePath, []byte("resume branch content\n"), 0o644); err != nil {
		t.Fatalf("write resume file: %v", err)
	}
	runGitCommand(t, ctx, worktreeAbsPath, "add", "resume.txt")
	runGitCommand(t, ctx, worktreeAbsPath, "commit", "-m", "save progress")

	if err := adapter.CleanupTaskWorktree(ctx, tempDirectory, worktreeRelPath, taskBranch); err != nil {
		t.Fatalf("cleanup worktree before recreate: %v", err)
	}

	if err := adapter.CreateTaskWorktree(ctx, tempDirectory, defaultBranch, taskBranch, worktreeRelPath); err != nil {
		t.Fatalf("recreate task worktree from existing branch: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(worktreeAbsPath, "resume.txt")); statErr != nil {
		t.Fatalf("expected recreated worktree to contain existing branch changes: %v", statErr)
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

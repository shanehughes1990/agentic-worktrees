package scm

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GitRunner interface {
	Run(ctx context.Context, directory string, arguments ...string) (string, error)
}

type ExecGitRunner struct{}

func NewExecGitRunner() *ExecGitRunner {
	return &ExecGitRunner{}
}

func (runner *ExecGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", arguments...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	trimmedOutput := strings.TrimSpace(string(output))
	if err != nil {
		if trimmedOutput == "" {
			return "", fmt.Errorf("git %s failed: %w", strings.Join(arguments, " "), err)
		}
		return "", fmt.Errorf("git %s failed: %s: %w", strings.Join(arguments, " "), trimmedOutput, err)
	}
	return trimmedOutput, nil
}

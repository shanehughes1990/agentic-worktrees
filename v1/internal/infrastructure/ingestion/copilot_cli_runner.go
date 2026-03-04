package ingestion

import (
	"agentic-orchestrator/internal/domain/failures"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CopilotCLIRunner struct {
	binaryPath string
}

const defaultCopilotModel = "gpt-5.3-codex"

func NewCopilotCLIRunner(binaryPath string) (*CopilotCLIRunner, error) {
	path := strings.TrimSpace(binaryPath)
	if path == "" {
		path = "copilot"
	}
	return &CopilotCLIRunner{binaryPath: path}, nil
}

func (runner *CopilotCLIRunner) GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string) error {
	if runner == nil {
		return failures.WrapTerminal(fmt.Errorf("copilot cli runner is not initialized"))
	}
	cleanSandboxDir := strings.TrimSpace(sandboxDir)
	cleanPrompt := strings.TrimSpace(prompt)
	cleanOutputPath := strings.TrimSpace(outputPath)
	if cleanSandboxDir == "" {
		return failures.WrapTerminal(fmt.Errorf("sandbox_dir is required"))
	}
	if cleanPrompt == "" {
		return failures.WrapTerminal(fmt.Errorf("prompt is required"))
	}
	if cleanOutputPath == "" {
		return failures.WrapTerminal(fmt.Errorf("output_path is required"))
	}
	cleanModel := strings.TrimSpace(model)
	if cleanModel == "" {
		cleanModel = defaultCopilotModel
	}

	if err := os.MkdirAll(filepath.Dir(cleanOutputPath), 0o755); err != nil {
		return failures.WrapTransient(fmt.Errorf("ensure output directory: %w", err))
	}

	command := exec.CommandContext(ctx, runner.binaryPath,
		"-p", cleanPrompt,
		"--model", cleanModel,
		"--allow-all",
		"--add-dir", cleanSandboxDir,
	)
	command.Dir = cleanSandboxDir

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer

	if err := command.Run(); err != nil {
		return failures.WrapTransient(fmt.Errorf("run copilot cli ingestion prompt: %w (stdout=%s stderr=%s)", err, strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String())))
	}
	if _, err := os.Stat(cleanOutputPath); err != nil {
		return failures.WrapTransient(fmt.Errorf("copilot cli did not generate taskboard output at %s: %w", cleanOutputPath, err))
	}
	return nil
}

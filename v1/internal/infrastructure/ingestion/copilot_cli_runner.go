package ingestion

import (
	applicationingestion "agentic-orchestrator/internal/application/ingestion"
	"agentic-orchestrator/internal/domain/failures"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

func (runner *CopilotCLIRunner) GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string, runContext applicationingestion.AgentRunContext) (applicationingestion.AgentRunContext, error) {
	if runner == nil {
		return applicationingestion.AgentRunContext{}, failures.WrapTerminal(fmt.Errorf("copilot cli runner is not initialized"))
	}
	cleanSandboxDir := strings.TrimSpace(sandboxDir)
	cleanPrompt := strings.TrimSpace(prompt)
	cleanOutputPath := strings.TrimSpace(outputPath)
	resolvedRunContext := runContext
	if strings.TrimSpace(resolvedRunContext.StreamID) == "" {
		resolvedRunContext.StreamID = fmt.Sprintf("ingestion-stream:%d", time.Now().UTC().UnixNano())
	}
	if strings.TrimSpace(resolvedRunContext.SessionID) == "" {
		resolvedRunContext.SessionID = fmt.Sprintf("ingestion-session:%s", strings.TrimPrefix(strings.TrimSpace(resolvedRunContext.StreamID), "ingestion-stream:"))
	}
	if cleanSandboxDir == "" {
		return applicationingestion.AgentRunContext{}, failures.WrapTerminal(fmt.Errorf("sandbox_dir is required"))
	}
	if cleanPrompt == "" {
		return applicationingestion.AgentRunContext{}, failures.WrapTerminal(fmt.Errorf("prompt is required"))
	}
	if cleanOutputPath == "" {
		return applicationingestion.AgentRunContext{}, failures.WrapTerminal(fmt.Errorf("output_path is required"))
	}
	cleanModel := strings.TrimSpace(model)
	if cleanModel == "" {
		cleanModel = defaultCopilotModel
	}

	if err := os.MkdirAll(filepath.Dir(cleanOutputPath), 0o755); err != nil {
		return applicationingestion.AgentRunContext{}, failures.WrapTransient(fmt.Errorf("ensure output directory: %w", err))
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
		return applicationingestion.AgentRunContext{}, failures.WrapTransient(fmt.Errorf("run copilot cli ingestion prompt: %w (stdout=%s stderr=%s)", err, strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String())))
	}
	if _, err := os.Stat(cleanOutputPath); err != nil {
		stdoutPayload := strings.TrimSpace(stdoutBuffer.String())
		if stdoutPayload == "" {
			return applicationingestion.AgentRunContext{}, failures.WrapTransient(fmt.Errorf("copilot cli did not generate taskboard output at %s and returned empty stdout: %w", cleanOutputPath, err))
		}
		candidate := extractJSONPayload(stdoutPayload)
		if !json.Valid([]byte(candidate)) {
			return applicationingestion.AgentRunContext{}, failures.WrapTransient(fmt.Errorf("copilot cli output is not valid JSON for taskboard ingestion (stdout=%s)", stdoutPayload))
		}
		if writeErr := os.WriteFile(cleanOutputPath, []byte(candidate), 0o644); writeErr != nil {
			return applicationingestion.AgentRunContext{}, failures.WrapTransient(fmt.Errorf("persist taskboard json from copilot stdout: %w", writeErr))
		}
	}
	return resolvedRunContext, nil
}

func extractJSONPayload(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "json"))
		trimmed = strings.TrimSpace(strings.TrimSuffix(trimmed, "```"))
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(trimmed[start : end+1])
	}
	return trimmed
}

package agent

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"agentic-orchestrator/internal/domain/failures"
)

type CopilotPromptRefiner struct {
	binaryPath string
	model      string
}

func NewCopilotPromptRefiner(binaryPath string, model string) (*CopilotPromptRefiner, error) {
	path := strings.TrimSpace(binaryPath)
	if path == "" {
		path = "copilot"
	}
	resolvedModel := strings.TrimSpace(model)
	if resolvedModel == "" {
		resolvedModel = "gpt-5.3-codex"
	}
	return &CopilotPromptRefiner{binaryPath: path, model: resolvedModel}, nil
}

func (refiner *CopilotPromptRefiner) RefinePrompt(ctx context.Context, input applicationcontrolplane.PromptRefinerInput) (string, error) {
	if refiner == nil {
		return "", failures.WrapTerminal(fmt.Errorf("copilot prompt refiner is not initialized"))
	}
	taskboardName := strings.TrimSpace(input.TaskboardName)
	if taskboardName == "" {
		return "", failures.WrapTerminal(fmt.Errorf("taskboard_name is required"))
	}
	userPrompt := strings.TrimSpace(input.UserPrompt)

	refinePrompt := buildRefinementPrompt(taskboardName, userPrompt)
	command := exec.CommandContext(ctx, refiner.binaryPath,
		"-p", refinePrompt,
		"--model", refiner.model,
		"--allow-all",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", failures.WrapTransient(fmt.Errorf("run copilot prompt refinement: %w (output=%s)", err, strings.TrimSpace(string(output))))
	}
	refined := strings.TrimSpace(stripCodeFence(strings.TrimSpace(string(output))))
	refined = strings.TrimSpace(stripCopilotUsageFooter(refined))
	if refined == "" {
		return "", failures.WrapTransient(fmt.Errorf("copilot prompt refinement returned empty output"))
	}
	return refined, nil
}

func buildRefinementPrompt(taskboardName string, userPrompt string) string {
	if strings.TrimSpace(userPrompt) == "" {
		return fmt.Sprintf("Rewrite and return a single concise instruction for a taskboard ingestion agent. Taskboard name: %q. Requirements: clear scope, prioritized epics/tasks, actionable wording, minimal ambiguity. Return plain text only.", taskboardName)
	}
	return fmt.Sprintf("Rewrite the following user intent into one clear and concise instruction for a taskboard ingestion agent. Taskboard name: %q. Original user text: %q. Requirements: retain intent, remove fluff, make it execution-ready, include scope and priorities if implied. Return plain text only.", taskboardName, userPrompt)
}

func stripCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "text"))
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "markdown"))
		trimmed = strings.TrimSpace(strings.TrimSuffix(trimmed, "```"))
	}
	return strings.TrimSpace(trimmed)
}

func stripCopilotUsageFooter(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	markers := []string{
		"Total usage est:",
		"API time spent:",
		"Total session time:",
		"Total code changes:",
		"Breakdown by AI model:",
	}
	cutAt := len(trimmed)
	for _, marker := range markers {
		if index := strings.Index(trimmed, marker); index >= 0 && index < cutAt {
			cutAt = index
		}
	}
	if cutAt < len(trimmed) {
		trimmed = strings.TrimSpace(trimmed[:cutAt])
	}
	return strings.TrimSpace(trimmed)
}

var _ applicationcontrolplane.PromptRefiner = (*CopilotPromptRefiner)(nil)

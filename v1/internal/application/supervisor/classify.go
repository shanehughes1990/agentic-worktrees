package supervisor

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/domain/failures"
	"strings"
)

func classifyExecutionFailure(record taskengine.ExecutionRecord) failures.Class {
	status := strings.TrimSpace(record.ErrorMessage)
	if status == "" {
		return failures.ClassUnknown
	}
	lower := strings.ToLower(status)
	if strings.Contains(lower, "invalid") || strings.Contains(lower, "required") || strings.Contains(lower, "not found") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "forbidden") {
		return failures.ClassTerminal
	}
	return failures.ClassTransient
}

package resolvers

import (
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"testing"
	"time"
)

func TestToGraphTaskboardMapsIngestionAudits(t *testing.T) {
	now := time.Now().UTC()
	inputTokens := 123
	outputTokens := 456
	board := domaintracker.Board{
		BoardID:   "board-1",
		ProjectID: "project-1",
		Name:      "Board One",
		State:     domaintracker.BoardStatePending,
		Epics: []domaintracker.Epic{{
			ID:      domaintracker.WorkItemID("epic-1"),
			BoardID: "board-1",
			Title:   "Epic One",
			State:   domaintracker.EpicStatePlanned,
			Tasks: []domaintracker.Task{{
				ID:       domaintracker.WorkItemID("task-1"),
				BoardID:  "board-1",
				EpicID:   domaintracker.WorkItemID("epic-1"),
				Title:    "Task One",
				TaskType: "analysis",
				State:    domaintracker.TaskStatePlanned,
				Audits: []domaintracker.TaskModelAudit{{
					ModelProvider: "copilot",
					ModelName:     "gpt-5.3-codex",
					ModelRunID:    "task-run-1",
				}},
			}},
		}},
		IngestionAudits: []domaintracker.TaskModelAudit{{
			ModelProvider:     "copilot",
			ModelName:         "gpt-5.3-codex",
			ModelVersion:      "2026-03-04",
			ModelRunID:        "run-1",
			AgentSessionID:    "session-1",
			AgentStreamID:     "stream-1",
			PromptFingerprint: "fp-1",
			InputTokens:       &inputTokens,
			OutputTokens:      &outputTokens,
			StartedAt:         &now,
			CompletedAt:       &now,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mapped := toGraphTaskboard(board)
	if mapped == nil {
		t.Fatalf("expected mapped taskboard")
	}
	if len(mapped.IngestionAudits) != 1 {
		t.Fatalf("expected one ingestion audit, got %d", len(mapped.IngestionAudits))
	}
	audit := mapped.IngestionAudits[0]
	if audit == nil {
		t.Fatalf("expected ingestion audit to be non-nil")
	}
	if audit.ModelProvider != "copilot" || audit.ModelName != "gpt-5.3-codex" {
		t.Fatalf("unexpected audit identity: %+v", audit)
	}
	if audit.InputTokens == nil || *audit.InputTokens != int32(inputTokens) {
		t.Fatalf("unexpected input tokens: %+v", audit.InputTokens)
	}
	if audit.OutputTokens == nil || *audit.OutputTokens != int32(outputTokens) {
		t.Fatalf("unexpected output tokens: %+v", audit.OutputTokens)
	}
	if len(mapped.Epics) != 1 || len(mapped.Epics[0].Tasks) != 1 {
		t.Fatalf("expected one task in one epic, got epics=%d tasks=%d", len(mapped.Epics), len(mapped.Epics[0].Tasks))
	}
	taskAudit := mapped.Epics[0].Tasks[0].Audits
	if len(taskAudit) != 1 {
		t.Fatalf("expected one task audit, got %d", len(taskAudit))
	}
	if taskAudit[0] == nil || taskAudit[0].ModelRunID == nil || *taskAudit[0].ModelRunID != "task-run-1" {
		t.Fatalf("unexpected task audit mapping: %+v", taskAudit)
	}
}

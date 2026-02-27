package taskboard

import (
	"testing"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func TestEvaluateBoardQualityPassesHighQualityBoard(t *testing.T) {
	board := buildQualityTestBoard([]domaintaskboard.Task{
		{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Define domain source contract", Status: domaintaskboard.StatusNotStarted}},
		{WorkItem: domaintaskboard.WorkItem{ID: "task-2", BoardID: "board-1", Title: "Implement filesystem source adapter", Status: domaintaskboard.StatusNotStarted}, DependsOn: []string{"task-1"}},
		{WorkItem: domaintaskboard.WorkItem{ID: "task-3", BoardID: "board-1", Title: "Add adapter parity tests", Status: domaintaskboard.StatusNotStarted}, DependsOn: []string{"task-2"}},
	})

	report := EvaluateBoardQuality(board, 85)
	if !report.Passed {
		t.Fatalf("expected board to pass quality gate, got report: %#v", report)
	}
	if report.Score < 85 {
		t.Fatalf("expected score >= 85, got %d", report.Score)
	}
}

func TestEvaluateBoardQualityFailsDuplicateAndConflict(t *testing.T) {
	board := buildQualityTestBoard([]domaintaskboard.Task{
		{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Wire filesystem adapter as default provider", Status: domaintaskboard.StatusNotStarted}},
		{WorkItem: domaintaskboard.WorkItem{ID: "task-2", BoardID: "board-1", Title: "Wire filesystem adapter as default provider", Status: domaintaskboard.StatusNotStarted}},
	})

	report := EvaluateBoardQuality(board, 85)
	if report.Passed {
		t.Fatalf("expected board to fail quality gate, got report: %#v", report)
	}
	if len(report.CriticalFailures) == 0 {
		t.Fatalf("expected critical failures for duplicate/conflict, got none")
	}
}

func TestEvaluateBoardQualityFailsDependencyCycle(t *testing.T) {
	board := buildQualityTestBoard([]domaintaskboard.Task{
		{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Define source contract package", Status: domaintaskboard.StatusNotStarted}, DependsOn: []string{"task-2"}},
		{WorkItem: domaintaskboard.WorkItem{ID: "task-2", BoardID: "board-1", Title: "Implement adapter package", Status: domaintaskboard.StatusNotStarted}, DependsOn: []string{"task-1"}},
	})

	report := EvaluateBoardQuality(board, 85)
	if report.Passed {
		t.Fatalf("expected board to fail cycle quality gate, got report: %#v", report)
	}
	if len(report.CriticalFailures) == 0 {
		t.Fatalf("expected critical cycle failures, got none")
	}
}

func TestEvaluateBoardQualityDoesNotFlagSingleActionTitleWithAnd(t *testing.T) {
	board := buildQualityTestBoard([]domaintaskboard.Task{
		{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Define source identity and metadata types", Status: domaintaskboard.StatusNotStarted}},
	})

	report := EvaluateBoardQuality(board, 85)
	for _, finding := range report.Findings {
		if finding.RuleID == "title.multi_intent" {
			t.Fatalf("did not expect multi-intent finding for single-action title, got %#v", finding)
		}
	}
}

func TestEvaluateBoardQualityPenalizesBundledTestChecks(t *testing.T) {
	board := buildQualityTestBoard([]domaintaskboard.Task{
		{WorkItem: domaintaskboard.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Add ingestion tests for fake provider plus lifecycle regression", Status: domaintaskboard.StatusNotStarted}},
	})

	report := EvaluateBoardQuality(board, 85)
	found := false
	for _, finding := range report.Findings {
		if finding.RuleID == "test.bundled_checks" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected bundled test checks finding, got report: %#v", report)
	}
	if report.CategoryScores["granularity"] >= 20 {
		t.Fatalf("expected granularity score penalty for bundled test checks, got %d", report.CategoryScores["granularity"])
	}
}

func buildQualityTestBoard(tasks []domaintaskboard.Task) *domaintaskboard.Board {
	now := time.Now().UTC()
	return &domaintaskboard.Board{
		BoardID:   "board-1",
		RunID:     "run-1",
		Title:     "board",
		Status:    domaintaskboard.StatusNotStarted,
		CreatedAt: now,
		UpdatedAt: now,
		Epics: []domaintaskboard.Epic{
			{
				WorkItem: domaintaskboard.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "epic", Status: domaintaskboard.StatusNotStarted, CreatedAt: now, UpdatedAt: now},
				Tasks:    tasks,
			},
		},
	}
}

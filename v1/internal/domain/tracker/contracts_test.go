package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
	"time"
)

func TestBoardValidateAcceptsCanonicalBoard(t *testing.T) {
	board := Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Source: SourceRef{
			Kind:     SourceKindInternal,
			Location: "taskboards/board-1.json",
		},
		Status: StatusInProgress,
		Epics: []Epic{
			{
				WorkItem: WorkItem{
					ID:        "epic-1",
					BoardID:   "board-1",
					Title:     "Bootstrap",
					Status:    StatusCompleted,
					Priority:  PriorityP1,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				Tasks: []Task{
					{
						WorkItem: WorkItem{
							ID:       "task-1",
							BoardID:  "board-1",
							Title:    "Read board",
							Status:   StatusCompleted,
							Priority: PriorityP2,
						},
					},
				},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := board.Validate(); err != nil {
		t.Fatalf("expected board to validate, got %v", err)
	}
}

func TestBoardValidateRejectsMissingDependency(t *testing.T) {
	board := Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Source: SourceRef{
			Kind:     SourceKindInternal,
			Location: "taskboards/board-1.json",
		},
		Status: StatusInProgress,
		Epics: []Epic{
			{
				WorkItem: WorkItem{
					ID:      "epic-1",
					BoardID: "board-1",
					Title:   "Bootstrap",
					Status:  StatusInProgress,
				},
				Tasks: []Task{
					{
						WorkItem: WorkItem{
							ID:      "task-1",
							BoardID: "board-1",
							Title:   "Read board",
							Status:  StatusInProgress,
						},
						DependsOn: []WorkItemID{"task-missing"},
					},
				},
			},
		},
	}
	err := board.Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestWorkItemValidateRejectsUnsupportedPriority(t *testing.T) {
	item := WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: StatusInProgress, Priority: Priority("urgent")}
	err := item.Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal priority validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestSourceRefValidateRequiresLocationForGitHubIssues(t *testing.T) {
	err := (SourceRef{Kind: SourceKindGitHubIssues}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestSourceRefValidateRejectsUnsupportedSourceKind(t *testing.T) {
	err := (SourceRef{Kind: SourceKind("jira")}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

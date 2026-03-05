package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	"testing"
	"time"
)

func TestBoardValidateAcceptsCanonicalBoard(t *testing.T) {
	now := time.Now().UTC()
	board := Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Name:    "Board",
		State:   BoardStateActive,
		Epics: []Epic{{
			ID:            WorkItemID("epic-1"),
			BoardID:       "board-1",
			Title:         "Epic",
			RepositoryIDs: []string{"repo-1"},
			Deliverables:  []string{"Epic brief finalized"},
			State:         EpicStateInProgress,
			Rank:          1,
			Tasks: []Task{{
				ID:            WorkItemID("task-1"),
				BoardID:       "board-1",
				EpicID:        WorkItemID("epic-1"),
				Title:         "Task",
				RepositoryIDs: []string{"repo-1"},
				Deliverables:  []string{"README section updated"},
				TaskType:      "implementation",
				State:         TaskStateInProgress,
				Rank:          1,
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := board.Validate(); err != nil {
		t.Fatalf("expected board to validate, got %v", err)
	}
}

func TestBoardValidateRejectsMissingTaskDependency(t *testing.T) {
	board := Board{
		BoardID: "board-1",
		RunID:   "run-1",
		State:   BoardStateActive,
		Epics: []Epic{{
			ID:            WorkItemID("epic-1"),
			BoardID:       "board-1",
			Title:         "Epic",
			RepositoryIDs: []string{"repo-1"},
			Deliverables:  []string{"Epic brief finalized"},
			State:         EpicStateInProgress,
			Rank:          1,
			Tasks: []Task{{
				ID:               WorkItemID("task-1"),
				BoardID:          "board-1",
				EpicID:           WorkItemID("epic-1"),
				Title:            "Task",
				RepositoryIDs:    []string{"repo-1"},
				Deliverables:     []string{"README section updated"},
				TaskType:         "implementation",
				State:            TaskStateInProgress,
				Rank:             1,
				DependsOnTaskIDs: []WorkItemID{WorkItemID("task-missing")},
			}},
		}},
	}
	err := board.Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestTaskOutcomeValidateRequiresErrorCodeWhenFailed(t *testing.T) {
	err := (TaskOutcome{Status: OutcomeStatusFailed, Summary: "failed without code"}).Validate()
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

package taskboard

import (
	"testing"
	"time"
)

func TestBoardValidateBasics(t *testing.T) {
	board := validBoard()
	if err := board.ValidateBasics(); err != nil {
		t.Fatalf("expected valid board, got error: %v", err)
	}

	board.Epics[0].Tasks = append(board.Epics[0].Tasks, Task{WorkItem: WorkItem{ID: "t1", BoardID: board.BoardID, Title: "dup", Status: StatusNotStarted}})
	if err := board.ValidateBasics(); err == nil {
		t.Fatalf("expected duplicate id validation error")
	}
}

func TestBoardValidateComplete(t *testing.T) {
	board := validBoard()
	if err := board.ValidateComplete(); err != nil {
		t.Fatalf("expected complete board, got error: %v", err)
	}

	board.Epics[0].Tasks = nil
	if err := board.ValidateComplete(); err == nil {
		t.Fatalf("expected validation error for missing tasks")
	}
}

func TestBoardValidateCompleteRejectsMissingDependency(t *testing.T) {
	board := validBoard()
	board.Epics[0].Tasks[0].DependsOn = []string{"missing-task"}

	if err := board.ValidateBasics(); err == nil {
		t.Fatalf("expected validation error for missing dependency link")
	}
}

func TestBoardSetTaskStatus(t *testing.T) {
	board := validBoard()
	before := board.UpdatedAt

	if err := board.SetTaskStatus("t1", StatusCompleted); err != nil {
		t.Fatalf("expected status update, got error: %v", err)
	}
	if board.Epics[0].Tasks[0].Status != StatusCompleted {
		t.Fatalf("expected task status completed, got %s", board.Epics[0].Tasks[0].Status)
	}
	if board.UpdatedAt.Before(before) {
		t.Fatalf("expected board updated_at not to move backward")
	}

	if err := board.SetTaskStatus("missing", StatusCompleted); err == nil {
		t.Fatalf("expected error for missing task")
	}
}

func TestBoardSetTaskOutcomePersistsResumeSessionID(t *testing.T) {
	board := validBoard()

	err := board.SetTaskOutcome("t1", TaskOutcome{
		Status:          "canceled",
		Reason:          "runner canceled",
		ResumeSessionID: "session-123",
	})
	if err != nil {
		t.Fatalf("expected outcome update, got error: %v", err)
	}
	outcome := board.Epics[0].Tasks[0].Outcome
	if outcome == nil {
		t.Fatalf("expected task outcome to be set")
	}
	if outcome.ResumeSessionID != "session-123" {
		t.Fatalf("expected resume session id to persist, got %q", outcome.ResumeSessionID)
	}
}

func validBoard() *Board {
	now := time.Now().UTC()
	return &Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  StatusInProgress,
		Epics: []Epic{
			{
				WorkItem: WorkItem{ID: "e1", BoardID: "board-1", Title: "Epic 1", Status: StatusInProgress},
				Tasks: []Task{
					{WorkItem: WorkItem{ID: "t1", BoardID: "board-1", Title: "Task 1", Status: StatusNotStarted}},
				},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

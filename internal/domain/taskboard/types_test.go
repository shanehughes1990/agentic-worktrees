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

	board.MicroTasks = append(board.MicroTasks, MicroTask{WorkItem: WorkItem{ID: "m1", BoardID: board.BoardID, Title: "dup", Status: StatusNotStarted}, ItemType: "micro_task", TaskID: "t1"})
	if err := board.ValidateBasics(); err == nil {
		t.Fatalf("expected duplicate id validation error")
	}
}

func TestBoardSetMicroTaskStatus(t *testing.T) {
	board := validBoard()
	before := board.UpdatedAt

	if err := board.SetMicroTaskStatus("m1", StatusCompleted); err != nil {
		t.Fatalf("expected status update, got error: %v", err)
	}
	if board.MicroTasks[0].Status != StatusCompleted {
		t.Fatalf("expected micro task status completed, got %s", board.MicroTasks[0].Status)
	}
	if !board.UpdatedAt.After(before) {
		t.Fatalf("expected board updated_at to move forward")
	}

	if err := board.SetMicroTaskStatus("missing", StatusCompleted); err == nil {
		t.Fatalf("expected error for missing micro task")
	}
}

func validBoard() *Board {
	now := time.Now().UTC()
	return &Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  StatusInProgress,
		Epics: []Epic{{WorkItem: WorkItem{ID: "e1", BoardID: "board-1", Title: "Epic 1", Status: StatusCompleted}, ItemType: "epic"}},
		Tasks: []Task{{WorkItem: WorkItem{ID: "t1", BoardID: "board-1", Title: "Task 1", Status: StatusCompleted}, ItemType: "task", EpicID: "e1"}},
		MicroTasks: []MicroTask{{WorkItem: WorkItem{ID: "m1", BoardID: "board-1", Title: "Micro 1", Status: StatusNotStarted}, ItemType: "micro_task", TaskID: "t1"}},
		Dependencies: []Dependency{{EdgeID: "d1", BoardID: "board-1", FromID: "t1", ToID: "m1"}},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

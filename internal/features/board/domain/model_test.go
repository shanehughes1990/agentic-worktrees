package domain

import (
	"testing"
	"time"
)

func validBoard() Board {
	return Board{
		SchemaVersion: 1,
		SourceScope:   "docs",
		GeneratedAt:   time.Now().UTC(),
		Epics: []Epic{{
			ID:    "epic-001",
			Title: "Core",
			Tasks: []Task{{
				ID:     "task-001",
				Title:  "Task",
				Lane:   "lane-a",
				Status: TaskStatusPending,
			}},
		}},
	}
}

func TestBoardValidateSuccess(t *testing.T) {
	if err := validBoard().Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBoardValidateFailures(t *testing.T) {
	cases := []Board{
		{SchemaVersion: 0, SourceScope: "docs", Epics: validBoard().Epics},
		{SchemaVersion: 1, SourceScope: "", Epics: validBoard().Epics},
		{SchemaVersion: 1, SourceScope: "docs", Epics: []Epic{}},
		{SchemaVersion: 1, SourceScope: "docs", Epics: []Epic{{ID: "", Tasks: []Task{{ID: "task"}}}}},
		{SchemaVersion: 1, SourceScope: "docs", Epics: []Epic{{ID: "epic-1", Tasks: []Task{}}}},
	}

	for i, board := range cases {
		if err := board.Validate(); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}

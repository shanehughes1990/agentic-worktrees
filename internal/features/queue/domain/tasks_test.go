package domain

import "testing"

func TestNewPlanBoardTaskAndParse(t *testing.T) {
	task, err := NewPlanBoardTask(PlanBoardPayload{
		RunID:     "run-1",
		TaskID:    "task-1",
		ScopePath: "docs",
		OutPath:   "state/board.json",
	}, "default")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	payload, err := ParsePlanBoardPayload(task)
	if err != nil {
		t.Fatalf("parse payload: %v", err)
	}
	if payload.RunID != "run-1" || payload.TaskID != "task-1" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.IdempotencyKey == "" {
		t.Fatalf("expected generated idempotency key")
	}
}

func TestNewPlanBoardTaskValidation(t *testing.T) {
	_, err := NewPlanBoardTask(PlanBoardPayload{}, "default")
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

package tasks

import (
	"encoding/json"
	"testing"
)

func TestNewTaskboardExecuteTaskValidatesInput(t *testing.T) {
	if _, _, err := NewTaskboardExecuteTask(TaskboardExecutePayload{}); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestNewTaskboardExecuteTaskBuildsTask(t *testing.T) {
	task, options, err := NewTaskboardExecuteTask(TaskboardExecutePayload{
		BoardID:        "board-1",
		SourceBranch:   "revamp",
		RepositoryRoot: ".",
		MaxTasks:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskTypeTaskboardExecute {
		t.Fatalf("unexpected task type: %s", task.Type())
	}
	if len(options) == 0 {
		t.Fatalf("expected default queue option")
	}
	payload := TaskboardExecutePayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if payload.IdempotencyKey == "" {
		t.Fatalf("expected idempotency key fallback to be set")
	}
}

func TestNewTaskboardExecuteTaskRejectsNegativeMaxTasks(t *testing.T) {
	_, _, err := NewTaskboardExecuteTask(TaskboardExecutePayload{
		BoardID:        "board-1",
		SourceBranch:   "revamp",
		RepositoryRoot: ".",
		MaxTasks:       -1,
	})
	if err == nil {
		t.Fatalf("expected max_tasks validation error")
	}
}

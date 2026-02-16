package queue

import (
	"testing"

	"github.com/hibiken/asynq"
)

func TestNewLifecycleTaskRequiresTaskID(t *testing.T) {
	_, err := NewLifecycleTask(TypePrepareWorktree, LifecyclePayload{RunID: "run-1"}, "default")
	if err == nil {
		t.Fatalf("expected error when task id is empty")
	}
}

func TestNewLifecycleTaskAndParsePayload(t *testing.T) {
	task, err := NewLifecycleTask(TypePrepareWorktree, LifecyclePayload{RunID: "run-1", TaskID: "task-1", Prompt: "hello"}, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload, err := ParseLifecycleTaskPayload(task)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if payload.RunID != "run-1" {
		t.Fatalf("unexpected run id: %q", payload.RunID)
	}
	if payload.TaskID != "task-1" {
		t.Fatalf("unexpected task id: %q", payload.TaskID)
	}
	if payload.WorktreeName != "task-1" {
		t.Fatalf("expected default worktree name to match task id, got %q", payload.WorktreeName)
	}
	if payload.Prompt != "hello" {
		t.Fatalf("unexpected prompt: %q", payload.Prompt)
	}
	if payload.OriginBranch != "main" {
		t.Fatalf("expected default origin branch main, got %q", payload.OriginBranch)
	}
}

func TestParseLifecyclePayloadRejectsInvalidBody(t *testing.T) {
	task := asynq.NewTask(TypePrepareWorktree, []byte("not-json"))
	if _, err := ParseLifecycleTaskPayload(task); err == nil {
		t.Fatalf("expected parse error for invalid payload")
	}
}

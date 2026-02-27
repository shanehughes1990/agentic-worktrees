package asynq

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
)

func TestMapTaskStateMarksOrphanedActiveAsResumable(t *testing.T) {
	info := &asynq.TaskInfo{State: asynq.TaskStateActive, IsOrphaned: true}
	if got := mapTaskState(info); got != apptaskboard.WorkflowStatusResumable {
		t.Fatalf("expected resumable status, got %s", got)
	}
}

func TestMapTaskToWorkflowIncludesResumeSessionDetails(t *testing.T) {
	payloadBody, err := json.Marshal(tasks.GitWorktreeFlowPayload{RunID: "run-1", ResumeSessionID: "session-123"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	info := &asynq.TaskInfo{
		ID:        "task-1",
		Type:      tasks.TaskTypeGitWorktreeFlow,
		Payload:   payloadBody,
		State:     asynq.TaskStateActive,
		IsOrphaned: true,
		Retried:   1,
		MaxRetry:  25,
	}
	workflow := mapTaskToWorkflow(info)
	if workflow.Status != apptaskboard.WorkflowStatusResumable {
		t.Fatalf("expected resumable workflow, got %s", workflow.Status)
	}
	if workflow.Details["resume_session_id"] != "session-123" {
		t.Fatalf("expected resume session detail to be present, got %#v", workflow.Details["resume_session_id"])
	}
	if workflow.Details["resumable"] != true {
		t.Fatalf("expected resumable=true detail, got %#v", workflow.Details["resumable"])
	}
}

func TestMapTaskMessageIncludesOrphanResumeHint(t *testing.T) {
	payloadBody, err := json.Marshal(tasks.GitWorktreeFlowPayload{RunID: "run-1", ResumeSessionID: "session-xyz"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	message := mapTaskMessage(&asynq.TaskInfo{Type: tasks.TaskTypeGitWorktreeFlow, Payload: payloadBody, State: asynq.TaskStateActive, IsOrphaned: true})
	if message == "" {
		t.Fatalf("expected non-empty orphan resume message")
	}
}

func TestMapTaskUpdatedAtPrefersCompletedAt(t *testing.T) {
	completedAt := time.Now().UTC().Add(-time.Minute)
	lastFailedAt := time.Now().UTC()
	updated := mapTaskUpdatedAt(&asynq.TaskInfo{CompletedAt: completedAt, LastFailedAt: lastFailedAt})
	if !updated.Equal(completedAt) {
		t.Fatalf("expected completed_at to be preferred, got %s", updated)
	}
}

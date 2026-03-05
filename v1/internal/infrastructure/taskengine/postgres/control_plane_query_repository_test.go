package postgres

import (
	"agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	lifecyclepostgres "agentic-orchestrator/internal/infrastructure/lifecycle/postgres"
	"context"
	"testing"
	"time"
)

func TestControlPlaneQueryRepositoryListsPersistedReadModels(t *testing.T) {
	db := newTestDB(t)
	repository, err := NewControlPlaneQueryRepository(db)
	if err != nil {
		t.Fatalf("new control-plane query repository: %v", err)
	}
	ledger, err := NewAdmissionLedger(db)
	if err != nil {
		t.Fatalf("new admission ledger: %v", err)
	}
	if err := ledger.Upsert(context.Background(), taskengine.AdmissionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindIngestionAgent,
		IdempotencyKey: "idem-1",
		QueueTaskID:    "queue-1",
		Queue:          "ingestion",
		Status:         taskengine.AdmissionStatusQueued,
		Duplicate:      false,
		EnqueuedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("upsert admission record: %v", err)
	}
	executionJournal, err := NewPostgresExecutionJournal(db)
	if err != nil {
		t.Fatalf("new execution journal: %v", err)
	}
	if err := executionJournal.Upsert(context.Background(), taskengine.ExecutionRecord{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		JobKind:        taskengine.JobKindIngestionAgent,
		IdempotencyKey: "idem-1",
		Step:           "sync",
		Status:         taskengine.ExecutionStatusSucceeded,
		UpdatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatalf("upsert execution record: %v", err)
	}
	audit, err := NewDeadLetterAudit(db)
	if err != nil {
		t.Fatalf("new dead-letter audit: %v", err)
	}
	if err := audit.Record(context.Background(), taskengine.DeadLetterEvent{Queue: "scm", TaskID: "task-archive-1", JobKind: taskengine.JobKindSCMWorkflow, Action: taskengine.DeadLetterActionRequeue, OccurredAt: time.Now().UTC()}); err != nil {
		t.Fatalf("record dead-letter event: %v", err)
	}

	sessions, err := repository.ListSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].RunID != "run-1" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}
	jobs, err := repository.ListWorkflowJobs(context.Background(), "run-1", "task-1", 10)
	if err != nil {
		t.Fatalf("list workflow jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].QueueTaskID != "queue-1" {
		t.Fatalf("unexpected workflow jobs: %+v", jobs)
	}
	executionHistory, err := repository.ListExecutionHistory(context.Background(), controlplane.CorrelationFilter{RunID: "run-1", TaskID: "task-1", JobID: "job-1"}, 10)
	if err != nil {
		t.Fatalf("list execution history: %v", err)
	}
	if len(executionHistory) != 1 || executionHistory[0].Step != "sync" {
		t.Fatalf("unexpected execution history: %+v", executionHistory)
	}
	deadLetters, err := repository.ListDeadLetterHistory(context.Background(), "scm", 10)
	if err != nil {
		t.Fatalf("list dead-letter history: %v", err)
	}
	if len(deadLetters) != 1 || deadLetters[0].TaskID != "task-archive-1" {
		t.Fatalf("unexpected dead-letter history: %+v", deadLetters)
	}
}

func TestControlPlaneQueryRepositoryListsLifecycleReadModels(t *testing.T) {
	db := newTestDB(t)
	if _, err := lifecyclepostgres.NewEventStore(db); err != nil {
		t.Fatalf("new lifecycle event store: %v", err)
	}
	repository, err := NewControlPlaneQueryRepository(db)
	if err != nil {
		t.Fatalf("new control-plane query repository: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	endedAt := now.Add(5 * time.Minute)

	insertSession := `
INSERT INTO project_sessions (
	project_id, run_id, pipeline_type, task_id, job_id, session_id, worker_id, source_runtime,
	current_state, current_severity, last_reason_code, last_reason_summary,
	last_liveness_at, last_activity_at, last_checkpoint_at,
	last_event_seq, last_project_event_seq, started_at, ended_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if execErr := db.Exec(insertSession,
		"project-1", "run-1", "agent", "task-1", "job-1", "session-1", "worker-1", "worker",
		"healthy_active", "info", "reason-code", "reason-summary",
		now, now, now,
		int64(2), int64(3), now, endedAt, now, now,
	).Error; execErr != nil {
		t.Fatalf("insert project session: %v", execErr)
	}

	insertHistory := `
INSERT INTO project_session_history (
	event_id, schema_version, project_id, run_id, task_id, job_id, session_id, worker_id,
	source_runtime, pipeline_type, project_event_seq, event_seq, occurred_at, ingested_at,
	event_type, payload_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if execErr := db.Exec(insertHistory,
		"event-1", 1, "project-1", "run-1", "task-1", "job-1", "session-1", "worker-1",
		"worker", "agent", int64(5), int64(1), now, now, "started", `{\"ok\":true}`, now,
	).Error; execErr != nil {
		t.Fatalf("insert project session history: %v", execErr)
	}
	if execErr := db.Exec(insertHistory,
		"event-2", 1, "project-1", "run-1", "task-1", "job-1", "session-1", "worker-1",
		"worker", "agent", int64(6), int64(2), now.Add(2*time.Second), now.Add(2*time.Second), "completed", `{\"ok\":true}`, now.Add(2*time.Second),
	).Error; execErr != nil {
		t.Fatalf("insert project session history #2: %v", execErr)
	}

	snapshots, err := repository.ListLifecycleSessionSnapshots(context.Background(), "project-1", "agent", 10)
	if err != nil {
		t.Fatalf("list lifecycle session snapshots: %v", err)
	}
	if len(snapshots) != 1 || snapshots[0].SessionID != "session-1" {
		t.Fatalf("unexpected lifecycle snapshots: %+v", snapshots)
	}

	history, err := repository.ListLifecycleSessionHistory(context.Background(), "project-1", "session-1", 1, 10)
	if err != nil {
		t.Fatalf("list lifecycle session history: %v", err)
	}
	if len(history) != 1 || history[0].EventID != "event-2" {
		t.Fatalf("unexpected lifecycle history: %+v", history)
	}

	treeNodes, err := repository.ListLifecycleTreeNodes(context.Background(), controlplane.LifecycleTreeFilter{ProjectID: "project-1", PipelineType: "agent"}, 100)
	if err != nil {
		t.Fatalf("list lifecycle tree nodes: %v", err)
	}
	if len(treeNodes) != 4 {
		t.Fatalf("expected run/task/job/session nodes, got %d (%+v)", len(treeNodes), treeNodes)
	}
}

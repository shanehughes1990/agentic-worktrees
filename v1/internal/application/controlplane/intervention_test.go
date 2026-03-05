package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	"context"
	"fmt"
	"testing"
	"time"
)

type fakeLifecycleEventService struct {
	events []domainlifecycle.Event
}

func (service *fakeLifecycleEventService) AppendEvent(_ context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	event.EventSeq = int64(len(service.events) + 1)
	event.ProjectEventSeq = int64(len(service.events) + 100)
	service.events = append(service.events, event)
	return event, nil
}

type interventionQueryRepoStub struct{}

func (repository *interventionQueryRepoStub) ListSessions(ctx context.Context, limit int) ([]SessionSummary, error) {
	_ = ctx
	_ = limit
	return nil, nil
}

func (repository *interventionQueryRepoStub) GetSession(ctx context.Context, runID string) (*SessionSummary, error) {
	_ = ctx
	_ = runID
	return nil, nil
}

func (repository *interventionQueryRepoStub) ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]WorkflowJob, error) {
	_ = ctx
	_ = limit
	return []WorkflowJob{{
		RunID:       runID,
		TaskID:      taskID,
		JobID:       "job-1",
		ProjectID:   "project-1",
		QueueTaskID: "queue-task-1",
		Queue:       "agent",
		Status:      "archived",
	}}, nil
}

func (repository *interventionQueryRepoStub) ListExecutionHistory(ctx context.Context, filter CorrelationFilter, limit int) ([]ExecutionHistoryRecord, error) {
	_ = ctx
	_ = filter
	_ = limit
	return nil, nil
}

func (repository *interventionQueryRepoStub) ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]DeadLetterHistoryRecord, error) {
	_ = ctx
	_ = queue
	_ = limit
	return nil, nil
}

func (repository *interventionQueryRepoStub) ListLifecycleSessionSnapshots(ctx context.Context, projectID string, pipelineType string, limit int) ([]LifecycleSessionSnapshot, error) {
	_ = ctx
	_ = pipelineType
	_ = limit
	return []LifecycleSessionSnapshot{{
		ProjectID:    projectID,
		RunID:        "run-1",
		TaskID:       "task-1",
		JobID:        "job-1",
		SessionID:    "session-1",
		PipelineType: "agent.workflow.run",
		CurrentState: "healthy_active",
		UpdatedAt:    time.Now().UTC(),
	}}, nil
}

func (repository *interventionQueryRepoStub) ListLifecycleSessionHistory(ctx context.Context, projectID string, sessionID string, fromEventSeq int64, limit int) ([]LifecycleHistoryEvent, error) {
	_ = ctx
	_ = projectID
	_ = sessionID
	_ = fromEventSeq
	_ = limit
	return nil, nil
}

func (repository *interventionQueryRepoStub) ListLifecycleTreeNodes(ctx context.Context, filter LifecycleTreeFilter, limit int) ([]LifecycleTreeNode, error) {
	_ = ctx
	_ = filter
	_ = limit
	return nil, nil
}

type interventionProjectRepoStub struct{}

type interventionDeadLetterManagerStub struct {
	requeued []struct {
		queue  string
		taskID string
	}
	requeueErr error
}

func (manager *interventionDeadLetterManagerStub) ListDeadLetters(ctx context.Context, queue string, limit int) ([]taskengine.DeadLetterTask, error) {
	_ = ctx
	_ = queue
	_ = limit
	return nil, nil
}

func (manager *interventionDeadLetterManagerStub) RequeueDeadLetter(ctx context.Context, queue string, taskID string) error {
	_ = ctx
	if manager.requeueErr != nil {
		return manager.requeueErr
	}
	manager.requeued = append(manager.requeued, struct {
		queue  string
		taskID string
	}{queue: queue, taskID: taskID})
	return nil
}

func (manager *interventionDeadLetterManagerStub) DeleteProjectTasks(ctx context.Context, projectID string) error {
	_ = ctx
	_ = projectID
	return nil
}

func (repository *interventionProjectRepoStub) ListProjectSetups(ctx context.Context, limit int) ([]ProjectSetup, error) {
	_ = ctx
	_ = limit
	return nil, nil
}

func (repository *interventionProjectRepoStub) GetProjectSetup(ctx context.Context, projectID string) (*ProjectSetup, error) {
	_ = ctx
	_ = projectID
	return nil, nil
}

func (repository *interventionProjectRepoStub) UpsertProjectSetup(ctx context.Context, setup ProjectSetup) (*ProjectSetup, error) {
	_ = ctx
	return &setup, nil
}

func (repository *interventionProjectRepoStub) DeleteProjectSetup(ctx context.Context, projectID string) error {
	_ = ctx
	_ = projectID
	return nil
}

func TestApplyManualInterventionPersistsLifecycleEvent(t *testing.T) {
	deadLetterManager := &interventionDeadLetterManagerStub{}
	service, err := NewService(nil, &interventionQueryRepoStub{}, &interventionProjectRepoStub{}, deadLetterManager)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	lifecycleService := &fakeLifecycleEventService{}
	service.SetLifecycleService(lifecycleService)

	result, err := service.ApplyManualIntervention(context.Background(), ManualInterventionRequest{
		ProjectID: "project-1",
		SessionID: "session-1",
		Action:    ManualInterventionActionRetry,
		Reason:    "Retry after transient network failure",
		ActorID:   "operator:alice",
	})
	if err != nil {
		t.Fatalf("apply manual intervention: %v", err)
	}
	if result.EventID == "" || result.EventSeq == 0 || result.ProjectEventSeq == 0 {
		t.Fatalf("expected populated intervention result, got %+v", result)
	}
	if len(lifecycleService.events) != 1 {
		t.Fatalf("expected one lifecycle event, got %d", len(lifecycleService.events))
	}
	if lifecycleService.events[0].Payload["actor_id"] != "operator:alice" {
		t.Fatalf("expected actor_id payload to be persisted")
	}
	if len(deadLetterManager.requeued) != 1 {
		t.Fatalf("expected retry to requeue one dead-letter task, got %d", len(deadLetterManager.requeued))
	}
	if deadLetterManager.requeued[0].queue != "agent" || deadLetterManager.requeued[0].taskID != "queue-task-1" {
		t.Fatalf("unexpected requeue target: %+v", deadLetterManager.requeued[0])
	}
}

func TestApplyManualInterventionRequiresAdminForTerminate(t *testing.T) {
	service, err := NewService(nil, &interventionQueryRepoStub{}, &interventionProjectRepoStub{}, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.SetLifecycleService(&fakeLifecycleEventService{})

	_, err = service.ApplyManualIntervention(context.Background(), ManualInterventionRequest{
		ProjectID: "project-1",
		SessionID: "session-1",
		Action:    ManualInterventionActionTerminate,
		Reason:    "Terminate hung run to prevent runaway execution",
		ActorID:   "operator:alice",
		Force:     true,
	})
	if err == nil {
		t.Fatalf("expected authorization error for non-admin terminate")
	}
}

func TestApplyManualInterventionRetryReturnsErrorWhenRequeueFails(t *testing.T) {
	deadLetterManager := &interventionDeadLetterManagerStub{requeueErr: fmt.Errorf("task not found")}
	service, err := NewService(nil, &interventionQueryRepoStub{}, &interventionProjectRepoStub{}, deadLetterManager)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	service.SetLifecycleService(&fakeLifecycleEventService{})

	_, err = service.ApplyManualIntervention(context.Background(), ManualInterventionRequest{
		ProjectID: "project-1",
		SessionID: "session-1",
		Action:    ManualInterventionActionRetry,
		Reason:    "Retry after transient network failure",
		ActorID:   "operator:alice",
	})
	if err == nil {
		t.Fatalf("expected retry to fail when dead-letter requeue fails")
	}
}

package resolvers

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationstream "agentic-orchestrator/internal/application/stream"
	"agentic-orchestrator/internal/application/taskengine"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"testing"
	"time"
)

type controlPlaneFakeEngine struct{}

func (engine *controlPlaneFakeEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	return taskengine.EnqueueResult{QueueTaskID: request.IdempotencyKey, Duplicate: false}, nil
}

type controlPlaneFakeDeadLetterManager struct {
	queue     string
	taskID    string
	projectID string
}

func (manager *controlPlaneFakeDeadLetterManager) ListDeadLetters(ctx context.Context, queue string, limit int) ([]taskengine.DeadLetterTask, error) {
	_ = ctx
	_ = queue
	_ = limit
	return nil, nil
}

func (manager *controlPlaneFakeDeadLetterManager) RequeueDeadLetter(ctx context.Context, queue string, taskID string) error {
	_ = ctx
	manager.queue = queue
	manager.taskID = taskID
	return nil
}

func (manager *controlPlaneFakeDeadLetterManager) DeleteProjectTasks(ctx context.Context, projectID string) error {
	_ = ctx
	manager.projectID = projectID
	return nil
}

type controlPlaneFakeQueryRepository struct{}

func (repository *controlPlaneFakeQueryRepository) ListSessions(ctx context.Context, limit int) ([]applicationcontrolplane.SessionSummary, error) {
	_ = ctx
	_ = limit
	return []applicationcontrolplane.SessionSummary{{RunID: "run-1", TaskCount: 1, JobCount: 2, UpdatedAt: time.Unix(1700000000, 0).UTC()}}, nil
}

func (repository *controlPlaneFakeQueryRepository) GetSession(ctx context.Context, runID string) (*applicationcontrolplane.SessionSummary, error) {
	_ = ctx
	if runID != "run-1" {
		return nil, nil
	}
	result := applicationcontrolplane.SessionSummary{RunID: "run-1", TaskCount: 1, JobCount: 2, UpdatedAt: time.Unix(1700000000, 0).UTC()}
	return &result, nil
}

func (repository *controlPlaneFakeQueryRepository) ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]applicationcontrolplane.WorkflowJob, error) {
	_ = ctx
	_ = runID
	_ = taskID
	_ = limit
	return []applicationcontrolplane.WorkflowJob{{RunID: "run-1", TaskID: "task-1", JobID: "job-1", JobKind: taskengine.JobKindSCMWorkflow, IdempotencyKey: "idem-1", QueueTaskID: "q-1", Queue: "scm", Status: "queued", Duplicate: false, EnqueuedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}, nil
}

func (repository *controlPlaneFakeQueryRepository) ListExecutionHistory(ctx context.Context, filter applicationcontrolplane.CorrelationFilter, limit int) ([]applicationcontrolplane.ExecutionHistoryRecord, error) {
	_ = ctx
	_ = filter
	_ = limit
	return []applicationcontrolplane.ExecutionHistoryRecord{{RunID: "run-1", TaskID: "task-1", JobID: "job-1", JobKind: taskengine.JobKindSCMWorkflow, IdempotencyKey: "idem-1", Step: "source_state", Status: taskengine.ExecutionStatusSucceeded, UpdatedAt: time.Now().UTC()}}, nil
}

func (repository *controlPlaneFakeQueryRepository) ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]applicationcontrolplane.DeadLetterHistoryRecord, error) {
	_ = ctx
	_ = queue
	_ = limit
	return []applicationcontrolplane.DeadLetterHistoryRecord{{Queue: "scm", TaskID: "archive-1", JobKind: taskengine.JobKindSCMWorkflow, Action: taskengine.DeadLetterActionRequeue, OccurredAt: time.Now().UTC()}}, nil
}

type controlPlaneFakeProjectRepository struct{}

func baseProjectSetup() applicationcontrolplane.ProjectSetup {
	return applicationcontrolplane.ProjectSetup{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		SCMs: []applicationcontrolplane.ProjectSCM{{
			SCMID:       "scm-1",
			SCMProvider: "github",
			SCMToken:    "token",
		}},
		Repositories: []applicationcontrolplane.ProjectRepository{{
			RepositoryID:  "repo-1",
			SCMID:         "scm-1",
			RepositoryURL: "https://github.com/acme/repo",
			IsPrimary:     true,
		}},
		Boards: []applicationcontrolplane.ProjectBoard{{
			BoardID:                  "board-1",
			TrackerProvider:          "internal",
			TaskboardName:            "Acme Repo Board",
			AppliesToAllRepositories: true,
			RepositoryIDs:            []string{},
		}},
		CreatedAt: time.Unix(1700000000, 0).UTC(),
		UpdatedAt: time.Unix(1700000000, 0).UTC(),
	}
}

func (repository *controlPlaneFakeProjectRepository) ListProjectSetups(ctx context.Context, limit int) ([]applicationcontrolplane.ProjectSetup, error) {
	_ = ctx
	_ = limit
	return []applicationcontrolplane.ProjectSetup{baseProjectSetup()}, nil
}

func (repository *controlPlaneFakeProjectRepository) GetProjectSetup(ctx context.Context, projectID string) (*applicationcontrolplane.ProjectSetup, error) {
	_ = ctx
	if projectID != "project-1" {
		return nil, nil
	}
	result := baseProjectSetup()
	return &result, nil
}

func (repository *controlPlaneFakeProjectRepository) UpsertProjectSetup(ctx context.Context, setup applicationcontrolplane.ProjectSetup) (*applicationcontrolplane.ProjectSetup, error) {
	_ = ctx
	setup.CreatedAt = time.Unix(1700000000, 0).UTC()
	setup.UpdatedAt = time.Unix(1700000000, 0).UTC()
	return &setup, nil
}

func (repository *controlPlaneFakeProjectRepository) DeleteProjectSetup(ctx context.Context, projectID string) error {
	_ = ctx
	if projectID != "project-1" {
		return context.Canceled
	}
	return nil
}

type controlPlaneMemoryStreamStore struct {
	events []domainstream.Event
}

func (store *controlPlaneMemoryStreamStore) Append(ctx context.Context, event domainstream.Event) (domainstream.Event, error) {
	_ = ctx
	event.StreamOffset = uint64(len(store.events) + 1)
	store.events = append(store.events, event)
	return event, nil
}

func (store *controlPlaneMemoryStreamStore) ListFromOffset(ctx context.Context, offset uint64, limit int) ([]domainstream.Event, error) {
	_ = ctx
	if limit <= 0 {
		limit = len(store.events)
	}
	result := make([]domainstream.Event, 0)
	for _, event := range store.events {
		if event.StreamOffset <= offset {
			continue
		}
		result = append(result, event)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func newControlPlaneResolverFixture(t *testing.T) *Resolver {
	t.Helper()
	scheduler, err := taskengine.NewScheduler(&controlPlaneFakeEngine{}, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
 	controlPlaneService, err := applicationcontrolplane.NewService(scheduler, &controlPlaneFakeQueryRepository{}, &controlPlaneFakeProjectRepository{}, &controlPlaneFakeDeadLetterManager{})
	if err != nil {
		t.Fatalf("new control-plane service: %v", err)
	}
	streamService, err := applicationstream.NewService(&controlPlaneMemoryStreamStore{})
	if err != nil {
		t.Fatalf("new stream service: %v", err)
	}
 	return NewResolver(scheduler, controlPlaneService, streamService, nil)
}

func TestControlPlaneSessionsQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&queryResolver{resolver}).Sessions(context.Background(), nil)
	if err != nil {
		t.Fatalf("Sessions() error = %v", err)
	}
	success, ok := result.(models.SessionsSuccess)
	if !ok {
		t.Fatalf("expected SessionsSuccess, got %T", result)
	}
	if len(success.Sessions) != 1 || success.Sessions[0].RunID != "run-1" {
		t.Fatalf("unexpected sessions payload: %+v", success.Sessions)
	}
}

func TestControlPlaneMutationsReturnTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	deleteResult, deleteErr := (&mutationResolver{resolver}).DeleteProjectSetup(context.Background(), models.DeleteProjectSetupInput{ProjectID: "project-1"})
	if deleteErr != nil {
		t.Fatalf("DeleteProjectSetup() error = %v", deleteErr)
	}
	if _, ok := deleteResult.(models.DeleteProjectSetupSuccess); !ok {
		t.Fatalf("expected DeleteProjectSetupSuccess, got %T", deleteResult)
	}
}

func TestControlPlaneAgentOutputSubscriptionPublishesTypedUnionEvent(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
 	runID := "run-1"
 	taskID := "task-1"
 	jobID := "job-1"
 	stream, err := (&subscriptionResolver{resolver}).AgentOutputStream(ctx, models.CorrelationInput{RunID: &runID, TaskID: &taskID, JobID: &jobID}, nil)
	if err != nil {
		t.Fatalf("AgentOutputStream() error = %v", err)
	}
	if _, appendErr := resolver.StreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "event-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceACP,
		EventType:  domainstream.EventAgentChunk,
		CorrelationIDs: domainstream.CorrelationIDs{
			RunID:         "run-1",
			TaskID:        "task-1",
			JobID:         "job-1",
			CorrelationID: "corr-1",
		},
		Payload: map[string]any{"chunk": "hello"},
	}); appendErr != nil {
		t.Fatalf("append stream event: %v", appendErr)
	}
	select {
	case message, ok := <-stream:
		if !ok {
			t.Fatalf("expected open stream channel")
		}
		success, ok := message.(models.StreamEventSuccess)
		if !ok {
			t.Fatalf("expected StreamEventSuccess, got %T", message)
		}
		if success.Event == nil || success.Event.EventID != "event-1" {
			t.Fatalf("unexpected stream event: %+v", success.Event)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for stream event")
	}
}

func TestControlPlaneTaskboardSubscriptionPublishesOnProjectSetupUpsert(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectID := "project-1"
 	stream, err := (&subscriptionResolver{resolver}).TaskboardStream(ctx, models.CorrelationInput{ProjectID: &projectID}, nil)
	if err != nil {
		t.Fatalf("TaskboardStream() error = %v", err)
	}

	_, mutationErr := (&mutationResolver{resolver}).UpsertProjectSetup(context.Background(), models.UpsertProjectSetupInput{
		ProjectID:   "project-1",
		ProjectName: "Project One",
		Scms: []*models.ProjectSCMInput{{
			ScmID:       "scm-1",
			ScmProvider: models.SCMProviderGithub,
			ScmToken:    "token",
		}},
		Repositories: []*models.ProjectRepositoryInput{{
			RepositoryID:  "repo-1",
			ScmID:         "scm-1",
			RepositoryURL: "https://github.com/acme/repo",
			IsPrimary:     true,
		}},
		Boards: []*models.ProjectBoardInput{{
			TrackerProvider:          models.TrackerSourceKindInternal,
			TaskboardName:            strPtr("Acme Repo Board"),
			AppliesToAllRepositories: true,
			RepositoryIDs:            []string{},
		}},
	})
	if mutationErr != nil {
		t.Fatalf("UpsertProjectSetup() error = %v", mutationErr)
	}

	select {
	case message, ok := <-stream:
		if !ok {
			t.Fatalf("expected open stream channel")
		}
		success, ok := message.(models.StreamEventSuccess)
		if !ok {
			t.Fatalf("expected StreamEventSuccess, got %T", message)
		}
		if success.Event == nil || success.Event.EventType != string(domainstream.EventTaskboardUpdated) {
			t.Fatalf("unexpected stream event: %+v", success.Event)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for taskboard stream event")
	}
}

func TestControlPlaneTaskboardSubscriptionReceivesCrossProcessEvents(t *testing.T) {
	store := &controlPlaneMemoryStreamStore{}
	subscriberStreamService, err := applicationstream.NewService(store)
	if err != nil {
		t.Fatalf("new subscriber stream service: %v", err)
	}
	publisherStreamService, err := applicationstream.NewService(store)
	if err != nil {
		t.Fatalf("new publisher stream service: %v", err)
	}

	scheduler, err := taskengine.NewScheduler(&controlPlaneFakeEngine{}, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
 	controlPlaneService, err := applicationcontrolplane.NewService(scheduler, &controlPlaneFakeQueryRepository{}, &controlPlaneFakeProjectRepository{}, &controlPlaneFakeDeadLetterManager{})
	if err != nil {
		t.Fatalf("new control-plane service: %v", err)
	}
 	resolver := NewResolver(scheduler, controlPlaneService, subscriberStreamService, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	projectID := "project-1"
 	stream, err := (&subscriptionResolver{resolver}).TaskboardStream(ctx, models.CorrelationInput{ProjectID: &projectID}, nil)
	if err != nil {
		t.Fatalf("TaskboardStream() error = %v", err)
	}

	_, appendErr := publisherStreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "cross-process-taskboard-event",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventTaskboardUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			CorrelationID: "project:project-1",
		},
		Payload: map[string]any{"project_id": projectID, "board_id": "board-1"},
	})
	if appendErr != nil {
		t.Fatalf("append stream event: %v", appendErr)
	}

	select {
	case message, ok := <-stream:
		if !ok {
			t.Fatalf("expected open stream channel")
		}
		success, ok := message.(models.StreamEventSuccess)
		if !ok {
			t.Fatalf("expected StreamEventSuccess, got %T", message)
		}
		if success.Event == nil || success.Event.EventID != "cross-process-taskboard-event" {
			t.Fatalf("unexpected stream event: %+v", success.Event)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for taskboard stream event")
	}
}


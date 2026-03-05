package resolvers

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	applicationstream "agentic-orchestrator/internal/application/stream"
	"agentic-orchestrator/internal/application/taskengine"
	domainlifecycle "agentic-orchestrator/internal/domain/lifecycle"
	domainstream "agentic-orchestrator/internal/domain/stream"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"strings"
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

type controlPlaneFakeLifecycleEventService struct{}

func (service *controlPlaneFakeLifecycleEventService) AppendEvent(_ context.Context, event domainlifecycle.Event) (domainlifecycle.Event, error) {
	event.EventSeq = 7
	event.ProjectEventSeq = 42
	return event, nil
}

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

func (repository *controlPlaneFakeQueryRepository) ListLifecycleSessionSnapshots(ctx context.Context, projectID string, pipelineType string, limit int) ([]applicationcontrolplane.LifecycleSessionSnapshot, error) {
	_ = ctx
	_ = pipelineType
	_ = limit
	return []applicationcontrolplane.LifecycleSessionSnapshot{{
		ProjectID:           projectID,
		RunID:               "run-1",
		TaskID:              "task-1",
		JobID:               "job-1",
		SessionID:           "session-1",
		PipelineType:        "agent",
		SourceRuntime:       "worker",
		CurrentState:        "healthy_active",
		CurrentSeverity:     "info",
		LastReasonCode:      "active",
		LastReasonSummary:   "healthy",
		LastEventSeq:        2,
		LastProjectEventSeq: 4,
		StartedAt:           time.Unix(1700000000, 0).UTC(),
		UpdatedAt:           time.Unix(1700000000, 0).UTC(),
	}}, nil
}

func (repository *controlPlaneFakeQueryRepository) ListLifecycleSessionHistory(ctx context.Context, projectID string, sessionID string, fromEventSeq int64, limit int) ([]applicationcontrolplane.LifecycleHistoryEvent, error) {
	_ = ctx
	_ = limit
	if projectID != "project-1" || sessionID != "session-1" || fromEventSeq >= 2 {
		return nil, nil
	}
	return []applicationcontrolplane.LifecycleHistoryEvent{{
		EventID:         "lifecycle-event-1",
		ProjectID:       "project-1",
		RunID:           "run-1",
		TaskID:          "task-1",
		JobID:           "job-1",
		SessionID:       "session-1",
		PipelineType:    "agent",
		SourceRuntime:   "worker",
		EventType:       "manual_retry",
		EventSeq:        2,
		ProjectEventSeq: 4,
		OccurredAt:      time.Unix(1700000005, 0).UTC(),
		PayloadJSON:     `{"ok":true}`,
	}, {
		EventID:         "lifecycle-event-2",
		ProjectID:       "project-1",
		RunID:           "run-1",
		TaskID:          "task-1",
		JobID:           "job-1",
		SessionID:       "session-1",
		PipelineType:    "agent",
		SourceRuntime:   "worker",
		EventType:       "completed",
		EventSeq:        3,
		ProjectEventSeq: 5,
		OccurredAt:      time.Unix(1700000015, 0).UTC(),
		PayloadJSON:     `{"ok":true}`,
	}}, nil
}

func (repository *controlPlaneFakeQueryRepository) ListLifecycleTreeNodes(ctx context.Context, filter applicationcontrolplane.LifecycleTreeFilter, limit int) ([]applicationcontrolplane.LifecycleTreeNode, error) {
	_ = ctx
	_ = limit
	return []applicationcontrolplane.LifecycleTreeNode{{
		NodeID:          "run:run-1",
		ParentNodeID:    "",
		NodeType:        applicationcontrolplane.LifecycleTreeNodeTypeRun,
		ProjectID:       filter.ProjectID,
		RunID:           "run-1",
		CurrentState:    "healthy_active",
		CurrentSeverity: "info",
		SessionCount:    1,
		UpdatedAt:       time.Unix(1700000000, 0).UTC(),
	}}, nil
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
	controlPlaneService.SetLifecycleService(&controlPlaneFakeLifecycleEventService{})
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
	// Simulate API table-change wake signal for cross-process replay.
	subscriberStreamService.NotifyExternalChange()

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

func TestControlPlaneProjectEventsSubscriptionReplaysCrossProcessEventsInPersistedOrder(t *testing.T) {
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
	stream, err := (&subscriptionResolver{resolver}).ProjectEventsStream(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("ProjectEventsStream() error = %v", err)
	}

	firstEvent, appendErr := publisherStreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "cross-process-project-event-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionStarted,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			SessionID:     "session-1",
			CorrelationID: "project:project-1",
		},
		Payload: map[string]any{
			"project_id":       projectID,
			"position":         1,
			"runtime_activity": true,
			"runtime_event":    "started",
		},
	})
	if appendErr != nil {
		t.Fatalf("append stream event #1: %v", appendErr)
	}
	secondEvent, appendErr := publisherStreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "cross-process-project-event-2",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventSessionHealth,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			SessionID:     "session-1",
			CorrelationID: "project:project-1",
		},
		Payload: map[string]any{
			"project_id":       projectID,
			"position":         2,
			"runtime_activity": true,
			"runtime_event":    "heartbeat",
		},
	})
	if appendErr != nil {
		t.Fatalf("append stream event #2: %v", appendErr)
	}

	// Simulate API live fan-out notification for cross-runtime delivery.
	subscriberStreamService.PublishLive(firstEvent)
	subscriberStreamService.PublishLive(secondEvent)

	receivedByID := map[string]models.StreamEventSuccess{}
	for len(receivedByID) < 2 {
		select {
		case message, ok := <-stream:
			if !ok {
				t.Fatalf("expected open stream channel")
			}
			success, ok := message.(models.StreamEventSuccess)
			if !ok {
				t.Fatalf("expected StreamEventSuccess, got %T", message)
			}
			if success.Event != nil {
				if success.Event.EventID == "cross-process-project-event-1" || success.Event.EventID == "cross-process-project-event-2" {
					receivedByID[success.Event.EventID] = success
				}
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for project events stream")
		}
	}

	firstReceived := receivedByID["cross-process-project-event-1"]
	secondReceived := receivedByID["cross-process-project-event-2"]
	if firstReceived.Event == nil {
		t.Fatalf("missing first cross-process project event")
	}
	if secondReceived.Event == nil {
		t.Fatalf("missing second cross-process project event")
	}
	if secondReceived.Event.StreamOffset <= firstReceived.Event.StreamOffset {
		t.Fatalf("expected strictly increasing stream offsets, got %d then %d", firstReceived.Event.StreamOffset, secondReceived.Event.StreamOffset)
	}
}

func TestControlPlaneProjectEventsSubscriptionAllowsEphemeralLiveEventsWithoutStreamOffset(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectID := "project-1"
	stream, err := (&subscriptionResolver{resolver}).ProjectEventsStream(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("ProjectEventsStream() error = %v", err)
	}

	// Live runtime fan-out signals may be published before persistence and carry offset 0.
	resolver.StreamService.PublishLive(domainstream.Event{
		EventID:      "runtime-live-ephemeral-1",
		StreamOffset: 0,
		OccurredAt:   time.Now().UTC(),
		Source:       domainstream.SourceWorker,
		EventType:    domainstream.EventSessionStarted,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			SessionID:     "session-ephemeral-1",
			CorrelationID: "session:session-ephemeral-1",
		},
		Payload: map[string]any{"runtime_activity": true, "runtime_event": "started"},
	})

	for {
		select {
		case message, ok := <-stream:
			if !ok {
				t.Fatalf("expected open stream channel")
			}
			success, ok := message.(models.StreamEventSuccess)
			if !ok {
				t.Fatalf("expected StreamEventSuccess, got %T", message)
			}
			if success.Event != nil && success.Event.EventID == "runtime-live-ephemeral-1" {
				return
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for ephemeral live project event")
		}
	}
}

func TestControlPlaneProjectEventsSubscriptionAcceptsWorkerLifecycleStartedWithoutRuntimeActivityFlag(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectID := "project-1"
	stream, err := (&subscriptionResolver{resolver}).ProjectEventsStream(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("ProjectEventsStream() error = %v", err)
	}

	// Lifecycle table-change fan-out can emit started semantics without runtime_activity.
	resolver.StreamService.PublishLive(domainstream.Event{
		EventID:      "lifecycle-started-live-1",
		StreamOffset: 0,
		OccurredAt:   time.Now().UTC(),
		Source:       domainstream.SourceWorker,
		EventType:    domainstream.EventSessionStarted,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			SessionID:     "session-lifecycle-1",
			CorrelationID: "session:session-lifecycle-1",
		},
		Payload: map[string]any{"lifecycle_event_type": "started"},
	})

	for {
		select {
		case message, ok := <-stream:
			if !ok {
				t.Fatalf("expected open stream channel")
			}
			success, ok := message.(models.StreamEventSuccess)
			if !ok {
				t.Fatalf("expected StreamEventSuccess, got %T", message)
			}
			if success.Event != nil && success.Event.EventID == "lifecycle-started-live-1" {
				return
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for lifecycle started live project event")
		}
	}
}

func TestControlPlaneProjectEventsSubscriptionSeedsActiveSnapshotImmediately(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectID := "project-1"
	stream, err := (&subscriptionResolver{resolver}).ProjectEventsStream(ctx, projectID, nil)
	if err != nil {
		t.Fatalf("ProjectEventsStream() error = %v", err)
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
		if success.Event == nil {
			t.Fatalf("expected seeded event payload")
		}
		if success.Event.SessionID == nil || *success.Event.SessionID != "session-1" {
			t.Fatalf("expected seeded session-1 event, got %+v", success.Event.SessionID)
		}
		if success.Event.EventType != string(domainstream.EventSessionHealth) {
			t.Fatalf("expected seeded stream.session.health event, got %q", success.Event.EventType)
		}
		if !strings.Contains(success.Event.EventID, "seed:session-1:") {
			t.Fatalf("expected seeded event id prefix, got %q", success.Event.EventID)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for seeded project event")
	}
}

func TestControlPlaneProjectEventsSubscriptionSkipsBootstrapWhenOffsetProvided(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	projectID := "project-1"
	fromOffset := int32(1)
	stream, err := (&subscriptionResolver{resolver}).ProjectEventsStream(ctx, projectID, &fromOffset)
	if err != nil {
		t.Fatalf("ProjectEventsStream() error = %v", err)
	}

	select {
	case message := <-stream:
		t.Fatalf("expected no bootstrap messages for offset replay subscription, got %T", message)
	case <-ctx.Done():
		// Expected: no live traffic and no bootstrap when fromOffset > 0.
	}
}

func TestControlPlaneProjectEventsQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	projectID := "project-1"
	_, appendErr := resolver.StreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "project-events-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventTaskboardUpdated,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     projectID,
			CorrelationID: "project:project-1",
		},
		Payload: map[string]any{"project_id": projectID, "gap_detected": true, "expected_event_seq": 3, "observed_event_seq": 2},
	})
	if appendErr != nil {
		t.Fatalf("append stream event: %v", appendErr)
	}

	result, err := (&queryResolver{resolver}).ProjectEvents(context.Background(), projectID, nil, nil)
	if err != nil {
		t.Fatalf("ProjectEvents() error = %v", err)
	}
	success, ok := result.(models.StreamEventsSuccess)
	if !ok {
		t.Fatalf("expected StreamEventsSuccess, got %T", result)
	}
	if len(success.Events) == 0 {
		t.Fatalf("expected at least one stream event")
	}
	if !success.Events[0].GapDetected {
		t.Fatalf("expected gapDetected=true in stream output")
	}
	if success.Events[0].ExpectedEventSeq == nil || *success.Events[0].ExpectedEventSeq != 3 {
		t.Fatalf("expected expectedEventSeq=3, got %+v", success.Events[0].ExpectedEventSeq)
	}
	if success.Events[0].ObservedEventSeq == nil || *success.Events[0].ObservedEventSeq != 2 {
		t.Fatalf("expected observedEventSeq=2, got %+v", success.Events[0].ObservedEventSeq)
	}
	if success.NextFromOffset < 1 {
		t.Fatalf("expected nextFromOffset >= 1, got %d", success.NextFromOffset)
	}
}

func TestControlPlanePipelineEventsSubscriptionFiltersCorrelation(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	runID := "run-1"
	projectID := "project-1"
	stream, err := (&subscriptionResolver{resolver}).PipelineEventsStream(ctx, models.CorrelationInput{ProjectID: &projectID, RunID: &runID}, nil)
	if err != nil {
		t.Fatalf("PipelineEventsStream() error = %v", err)
	}

	_, appendErr := resolver.StreamService.AppendAndPublish(context.Background(), domainstream.Event{
		EventID:    "pipeline-events-1",
		OccurredAt: time.Now().UTC(),
		Source:     domainstream.SourceWorker,
		EventType:  domainstream.EventToolStarted,
		CorrelationIDs: domainstream.CorrelationIDs{
			ProjectID:     "project-1",
			RunID:         "run-1",
			CorrelationID: "run:run-1",
		},
		Payload: map[string]any{"phase": "start"},
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
		if success.Event == nil || success.Event.EventID != "pipeline-events-1" {
			t.Fatalf("unexpected stream event: %+v", success.Event)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for stream event")
	}
}

func TestControlPlanePipelineEventsQueryRequiresProjectScope(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	runID := "run-1"
	result, err := (&queryResolver{resolver}).PipelineEvents(context.Background(), models.CorrelationInput{RunID: &runID}, nil, nil)
	if err != nil {
		t.Fatalf("PipelineEvents() error = %v", err)
	}
	graphErr, ok := result.(models.GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", result)
	}
	if graphErr.Message == "" {
		t.Fatalf("expected non-empty graph error message")
	}
}

func TestControlPlanePipelineEventsSubscriptionRequiresProjectScope(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	runID := "run-1"
	stream, err := (&subscriptionResolver{resolver}).PipelineEventsStream(context.Background(), models.CorrelationInput{RunID: &runID}, nil)
	if err != nil {
		t.Fatalf("PipelineEventsStream() error = %v", err)
	}
	message, ok := <-stream
	if !ok {
		t.Fatalf("expected one graph error message")
	}
	if _, isGraphError := message.(models.GraphError); !isGraphError {
		t.Fatalf("expected GraphError, got %T", message)
	}
}

func TestControlPlaneLifecycleSessionSnapshotsQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&queryResolver{resolver}).LifecycleSessionSnapshots(context.Background(), "project-1", nil, nil)
	if err != nil {
		t.Fatalf("LifecycleSessionSnapshots() error = %v", err)
	}
	success, ok := result.(models.LifecycleSessionSnapshotsSuccess)
	if !ok {
		t.Fatalf("expected LifecycleSessionSnapshotsSuccess, got %T", result)
	}
	if len(success.Sessions) != 1 || success.Sessions[0].SessionID != "session-1" {
		t.Fatalf("unexpected lifecycle sessions payload: %+v", success.Sessions)
	}
}

func TestControlPlaneLifecycleSessionHistoryQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&queryResolver{resolver}).LifecycleSessionHistory(context.Background(), "project-1", "session-1", nil, nil)
	if err != nil {
		t.Fatalf("LifecycleSessionHistory() error = %v", err)
	}
	success, ok := result.(models.LifecycleHistorySuccess)
	if !ok {
		t.Fatalf("expected LifecycleHistorySuccess, got %T", result)
	}
	if len(success.Events) != 2 || success.Events[0].EventID != "lifecycle-event-1" {
		t.Fatalf("unexpected lifecycle history payload: %+v", success.Events)
	}
	if success.NextFromEventSeq < 3 {
		t.Fatalf("expected nextFromEventSeq >= 3, got %d", success.NextFromEventSeq)
	}
}

func TestControlPlaneLifecycleTreeNodesQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&queryResolver{resolver}).LifecycleTreeNodes(context.Background(), models.LifecycleTreeFilterInput{ProjectID: "project-1"}, nil)
	if err != nil {
		t.Fatalf("LifecycleTreeNodes() error = %v", err)
	}
	success, ok := result.(models.LifecycleTreeNodesSuccess)
	if !ok {
		t.Fatalf("expected LifecycleTreeNodesSuccess, got %T", result)
	}
	if len(success.Nodes) != 1 || success.Nodes[0].NodeID != "run:run-1" {
		t.Fatalf("unexpected lifecycle tree payload: %+v", success.Nodes)
	}
}

func TestControlPlaneApplyManualInterventionMutationReturnsSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&mutationResolver{resolver}).ApplyManualIntervention(context.Background(), models.ApplyManualInterventionInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		Action:    models.ManualInterventionActionRetry,
		Reason:    "Retry after transient checkout failure",
		ActorID:   "operator:alice",
	})
	if err != nil {
		t.Fatalf("ApplyManualIntervention() error = %v", err)
	}
	success, ok := result.(models.ManualInterventionSuccess)
	if !ok {
		t.Fatalf("expected ManualInterventionSuccess, got %T", result)
	}
	if !success.Ok || success.EventID == "" || success.ProjectEventSeq == 0 {
		t.Fatalf("unexpected mutation payload: %+v", success)
	}
}

func TestControlPlaneApplyManualInterventionMutationEnforcesAuthorization(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	force := true
	result, err := (&mutationResolver{resolver}).ApplyManualIntervention(context.Background(), models.ApplyManualInterventionInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		Action:    models.ManualInterventionActionTerminate,
		Reason:    "Terminate due runaway execution impacting shared workers",
		ActorID:   "operator:alice",
		Force:     &force,
	})
	if err != nil {
		t.Fatalf("ApplyManualIntervention() error = %v", err)
	}
	graphErr, ok := result.(models.GraphError)
	if !ok {
		t.Fatalf("expected GraphError, got %T", result)
	}
	if graphErr.Code != models.GraphErrorCodeForbidden {
		t.Fatalf("expected FORBIDDEN, got %s", graphErr.Code)
	}
}

func TestControlPlaneInterventionMetricsQueryReturnsTypedUnionSuccess(t *testing.T) {
	resolver := newControlPlaneResolverFixture(t)
	result, err := (&queryResolver{resolver}).InterventionMetrics(context.Background(), "project-1", nil)
	if err != nil {
		t.Fatalf("InterventionMetrics() error = %v", err)
	}
	success, ok := result.(models.InterventionMetricsSuccess)
	if !ok {
		t.Fatalf("expected InterventionMetricsSuccess, got %T", result)
	}
	if success.Metrics == nil || success.Metrics.InterventionCount < 1 {
		t.Fatalf("unexpected intervention metrics payload: %+v", success.Metrics)
	}
	if success.Metrics.SuccessfulOutcomeCount < 1 {
		t.Fatalf("expected at least one successful intervention outcome: %+v", success.Metrics)
	}
}

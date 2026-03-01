package resolvers

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	"agentic-orchestrator/internal/interface/graphql/models"
	workerinterface "agentic-orchestrator/internal/interface/worker"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type integrationQueueEngine struct {
	requests            []taskengine.EnqueueRequest
	seenIdempotencyKeys map[string]struct{}
}

func (engine *integrationQueueEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	if request.IdempotencyKey != "" {
		if engine.seenIdempotencyKeys == nil {
			engine.seenIdempotencyKeys = make(map[string]struct{})
		}
		if _, exists := engine.seenIdempotencyKeys[request.IdempotencyKey]; exists {
			return taskengine.EnqueueResult{QueueTaskID: request.IdempotencyKey, Duplicate: true}, nil
		}
		engine.seenIdempotencyKeys[request.IdempotencyKey] = struct{}{}
	}
	engine.requests = append(engine.requests, request)
	queueTaskID := fmt.Sprintf("queue-task-%d", len(engine.requests))
	if request.IdempotencyKey != "" {
		queueTaskID = request.IdempotencyKey
	}
	return taskengine.EnqueueResult{QueueTaskID: queueTaskID}, nil
}

func (engine *integrationQueueEngine) dispatchNext(ctx context.Context, handler taskengine.Handler) error {
	if len(engine.requests) == 0 {
		return fmt.Errorf("no queued requests")
	}
	request := engine.requests[0]
	engine.requests = engine.requests[1:]
	return handler.Handle(ctx, taskengine.Job{
		Kind:        request.Kind,
		QueueTaskID: request.IdempotencyKey,
		Payload:     request.Payload,
	})
}

type failIfCalledIntegrationGitRunner struct{}

func (runner *failIfCalledIntegrationGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	_ = directory
	return "", fmt.Errorf("unexpected git runner call: %v", arguments)
}

type scmAdmissionQueueWorkerAdapterFixture struct {
	resolver *Resolver
	queue    *integrationQueueEngine
	handler  taskengine.Handler

	repositoryEndpointCalled bool
	commitEndpointCalled     bool
	close                    func()
}

func newSCMAdmissionQueueWorkerAdapterFixture(t *testing.T) *scmAdmissionQueueWorkerAdapterFixture {
	t.Helper()

	fixture := &scmAdmissionQueueWorkerAdapterFixture{queue: &integrationQueueEngine{}}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/acme/repo":
			fixture.repositoryEndpointCalled = true
			_, _ = writer.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/acme/repo/commits/main":
			fixture.commitEndpointCalled = true
			_, _ = writer.Write([]byte(`{"sha":"abc123"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	fixture.close = server.Close

	adapter, err := infrascm.NewGitHubAdapter(
		infrascm.GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"},
		server.Client(),
		infrascm.NewStaticTokenProvider("token"),
		&failIfCalledIntegrationGitRunner{},
	)
	if err != nil {
		t.Fatalf("new github adapter: %v", err)
	}
	service, err := applicationscm.NewService(adapter)
	if err != nil {
		t.Fatalf("new scm service: %v", err)
	}
	handler, err := workerinterface.NewSCMWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new scm workflow handler: %v", err)
	}

	scheduler, err := taskengine.NewScheduler(fixture.queue, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	fixture.resolver = NewResolver(scheduler, nil, nil, nil, nil)
	fixture.handler = handler
	return fixture
}

func TestSCMAdmissionQueueWorkerAdapterFixtureSourceStatePath(t *testing.T) {
	fixture := newSCMAdmissionQueueWorkerAdapterFixture(t)
	defer fixture.close()

	result, err := fixture.resolver.Mutation().EnqueueScmWorkflow(context.Background(), models.EnqueueSCMWorkflowInput{
		Operation:      models.SCMOperationSourceState,
		Provider:       models.SCMProviderGithub,
		Owner:          "acme",
		Repository:     "repo",
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
	})
	if err != nil {
		t.Fatalf("enqueue scm workflow: %v", err)
	}
	success, ok := result.(models.EnqueueSCMWorkflowSuccess)
	if !ok {
		t.Fatalf("expected enqueue success result, got %T", result)
	}
	if success.QueueTaskID != "id-1" {
		t.Fatalf("expected queue task id id-1, got %q", success.QueueTaskID)
	}
	if len(fixture.queue.requests) != 1 {
		t.Fatalf("expected one queued request, got %d", len(fixture.queue.requests))
	}
	if fixture.queue.requests[0].Queue != "scm" {
		t.Fatalf("expected scm queue, got %q", fixture.queue.requests[0].Queue)
	}
	if fixture.queue.requests[0].IdempotencyKey != "id-1" {
		t.Fatalf("expected deterministic idempotency key id-1, got %q", fixture.queue.requests[0].IdempotencyKey)
	}

	if err := fixture.queue.dispatchNext(context.Background(), fixture.handler); err != nil {
		t.Fatalf("dispatch queued scm workflow: %v", err)
	}
	if !fixture.repositoryEndpointCalled || !fixture.commitEndpointCalled {
		t.Fatalf("expected API admission -> queue -> worker -> scm adapter path to call repository and commit endpoints")
	}
}

func TestSCMAdmissionEnqueuesExactlyOneJobForDeterministicIdempotencyKey(t *testing.T) {
	fixture := newSCMAdmissionQueueWorkerAdapterFixture(t)
	defer fixture.close()

	input := models.EnqueueSCMWorkflowInput{
		Operation:      models.SCMOperationSourceState,
		Provider:       models.SCMProviderGithub,
		Owner:          "acme",
		Repository:     "repo",
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-deterministic",
	}

	first, firstErr := fixture.resolver.Mutation().EnqueueScmWorkflow(context.Background(), input)
	if firstErr != nil {
		t.Fatalf("first enqueue scm workflow: %v", firstErr)
	}
	firstSuccess, ok := first.(models.EnqueueSCMWorkflowSuccess)
	if !ok {
		t.Fatalf("expected first enqueue success result, got %T", first)
	}
	if firstSuccess.QueueTaskID != "id-deterministic" {
		t.Fatalf("expected first queue task id id-deterministic, got %q", firstSuccess.QueueTaskID)
	}
	if firstSuccess.Duplicate {
		t.Fatalf("expected first enqueue to not be duplicate")
	}

	second, secondErr := fixture.resolver.Mutation().EnqueueScmWorkflow(context.Background(), input)
	if secondErr != nil {
		t.Fatalf("second enqueue scm workflow: %v", secondErr)
	}
	secondSuccess, ok := second.(models.EnqueueSCMWorkflowSuccess)
	if !ok {
		t.Fatalf("expected second enqueue success result, got %T", second)
	}
	if secondSuccess.QueueTaskID != "id-deterministic" {
		t.Fatalf("expected second queue task id id-deterministic, got %q", secondSuccess.QueueTaskID)
	}
	if !secondSuccess.Duplicate {
		t.Fatalf("expected second enqueue to be marked duplicate")
	}
	if len(fixture.queue.requests) != 1 {
		t.Fatalf("expected exactly one queued request, got %d", len(fixture.queue.requests))
	}
	if fixture.queue.requests[0].IdempotencyKey != "id-deterministic" {
		t.Fatalf("expected queued idempotency key id-deterministic, got %q", fixture.queue.requests[0].IdempotencyKey)
	}
}

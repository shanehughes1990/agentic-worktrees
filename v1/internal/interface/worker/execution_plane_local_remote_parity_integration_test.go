package worker

import (
	applicationtaskengine "agentic-orchestrator/internal/application/taskengine"
	applicationscm "agentic-orchestrator/internal/application/scm"
	domainscm "agentic-orchestrator/internal/domain/scm"
	infraagent "agentic-orchestrator/internal/infrastructure/agent"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type parityQueueEngine struct {
	requests []applicationtaskengine.EnqueueRequest
}

func (engine *parityQueueEngine) Enqueue(ctx context.Context, request applicationtaskengine.EnqueueRequest) (applicationtaskengine.EnqueueResult, error) {
	_ = ctx
	engine.requests = append(engine.requests, request)
	return applicationtaskengine.EnqueueResult{QueueTaskID: fmt.Sprintf("queue-task-%d", len(engine.requests))}, nil
}

func (engine *parityQueueEngine) dispatchNext(ctx context.Context, handler applicationtaskengine.Handler) error {
	if len(engine.requests) == 0 {
		return fmt.Errorf("no queued requests")
	}
	request := engine.requests[0]
	engine.requests = engine.requests[1:]
	return handler.Handle(ctx, applicationtaskengine.Job{
		Kind:        request.Kind,
		QueueTaskID: request.IdempotencyKey,
		Payload:     request.Payload,
	})
}

type parityGitRunner struct {
	worktreeAddCalls int
	revParseCalls    int
	fetchCalls       int
}

func (runner *parityGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	if len(arguments) == 0 {
		return "", fmt.Errorf("git arguments are required")
	}
	if arguments[0] == "fetch" {
		runner.fetchCalls++
		return "", nil
	}
	if arguments[0] == "worktree" && len(arguments) >= 6 && arguments[1] == "add" {
		runner.worktreeAddCalls++
		return "", nil
	}
	if arguments[0] == "rev-parse" {
		runner.revParseCalls++
		return "abc123", nil
	}
	return "", fmt.Errorf("unexpected git command in %s: %s", directory, strings.Join(arguments, " "))
}

type scmBootstrapAdapter struct {
	service *applicationscm.Service

	metadata applicationscm.Metadata
}

func (adapter *scmBootstrapAdapter) SourceState(ctx context.Context, repository domainscm.Repository) (domainscm.SourceState, error) {
	return adapter.service.SourceState(ctx, applicationscm.SourceStateRequest{Repository: repository, Metadata: adapter.metadata})
}

func (adapter *scmBootstrapAdapter) EnsureWorktree(ctx context.Context, repository domainscm.Repository, spec domainscm.WorktreeSpec) (domainscm.WorktreeState, error) {
	return adapter.service.EnsureWorktree(ctx, applicationscm.EnsureWorktreeRequest{Repository: repository, Spec: spec, Metadata: adapter.metadata})
}

func TestExecutionPlaneLocalAndRemotePathsShareSCMJobContract(t *testing.T) {
	repositoryEndpointCalls := 0
	commitEndpointCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/acme/repo":
			repositoryEndpointCalls++
			_, _ = writer.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/acme/repo/commits/main":
			commitEndpointCalls++
			_, _ = writer.Write([]byte(`{"sha":"abc123"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	gitRunner := &parityGitRunner{}
	githubAdapter, err := infrascm.NewGitHubAdapter(
		infrascm.GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktree"},
		server.Client(),
		infrascm.NewStaticTokenProvider("token"),
		gitRunner,
	)
	if err != nil {
		t.Fatalf("new github adapter: %v", err)
	}
	scmService, err := applicationscm.NewService(githubAdapter)
	if err != nil {
		t.Fatalf("new scm service: %v", err)
	}
	localHandler, err := NewSCMWorkflowHandler(scmService)
	if err != nil {
		t.Fatalf("new scm workflow handler: %v", err)
	}

	queueEngine := &parityQueueEngine{}
	scheduler, err := applicationtaskengine.NewScheduler(queueEngine, applicationtaskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	payload, err := json.Marshal(SCMWorkflowPayload{
		Operation:      "source_state",
		Provider:       "github",
		Owner:          "acme",
		Repository:     "repo",
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if _, err := scheduler.Enqueue(context.Background(), applicationtaskengine.EnqueueRequest{
		Kind:           applicationtaskengine.JobKindSCMWorkflow,
		Payload:        payload,
		IdempotencyKey: "id-1",
	}); err != nil {
		t.Fatalf("enqueue local scm workflow: %v", err)
	}
	if err := queueEngine.dispatchNext(context.Background(), localHandler); err != nil {
		t.Fatalf("dispatch local scm workflow: %v", err)
	}

	remoteAdapter, err := infraagent.NewLoopbackRemoteWorkerAdapter("remote-worker-1", localHandler)
	if err != nil {
		t.Fatalf("new loopback remote adapter: %v", err)
	}
	bootstrapService, err := applicationtaskengine.NewRemoteBootstrapService(&scmBootstrapAdapter{
		service: scmService,
		metadata: applicationscm.Metadata{
			CorrelationIDs: applicationtaskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"},
			IdempotencyKey: "id-1",
		},
	}, remoteAdapter)
	if err != nil {
		t.Fatalf("new remote bootstrap service: %v", err)
	}
	remoteResult, err := bootstrapService.Execute(context.Background(), applicationtaskengine.RemoteBootstrapRequest{
		Job: applicationtaskengine.Job{
			Kind:        applicationtaskengine.JobKindSCMWorkflow,
			QueueTaskID: "remote-task-1",
			Payload:     payload,
		},
		CorrelationIDs: applicationtaskengine.CorrelationIDs{RunID: "run-1", TaskID: "task-1", JobID: "job-1"},
		IdempotencyKey: "id-1",
		Repository:     domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		TargetBranch:   "feature/remote-task-1",
		WorktreePath:   "/tmp/worktree/remote-task-1",
	})
	if err != nil {
		t.Fatalf("execute remote bootstrap workflow: %v", err)
	}
	if remoteResult.WorkerID != "remote-worker-1" {
		t.Fatalf("expected remote-worker-1, got %q", remoteResult.WorkerID)
	}
	if remoteResult.CompletedCheckpoint == nil || remoteResult.CompletedCheckpoint.Step != "ensure_worktree" {
		t.Fatalf("expected ensure_worktree completed checkpoint, got %+v", remoteResult.CompletedCheckpoint)
	}
	if gitRunner.worktreeAddCalls == 0 || gitRunner.revParseCalls == 0 || gitRunner.fetchCalls == 0 {
		t.Fatalf("expected remote bootstrap to perform SCM-backed worktree bootstrap, got fetch=%d add=%d rev-parse=%d", gitRunner.fetchCalls, gitRunner.worktreeAddCalls, gitRunner.revParseCalls)
	}
	if repositoryEndpointCalls < 2 || commitEndpointCalls < 2 {
		t.Fatalf("expected source-state SCM endpoints to be called by both local and remote paths, got repos=%d commits=%d", repositoryEndpointCalls, commitEndpointCalls)
	}
}

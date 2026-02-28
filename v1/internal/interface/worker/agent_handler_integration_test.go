package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	"agentic-orchestrator/internal/application/taskengine"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failIfCalledAgentGitRunner struct{}

func (runner *failIfCalledAgentGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	_ = directory
	return "", fmt.Errorf("unexpected git runner call: %v", arguments)
}

type queuedAgentEngine struct {
	requests []taskengine.EnqueueRequest
}

func (engine *queuedAgentEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	engine.requests = append(engine.requests, request)
	return taskengine.EnqueueResult{QueueTaskID: fmt.Sprintf("queue-task-%d", len(engine.requests))}, nil
}

func (engine *queuedAgentEngine) dispatchNext(ctx context.Context, handler taskengine.Handler) error {
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

func TestAgentWorkflowHandlerExecutesQueuedJobThroughSCMPortAdapterPath(t *testing.T) {
	repositoryEndpointCalled := false
	commitEndpointCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/acme/repo":
			repositoryEndpointCalled = true
			_, _ = writer.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/acme/repo/commits/main":
			commitEndpointCalled = true
			_, _ = writer.Write([]byte(`{"sha":"abc123"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter, err := infrascm.NewGitHubAdapter(
		infrascm.GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo"},
		server.Client(),
		infrascm.NewStaticTokenProvider("token"),
		&failIfCalledAgentGitRunner{},
	)
	if err != nil {
		t.Fatalf("new github adapter: %v", err)
	}

	agentService, err := applicationagent.NewService(adapter)
	if err != nil {
		t.Fatalf("new agent service: %v", err)
	}

	handler, err := NewAgentWorkflowHandler(agentService)
	if err != nil {
		t.Fatalf("new agent workflow handler: %v", err)
	}

	queueEngine := &queuedAgentEngine{}
	scheduler, err := taskengine.NewScheduler(queueEngine, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}

	payload, err := json.Marshal(AgentWorkflowPayload{
		SessionID:      "session-1",
		Prompt:         "run analysis",
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

	enqueueResult, err := scheduler.Enqueue(context.Background(), taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindAgentWorkflow,
		Payload:        payload,
		IdempotencyKey: "id-1",
	})
	if err != nil {
		t.Fatalf("enqueue agent workflow job: %v", err)
	}
	if enqueueResult.QueueTaskID != "queue-task-1" {
		t.Fatalf("expected queue task id queue-task-1, got %q", enqueueResult.QueueTaskID)
	}
	if len(queueEngine.requests) != 1 {
		t.Fatalf("expected one queued request, got %d", len(queueEngine.requests))
	}
	if queueEngine.requests[0].Queue != "agent" {
		t.Fatalf("expected agent queue, got %q", queueEngine.requests[0].Queue)
	}

	if err := queueEngine.dispatchNext(context.Background(), handler); err != nil {
		t.Fatalf("dispatch agent workflow job: %v", err)
	}
	if !repositoryEndpointCalled || !commitEndpointCalled {
		t.Fatalf("expected queued agent workflow to invoke SCM port via adapter source-state endpoints")
	}
}

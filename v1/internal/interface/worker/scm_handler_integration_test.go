package worker

import (
	applicationscm "agentic-orchestrator/internal/application/scm"
	"agentic-orchestrator/internal/application/taskengine"
	infrascm "agentic-orchestrator/internal/infrastructure/scm"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failIfCalledGitRunner struct{}

func (runner *failIfCalledGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	_ = directory
	return "", fmt.Errorf("unexpected git runner call: %v", arguments)
}

func TestSCMWorkflowHandlerSourceStateIntegration(t *testing.T) {
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
		&failIfCalledGitRunner{},
	)
	if err != nil {
		t.Fatalf("new github adapter: %v", err)
	}

	scmService, err := applicationscm.NewService(adapter)
	if err != nil {
		t.Fatalf("new scm service: %v", err)
	}

	handler, err := NewSCMWorkflowHandler(scmService)
	if err != nil {
		t.Fatalf("new scm workflow handler: %v", err)
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

	handleErr := handler.Handle(context.Background(), taskengine.Job{
		Kind:    taskengine.JobKindSCMWorkflow,
		Payload: payload,
	})
	if handleErr != nil {
		t.Fatalf("handle scm workflow job: %v", handleErr)
	}
	if !repositoryEndpointCalled || !commitEndpointCalled {
		t.Fatalf("expected source_state integration to call both repository and commit endpoints")
	}
}

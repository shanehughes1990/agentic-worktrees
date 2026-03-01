package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type staticTokenProvider struct {
	token string
}

func (provider staticTokenProvider) AccessToken(ctx context.Context) (string, error) {
	_ = ctx
	return provider.token, nil
}

func TestGitHubIssuesProviderSyncsBoardFromGitHubIssuesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/octo/repo/issues" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("expected authorization header")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[
			{
				"number": 42,
				"title": "Implement tracker intake",
				"body": "Support github issues provider",
				"state": "open",
				"html_url": "https://github.com/octo/repo/issues/42",
				"created_at": "2026-03-01T01:00:00Z",
				"updated_at": "2026-03-01T01:30:00Z",
				"labels": [{"name":"p1"}]
			},
			{
				"number": 99,
				"title": "PR shadow object",
				"state": "open",
				"pull_request": {"url": "https://api.github.com/repos/octo/repo/pulls/99"}
			}
		]`))
	}))
	defer server.Close()

	provider, err := NewGitHubIssuesProviderWithConfig(server.URL, server.Client(), staticTokenProvider{token: "test-token"})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	board, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: "octo/repo",
		},
	})
	if err != nil {
		t.Fatalf("SyncBoard() error = %v", err)
	}
	if board.BoardID != "github:octo/repo" {
		t.Fatalf("expected board id github:octo/repo, got %q", board.BoardID)
	}
	if len(board.Epics) != 1 || len(board.Epics[0].Tasks) != 1 {
		t.Fatalf("expected one epic with one issue task")
	}
	task := board.Epics[0].Tasks[0]
	if task.Title != "Implement tracker intake" {
		t.Fatalf("unexpected task title: %q", task.Title)
	}
	if task.Priority != domaintracker.PriorityP1 {
		t.Fatalf("expected p1 priority, got %q", task.Priority)
	}
	if task.Status != domaintracker.StatusNotStarted {
		t.Fatalf("expected not-started status, got %q", task.Status)
	}
	if reference, ok := task.Metadata["issue_reference"].(string); !ok || reference != "octo/repo#42" {
		t.Fatalf("expected issue_reference octo/repo#42, got %#v", task.Metadata["issue_reference"])
	}
}

func TestGitHubIssuesProviderRejectsInvalidLocation(t *testing.T) {
	provider := NewGitHubIssuesProvider()
	_, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: "octo",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestGitHubIssuesProviderReturnsTerminalWhenNoIssuesFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`[]`))
	}))
	defer server.Close()
	provider, err := NewGitHubIssuesProviderWithConfig(server.URL, server.Client(), nil)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	_, err = provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: "octo/repo",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal no-issues error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestJiraProviderDefinesBoundaryAndReturnsNotImplemented(t *testing.T) {
	provider := NewJiraProvider()
	_, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:    domaintracker.SourceKindJira,
			BoardID: "TEAM-1",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal not-implemented error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestLinearProviderDefinesBoundaryAndReturnsNotImplemented(t *testing.T) {
	provider := NewLinearProvider()
	_, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:    domaintracker.SourceKindLinear,
			BoardID: "TEAM-1",
		},
	})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal not-implemented error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestGitHubIssuesProviderConfigRejectsInvalidBaseURL(t *testing.T) {
	_, err := NewGitHubIssuesProviderWithConfig("://invalid-url", nil, nil)
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal config error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestGitHubIssuesProviderPropagatesNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()
	provider, err := NewGitHubIssuesProviderWithConfig(server.URL, server.Client(), nil)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	_, err = provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: "octo/repo",
		},
	})
	if err == nil {
		t.Fatalf("expected github status error")
	}
	if failures.ClassOf(err) != failures.ClassTerminal {
		t.Fatalf("expected terminal class, got %q (%v)", failures.ClassOf(err), err)
	}
	if !strings.Contains(err.Error(), "status=404") {
		t.Fatalf("expected 404 in error, got %v", err)
	}
}

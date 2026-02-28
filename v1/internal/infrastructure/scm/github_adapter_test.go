package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeGitRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (runner *fakeGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	key := directory + "::" + fmt.Sprint(arguments)
	if err, ok := runner.errors[key]; ok {
		return "", err
	}
	if output, ok := runner.outputs[key]; ok {
		return output, nil
	}
	return "", nil
}

type recordingGitRunner struct {
	calls []string
}

func (runner *recordingGitRunner) Run(ctx context.Context, directory string, arguments ...string) (string, error) {
	_ = ctx
	runner.calls = append(runner.calls, directory+"::"+fmt.Sprint(arguments))
	if len(arguments) >= 2 && arguments[0] == "rev-parse" && arguments[1] == "HEAD" {
		return "abc123", nil
	}
	return "", nil
}

func TestSourceStateReadsDefaultBranchAndHead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/repos/acme/repo":
			_, _ = writer.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/acme/repo/commits/main":
			_, _ = writer.Write([]byte(`{"sha":"abc123"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter, err := NewGitHubAdapter(GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"}, server.Client(), NewStaticTokenProvider("token"), &fakeGitRunner{})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	state, stateErr := adapter.SourceState(context.Background(), domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"})
	if stateErr != nil {
		t.Fatalf("source state: %v", stateErr)
	}
	if state.DefaultBranch != "main" || state.HeadSHA != "abc123" {
		t.Fatalf("unexpected source state: %+v", state)
	}
}

func TestEnsureWorktreeFetchesOriginBeforeWorktreeAdd(t *testing.T) {
	runner := &recordingGitRunner{}
	adapter, err := NewGitHubAdapter(
		GitHubAdapterConfig{APIBaseURL: "https://api.github.com", RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktree"},
		nil,
		NewStaticTokenProvider("token"),
		runner,
	)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, ensureErr := adapter.EnsureWorktree(context.Background(), domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"}, domainscm.WorktreeSpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/worktree/feature-one"})
	if ensureErr != nil {
		t.Fatalf("ensure worktree: %v", ensureErr)
	}
	if len(runner.calls) < 3 {
		t.Fatalf("expected at least 3 git calls, got %d (%v)", len(runner.calls), runner.calls)
	}
	expectedFetch := "/tmp/repo::[fetch origin main]"
	if runner.calls[0] != expectedFetch {
		t.Fatalf("expected first call %q, got %q", expectedFetch, runner.calls[0])
	}
	expectedWorktreeAdd := "/tmp/repo::[worktree add -B feature/one /tmp/worktree/feature-one origin/main]"
	if runner.calls[1] != expectedWorktreeAdd {
		t.Fatalf("expected second call %q, got %q", expectedWorktreeAdd, runner.calls[1])
	}
}

func TestEnsureWorktreeRejectsPathOutsideConfiguredWorktreeRoot(t *testing.T) {
	adapter, err := NewGitHubAdapter(
		GitHubAdapterConfig{APIBaseURL: "https://api.github.com", RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"},
		nil,
		NewStaticTokenProvider("token"),
		&fakeGitRunner{},
	)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, ensureErr := adapter.EnsureWorktree(context.Background(), domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"}, domainscm.WorktreeSpec{BaseBranch: "main", TargetBranch: "feature/one", Path: "/tmp/other/worktree"})
	if !failures.IsClass(ensureErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal path validation error, got %q (%v)", failures.ClassOf(ensureErr), ensureErr)
	}
}

func TestSyncWorktreeResolvesRelativePathInsideRoot(t *testing.T) {
	runner := &recordingGitRunner{}
	adapter, err := NewGitHubAdapter(
		GitHubAdapterConfig{APIBaseURL: "https://api.github.com", RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"},
		nil,
		NewStaticTokenProvider("token"),
		runner,
	)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, syncErr := adapter.SyncWorktree(context.Background(), domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"}, "feature-one")
	if syncErr != nil {
		t.Fatalf("sync worktree: %v", syncErr)
	}
	if len(runner.calls) == 0 {
		t.Fatalf("expected git calls")
	}
	if runner.calls[0] != "/tmp/worktrees/feature-one::[rev-parse --abbrev-ref HEAD]" {
		t.Fatalf("expected worktree path under root, got %q", runner.calls[0])
	}
}

func TestCreateOrUpdatePullRequestCreatesWhenNoOpenPullRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/repos/acme/repo/pulls" && request.Method == http.MethodGet:
			_, _ = writer.Write([]byte(`[]`))
		case request.URL.Path == "/repos/acme/repo/pulls" && request.Method == http.MethodPost:
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"number":42,"html_url":"https://github.com/acme/repo/pull/42","state":"open","head":{"sha":"headsha"}}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	adapter, err := NewGitHubAdapter(GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"}, server.Client(), NewStaticTokenProvider("token"), &fakeGitRunner{})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	state, createErr := adapter.CreateOrUpdatePullRequest(context.Background(), domainscm.PullRequestSpec{
		Repository:   domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		SourceBranch: "feature/one",
		TargetBranch: "main",
		Title:        "Feature one",
		Body:         "Implements feature one",
	})
	if createErr != nil {
		t.Fatalf("create pull request: %v", createErr)
	}
	if state.Number != 42 {
		t.Fatalf("expected pull request number 42, got %d", state.Number)
	}
}

func TestSubmitReviewUsesGitHubReviewEndpoint(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/repos/acme/repo/pulls/7/reviews" && request.Method == http.MethodPost {
			called = true
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"id":1}`))
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	adapter, err := NewGitHubAdapter(GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"}, server.Client(), NewStaticTokenProvider("token"), &fakeGitRunner{})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	decision, reviewErr := adapter.SubmitReview(context.Background(), domainscm.ReviewSpec{
		Repository:        domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"},
		PullRequestNumber: 7,
		Decision:          domainscm.ReviewDecisionApprove,
		Body:              "LGTM",
	})
	if reviewErr != nil {
		t.Fatalf("submit review: %v", reviewErr)
	}
	if decision != domainscm.ReviewDecisionApprove {
		t.Fatalf("expected approve decision, got %q", decision)
	}
	if !called {
		t.Fatalf("expected github review endpoint to be called")
	}
}

func TestDoJSONClassifiesRateLimitAsTransient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = writer.Write([]byte(`{"message":"rate limited"}`))
	}))
	defer server.Close()

	adapter, err := NewGitHubAdapter(GitHubAdapterConfig{APIBaseURL: server.URL, RepoPath: "/tmp/repo", WorktreeRootPath: "/tmp/worktrees"}, server.Client(), NewStaticTokenProvider("token"), &fakeGitRunner{})
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	_, sourceErr := adapter.SourceState(context.Background(), domainscm.Repository{Provider: "github", Owner: "acme", Name: "repo"})
	if !failures.IsClass(sourceErr, failures.ClassTransient) {
		t.Fatalf("expected transient classification, got %q (%v)", failures.ClassOf(sourceErr), sourceErr)
	}
}

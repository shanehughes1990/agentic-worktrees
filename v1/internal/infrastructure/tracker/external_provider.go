package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GitHubTokenProvider interface {
	AccessToken(ctx context.Context) (string, error)
}

type ExternalProvider struct {
	sourceKind    domaintracker.SourceKind
	baseURL       string
	httpClient    *http.Client
	tokenProvider GitHubTokenProvider
}

func NewGitHubIssuesProvider() *ExternalProvider {
	provider, _ := NewGitHubIssuesProviderWithConfig("", nil, nil)
	return provider
}

func NewGitHubIssuesProviderWithConfig(baseURL string, httpClient *http.Client, tokenProvider GitHubTokenProvider) (*ExternalProvider, error) {
	cleanBaseURL := strings.TrimSpace(baseURL)
	if cleanBaseURL == "" {
		cleanBaseURL = "https://api.github.com"
	}
	if _, err := url.Parse(cleanBaseURL); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("parse github api base url: %w", err))
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ExternalProvider{
		sourceKind:    domaintracker.SourceKindGitHubIssues,
		baseURL:       strings.TrimRight(cleanBaseURL, "/"),
		httpClient:    httpClient,
		tokenProvider: tokenProvider,
	}, nil
}

func (provider *ExternalProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if request.Source.Kind != provider.sourceKind {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("external provider kind %q does not support source kind %q", provider.sourceKind, request.Source.Kind))
	}
	if provider.sourceKind == domaintracker.SourceKindGitHubIssues {
		return provider.syncGitHubIssues(ctx, request)
	}
	return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("%s tracker provider sync boundary is defined but not implemented", provider.sourceKind))
}

func (provider *ExternalProvider) syncGitHubIssues(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	owner, repository, err := parseGitHubLocation(request.Source.Location)
	if err != nil {
		return domaintracker.Board{}, err
	}
	issues, err := provider.fetchGitHubIssues(ctx, owner, repository)
	if err != nil {
		return domaintracker.Board{}, err
	}
	if len(issues) == 0 {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("no github issues found for %s/%s", owner, repository))
	}

	boardID := strings.TrimSpace(request.Source.BoardID)
	if boardID == "" {
		boardID = fmt.Sprintf("github:%s/%s", owner, repository)
	}
	now := time.Now().UTC()
	tasks := make([]domaintracker.Task, 0, len(issues))
	for _, issue := range issues {
		if issue.Number <= 0 || strings.TrimSpace(issue.Title) == "" {
			continue
		}
		priority := extractPriority(issue.Labels)
		issueReference := fmt.Sprintf("%s/%s#%d", owner, repository, issue.Number)
		task := domaintracker.Task{
			WorkItem: domaintracker.WorkItem{
				ID:          domaintracker.WorkItemID(fmt.Sprintf("gh-issue-%d", issue.Number)),
				BoardID:     boardID,
				Title:       strings.TrimSpace(issue.Title),
				Description: strings.TrimSpace(issue.Body),
				Status:      statusFromGitHubIssueState(issue.State),
				Priority:    priority,
				Metadata: map[string]any{
					"source":          string(domaintracker.SourceKindGitHubIssues),
					"issue_reference": issueReference,
					"issue_number":    issue.Number,
					"issue_url":       strings.TrimSpace(issue.HTMLURL),
				},
				CreatedAt: issue.CreatedAt,
				UpdatedAt: issue.UpdatedAt,
			},
		}
		tasks = append(tasks, task)
	}
	if len(tasks) == 0 {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("github issues payload contained no actionable issues for %s/%s", owner, repository))
	}

	epicStatus := domaintracker.StatusInProgress
	allCompleted := true
	for _, task := range tasks {
		if task.Status != domaintracker.StatusCompleted {
			allCompleted = false
			break
		}
	}
	if allCompleted {
		epicStatus = domaintracker.StatusCompleted
	}

	board := domaintracker.Board{
		BoardID: boardID,
		RunID:   request.RunID,
		Title:   fmt.Sprintf("GitHub Issues %s/%s", owner, repository),
		Goal:    fmt.Sprintf("Execute tracker intake from github issues for %s/%s", owner, repository),
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: request.Source.Location,
			BoardID:  boardID,
			Config:   request.Source.Config,
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{
				ID:        domaintracker.WorkItemID("epic-github-issues"),
				BoardID:   boardID,
				Title:     "GitHub Issues Intake",
				Status:    epicStatus,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Tasks: tasks,
		}},
		Metadata: map[string]any{
			"project_id":  request.ProjectID,
			"workflow_id": request.WorkflowID,
			"owner":       owner,
			"repository":  repository,
			"issue_count": len(tasks),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	return board, nil
}

type githubIssue struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	State      string    `json:"state"`
	HTMLURL    string    `json:"html_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	PullRequest any      `json:"pull_request"`
	Labels     []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

func (provider *ExternalProvider) fetchGitHubIssues(ctx context.Context, owner, repository string) ([]githubIssue, error) {
	if provider == nil || provider.httpClient == nil {
		return nil, failures.WrapTerminal(fmt.Errorf("github issues provider is not configured"))
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues?state=all&per_page=100", provider.baseURL, owner, repository)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("build github issues request: %w", err))
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	if provider.tokenProvider != nil {
		token, tokenErr := provider.tokenProvider.AccessToken(ctx)
		if tokenErr != nil {
			return nil, failures.WrapTerminal(fmt.Errorf("load github auth token: %w", tokenErr))
		}
		if strings.TrimSpace(token) != "" {
			request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
		}
	}
	response, err := provider.httpClient.Do(request)
	if err != nil {
		return nil, failures.WrapTransient(fmt.Errorf("execute github issues request: %w", err))
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, failures.WrapTransient(fmt.Errorf("read github issues response: %w", err))
	}
	if response.StatusCode >= 400 {
		statusErr := fmt.Errorf("github issues request failed: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(payload)))
		if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
			return nil, failures.WrapTransient(statusErr)
		}
		return nil, failures.WrapTerminal(statusErr)
	}
	var issues []githubIssue
	if err := json.Unmarshal(payload, &issues); err != nil {
		return nil, failures.WrapTerminal(fmt.Errorf("decode github issues response: %w", err))
	}
	filtered := make([]githubIssue, 0, len(issues))
	for _, issue := range issues {
		if issue.PullRequest != nil {
			continue
		}
		filtered = append(filtered, issue)
	}
	return filtered, nil
}

func parseGitHubLocation(location string) (string, string, error) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(location), "/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", failures.WrapTerminal(fmt.Errorf("github_issues source location must be owner/repo: %q", location))
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func statusFromGitHubIssueState(state string) domaintracker.Status {
	if strings.EqualFold(strings.TrimSpace(state), "closed") {
		return domaintracker.StatusCompleted
	}
	return domaintracker.StatusNotStarted
}

func extractPriority(labels []struct{ Name string `json:"name"` }) domaintracker.Priority {
	for _, label := range labels {
		normalized := strings.ToLower(strings.TrimSpace(label.Name))
		switch normalized {
		case "p0", "priority:p0", "priority/critical", "priority-critical", "critical":
			return domaintracker.PriorityP0
		case "p1", "priority:p1", "priority/high", "priority-high", "high":
			return domaintracker.PriorityP1
		case "p2", "priority:p2", "priority/medium", "priority-medium", "medium":
			return domaintracker.PriorityP2
		case "p3", "priority:p3", "priority/low", "priority-low", "low":
			return domaintracker.PriorityP3
		}
	}
	return ""
}

func issueReferenceFromMetadata(metadata map[string]any) string {
	if metadata == nil {
		return ""
	}
	if value, ok := metadata["issue_reference"]; ok {
		switch reference := value.(type) {
		case string:
			return strings.TrimSpace(reference)
		case fmt.Stringer:
			return strings.TrimSpace(reference.String())
		case int:
			return strconv.Itoa(reference)
		}
	}
	return ""
}

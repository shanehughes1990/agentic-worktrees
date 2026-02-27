package scm

import (
	"agentic-orchestrator/internal/domain/failures"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type GitHubAdapterConfig struct {
	APIBaseURL string
	RepoPath   string
}

type GitHubAdapter struct {
	baseURL       string
	repoPath      string
	httpClient    *http.Client
	tokenProvider TokenProvider
	gitRunner     GitRunner
}

func NewGitHubAdapter(config GitHubAdapterConfig, httpClient *http.Client, tokenProvider TokenProvider, gitRunner GitRunner) (*GitHubAdapter, error) {
	baseURL := strings.TrimSpace(config.APIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("parse github api base url: %w", err)
	}
	if strings.TrimSpace(config.RepoPath) == "" {
		return nil, failures.WrapTerminal(fmt.Errorf("repo path is required"))
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if tokenProvider == nil {
		return nil, failures.WrapTerminal(fmt.Errorf("token provider is required"))
	}
	if gitRunner == nil {
		gitRunner = NewExecGitRunner()
	}
	return &GitHubAdapter{
		baseURL:       strings.TrimRight(baseURL, "/"),
		repoPath:      config.RepoPath,
		httpClient:    httpClient,
		tokenProvider: tokenProvider,
		gitRunner:     gitRunner,
	}, nil
}

func (adapter *GitHubAdapter) SourceState(ctx context.Context, repository domainscm.Repository) (domainscm.SourceState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.SourceState{}, err
	}

	var repositoryResponse struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, adapter.repoPathURL(repository, ""), nil, &repositoryResponse); err != nil {
		return domainscm.SourceState{}, err
	}

	var commitResponse struct {
		SHA string `json:"sha"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, adapter.repoPathURL(repository, path.Join("commits", repositoryResponse.DefaultBranch)), nil, &commitResponse); err != nil {
		return domainscm.SourceState{}, err
	}

	state := domainscm.SourceState{DefaultBranch: repositoryResponse.DefaultBranch, HeadSHA: commitResponse.SHA}
	if err := state.Validate(); err != nil {
		return domainscm.SourceState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) EnsureWorktree(ctx context.Context, repository domainscm.Repository, spec domainscm.WorktreeSpec) (domainscm.WorktreeState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	if err := spec.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	if _, err := adapter.gitRunner.Run(ctx, adapter.repoPath, "worktree", "add", "-B", spec.TargetBranch, spec.Path, spec.BaseBranch); err != nil {
		return domainscm.WorktreeState{}, failures.WrapTerminal(err)
	}
	headSHA, err := adapter.worktreeHeadSHA(ctx, spec.Path)
	if err != nil {
		return domainscm.WorktreeState{}, err
	}
	state := domainscm.WorktreeState{Path: spec.Path, Branch: spec.TargetBranch, Base: spec.BaseBranch, HeadSHA: headSHA, IsInSync: true, IsCleaned: false}
	if err := state.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) SyncWorktree(ctx context.Context, repository domainscm.Repository, worktreePath string) (domainscm.WorktreeState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	if strings.TrimSpace(worktreePath) == "" {
		return domainscm.WorktreeState{}, failures.WrapTerminal(fmt.Errorf("worktree path is required"))
	}
	branchName, err := adapter.gitRunner.Run(ctx, worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return domainscm.WorktreeState{}, failures.WrapTerminal(err)
	}
	if _, err := adapter.gitRunner.Run(ctx, worktreePath, "fetch", "origin", strings.TrimSpace(branchName)); err != nil {
		return domainscm.WorktreeState{}, failures.WrapTransient(err)
	}
	if _, err := adapter.gitRunner.Run(ctx, worktreePath, "reset", "--hard", "origin/"+strings.TrimSpace(branchName)); err != nil {
		return domainscm.WorktreeState{}, failures.WrapTransient(err)
	}
	headSHA, err := adapter.worktreeHeadSHA(ctx, worktreePath)
	if err != nil {
		return domainscm.WorktreeState{}, err
	}
	state := domainscm.WorktreeState{Path: worktreePath, Branch: strings.TrimSpace(branchName), Base: strings.TrimSpace(branchName), HeadSHA: headSHA, IsInSync: true, IsCleaned: false}
	if err := state.Validate(); err != nil {
		return domainscm.WorktreeState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) CleanupWorktree(ctx context.Context, repository domainscm.Repository, worktreePath string) error {
	if err := repository.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(worktreePath) == "" {
		return failures.WrapTerminal(fmt.Errorf("worktree path is required"))
	}
	if _, err := adapter.gitRunner.Run(ctx, adapter.repoPath, "worktree", "remove", "--force", worktreePath); err != nil {
		return failures.WrapTransient(err)
	}
	return nil
}

func (adapter *GitHubAdapter) EnsureBranch(ctx context.Context, repository domainscm.Repository, spec domainscm.BranchSpec) (domainscm.BranchState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	if err := spec.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}

	var refResponse struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, adapter.repoPathURL(repository, path.Join("git/ref/heads", spec.TargetBranch)), nil, &refResponse); err == nil {
		return domainscm.BranchState{Name: spec.TargetBranch, Base: spec.BaseBranch, HeadSHA: refResponse.Object.SHA, InSync: true}, nil
	}

	baseSHA, err := adapter.resolveBranchSHA(ctx, repository, spec.BaseBranch)
	if err != nil {
		return domainscm.BranchState{}, err
	}

	requestPayload := map[string]any{"ref": "refs/heads/" + spec.TargetBranch, "sha": baseSHA}
	if err := adapter.doJSON(ctx, http.MethodPost, adapter.repoPathURL(repository, "git/refs"), requestPayload, &refResponse); err != nil {
		return domainscm.BranchState{}, err
	}
	state := domainscm.BranchState{Name: spec.TargetBranch, Base: spec.BaseBranch, HeadSHA: refResponse.Object.SHA, InSync: true}
	if err := state.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) SyncBranch(ctx context.Context, repository domainscm.Repository, branchName string) (domainscm.BranchState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	if strings.TrimSpace(branchName) == "" {
		return domainscm.BranchState{}, failures.WrapTerminal(fmt.Errorf("branch_name is required"))
	}
	sha, err := adapter.resolveBranchSHA(ctx, repository, branchName)
	if err != nil {
		return domainscm.BranchState{}, err
	}
	state := domainscm.BranchState{Name: branchName, Base: branchName, HeadSHA: sha, InSync: true}
	if err := state.Validate(); err != nil {
		return domainscm.BranchState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) CreateOrUpdatePullRequest(ctx context.Context, spec domainscm.PullRequestSpec) (domainscm.PullRequestState, error) {
	if err := spec.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}

	pullRequestState, found, err := adapter.findOpenPullRequest(ctx, spec.Repository, spec.SourceBranch, spec.TargetBranch)
	if err != nil {
		return domainscm.PullRequestState{}, err
	}
	if found {
		return pullRequestState, nil
	}

	requestPayload := map[string]any{
		"title": spec.Title,
		"body":  spec.Body,
		"head":  spec.SourceBranch,
		"base":  spec.TargetBranch,
	}
	var response struct {
		Number int    `json:"number"`
		HTMLURL string `json:"html_url"`
		State string `json:"state"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := adapter.doJSON(ctx, http.MethodPost, adapter.repoPathURL(spec.Repository, "pulls"), requestPayload, &response); err != nil {
		return domainscm.PullRequestState{}, err
	}
	state := domainscm.PullRequestState{Number: response.Number, URL: response.HTMLURL, State: response.State, HeadSHA: response.Head.SHA}
	if err := state.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) GetPullRequest(ctx context.Context, repository domainscm.Repository, pullRequestNumber int) (domainscm.PullRequestState, error) {
	if err := repository.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	if pullRequestNumber <= 0 {
		return domainscm.PullRequestState{}, failures.WrapTerminal(fmt.Errorf("pull_request_number is required"))
	}
	var response struct {
		Number int    `json:"number"`
		HTMLURL string `json:"html_url"`
		State string `json:"state"`
		Merged bool `json:"merged"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, adapter.repoPathURL(repository, path.Join("pulls", strconv.Itoa(pullRequestNumber))), nil, &response); err != nil {
		return domainscm.PullRequestState{}, err
	}
	state := domainscm.PullRequestState{Number: response.Number, URL: response.HTMLURL, State: response.State, Merged: response.Merged, HeadSHA: response.Head.SHA}
	if err := state.Validate(); err != nil {
		return domainscm.PullRequestState{}, err
	}
	return state, nil
}

func (adapter *GitHubAdapter) SubmitReview(ctx context.Context, spec domainscm.ReviewSpec) (domainscm.ReviewDecision, error) {
	if err := spec.Validate(); err != nil {
		return "", err
	}
	event := "COMMENT"
	switch spec.Decision {
	case domainscm.ReviewDecisionApprove:
		event = "APPROVE"
	case domainscm.ReviewDecisionRequestChanges:
		event = "REQUEST_CHANGES"
	}
	requestPayload := map[string]any{"body": spec.Body, "event": event}
	if err := adapter.doJSON(ctx, http.MethodPost, adapter.repoPathURL(spec.Repository, path.Join("pulls", strconv.Itoa(spec.PullRequestNumber), "reviews")), requestPayload, nil); err != nil {
		return "", err
	}
	return spec.Decision, nil
}

func (adapter *GitHubAdapter) CheckMergeReadiness(ctx context.Context, repository domainscm.Repository, pullRequestNumber int) (domainscm.MergeReadiness, error) {
	state, err := adapter.GetPullRequest(ctx, repository, pullRequestNumber)
	if err != nil {
		return domainscm.MergeReadiness{}, err
	}
	if strings.EqualFold(state.State, "open") {
		return domainscm.MergeReadiness{CanMerge: true}, nil
	}
	return domainscm.MergeReadiness{CanMerge: false, Reason: "pull request is not open"}, nil
}

func (adapter *GitHubAdapter) findOpenPullRequest(ctx context.Context, repository domainscm.Repository, sourceBranch, targetBranch string) (domainscm.PullRequestState, bool, error) {
	query := url.Values{}
	query.Set("state", "open")
	query.Set("head", repository.Owner+":"+sourceBranch)
	query.Set("base", targetBranch)

	endpoint := adapter.repoPathURL(repository, "pulls") + "?" + query.Encode()
	var response []struct {
		Number int `json:"number"`
		HTMLURL string `json:"html_url"`
		State string `json:"state"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return domainscm.PullRequestState{}, false, err
	}
	if len(response) == 0 {
		return domainscm.PullRequestState{}, false, nil
	}
	state := domainscm.PullRequestState{Number: response[0].Number, URL: response[0].HTMLURL, State: response[0].State, HeadSHA: response[0].Head.SHA}
	if err := state.Validate(); err != nil {
		return domainscm.PullRequestState{}, false, err
	}
	return state, true, nil
}

func (adapter *GitHubAdapter) resolveBranchSHA(ctx context.Context, repository domainscm.Repository, branchName string) (string, error) {
	var response struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := adapter.doJSON(ctx, http.MethodGet, adapter.repoPathURL(repository, path.Join("branches", branchName)), nil, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.Commit.SHA) == "" {
		return "", failures.WrapTerminal(fmt.Errorf("github response missing commit sha for branch %q", branchName))
	}
	return strings.TrimSpace(response.Commit.SHA), nil
}

func (adapter *GitHubAdapter) worktreeHeadSHA(ctx context.Context, worktreePath string) (string, error) {
	sha, err := adapter.gitRunner.Run(ctx, worktreePath, "rev-parse", "HEAD")
	if err != nil {
		return "", failures.WrapTransient(err)
	}
	if strings.TrimSpace(sha) == "" {
		return "", failures.WrapTerminal(fmt.Errorf("empty worktree head sha for %q", worktreePath))
	}
	return strings.TrimSpace(sha), nil
}

func (adapter *GitHubAdapter) repoPathURL(repository domainscm.Repository, suffix string) string {
	cleanSuffix := strings.TrimPrefix(strings.TrimSpace(suffix), "/")
	if cleanSuffix == "" {
		return fmt.Sprintf("%s/repos/%s/%s", adapter.baseURL, repository.Owner, repository.Name)
	}
	return fmt.Sprintf("%s/repos/%s/%s/%s", adapter.baseURL, repository.Owner, repository.Name, cleanSuffix)
}

func (adapter *GitHubAdapter) doJSON(ctx context.Context, method, endpoint string, body any, target any) error {
	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			return failures.WrapTerminal(fmt.Errorf("encode github request body: %w", err))
		}
		requestBody = bytes.NewReader(encodedBody)
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return failures.WrapTerminal(fmt.Errorf("build github request: %w", err))
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Content-Type", "application/json")
	token, tokenErr := adapter.tokenProvider.AccessToken(ctx)
	if tokenErr != nil {
		return failures.WrapTerminal(fmt.Errorf("load github auth token: %w", tokenErr))
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := adapter.httpClient.Do(request)
	if err != nil {
		return failures.WrapTransient(fmt.Errorf("execute github request: %w", err))
	}
	defer response.Body.Close()

	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return failures.WrapTransient(fmt.Errorf("read github response body: %w", readErr))
	}

	if response.StatusCode >= 400 {
		statusErr := fmt.Errorf("github request failed (%s %s): status=%d body=%s", method, endpoint, response.StatusCode, strings.TrimSpace(string(responseBody)))
		if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
			return failures.WrapTransient(statusErr)
		}
		return failures.WrapTerminal(statusErr)
	}

	if target == nil || len(responseBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(responseBody, target); err != nil {
		return failures.WrapTerminal(fmt.Errorf("decode github response: %w", err))
	}
	return nil
}

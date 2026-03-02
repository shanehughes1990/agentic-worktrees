package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeTrackerService struct {
	request applicationtracker.SyncBoardRequest
	called  bool
	err     error
	board   domaintracker.Board
}

func (service *fakeTrackerService) SyncBoard(ctx context.Context, request applicationtracker.SyncBoardRequest) (domaintracker.Board, error) {
	_ = ctx
	service.request = request
	service.called = true
	if service.err != nil {
		return domaintracker.Board{}, service.err
	}
	return service.board, nil
}

type fakeSupervisorSignalService struct {
	issueOpenedCalls []struct {
		source         string
		issueReference string
	}
	trackerAttentionCalls int
}

func (service *fakeSupervisorSignalService) OnExecution(ctx context.Context, record taskengine.ExecutionRecord, attempt int, maxRetry int) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = record
	_ = attempt
	_ = maxRetry
	return domainsupervisor.Decision{}, nil
}

func (service *fakeSupervisorSignalService) OnCheckpointSaved(ctx context.Context, correlation taskengine.CorrelationIDs, jobKind taskengine.JobKind, idempotencyKey string, step string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	_ = jobKind
	_ = idempotencyKey
	_ = step
	return domainsupervisor.Decision{}, nil
}

func (service *fakeSupervisorSignalService) OnTrackerAttention(ctx context.Context, correlation taskengine.CorrelationIDs, reason string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	_ = reason
	service.trackerAttentionCalls++
	return domainsupervisor.Decision{}, nil
}

func (service *fakeSupervisorSignalService) OnIssueOpened(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	service.issueOpenedCalls = append(service.issueOpenedCalls, struct {
		source         string
		issueReference string
	}{source: source, issueReference: issueReference})
	return domainsupervisor.Decision{}, nil
}

func (service *fakeSupervisorSignalService) OnIssueApproved(ctx context.Context, correlation taskengine.CorrelationIDs, source string, issueReference string, approvedBy string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	_ = source
	_ = issueReference
	_ = approvedBy
	return domainsupervisor.Decision{}, nil
}

func TestIngestionAgentHandlerDispatchesTrackerSync(t *testing.T) {
	service := &fakeTrackerService{}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, err := json.Marshal(IngestionAgentPayload{
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
		Prompt:         "ingest tracker board",
		ProjectID:      "project-1",
		WorkflowID:     "workflow-1",
		BoardSources: []IngestionBoardSourcePayload{{
			BoardID:                  "board-1",
			Kind:                     "internal",
			Location:                 "board-1.json",
			AppliesToAllRepositories: true,
		}},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !service.called {
		t.Fatalf("expected tracker service to be called")
	}
	if service.request.ProjectID != "project-1" || service.request.WorkflowID != "workflow-1" {
		t.Fatalf("expected project/workflow boundary propagation, got %+v", service.request)
	}
	if service.request.Source.Kind != domaintracker.SourceKindInternal {
		t.Fatalf("expected source kind internal, got %q", service.request.Source.Kind)
	}
	if service.request.Source.Location != "board-1.json" {
		t.Fatalf("expected unmodified location, got %q", service.request.Source.Location)
	}
	if service.request.Source.BoardID != "board-1" {
		t.Fatalf("expected board id fallback to payload board_id, got %q", service.request.Source.BoardID)
	}
}

func TestIngestionAgentHandlerReturnsServiceError(t *testing.T) {
	service := &fakeTrackerService{err: errors.New("boom")}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, _ := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSources: []IngestionBoardSourcePayload{{
			BoardID:                  "board-1",
			Kind:                     "internal",
			Location:                 "board-1.json",
			AppliesToAllRepositories: true,
		}},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err == nil {
		t.Fatalf("expected service error")
	}
}

func TestIngestionAgentHandlerRejectsMultipleBoardSources(t *testing.T) {
	service := &fakeTrackerService{}
	handler, err := NewIngestionAgentHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	payload, _ := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		TaskID:     "task-1",
		JobID:      "job-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSources: []IngestionBoardSourcePayload{
			{BoardID: "board-1", Kind: "internal", Location: "board-1.json", AppliesToAllRepositories: true},
			{BoardID: "board-2", Kind: "internal", Location: "board-2.json", AppliesToAllRepositories: true},
		},
	})

	err = handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload})
	if err == nil {
		t.Fatalf("expected validation error for multiple board sources")
	}
	if service.called {
		t.Fatalf("did not expect tracker service call")
	}
}

func TestIngestionAgentHandlerEmitsIssueOpenedPerGitHubIssueTask(t *testing.T) {
	supervisorService := &fakeSupervisorSignalService{}
	trackerService := &fakeTrackerService{board: domaintracker.Board{
		BoardID: "github:octo/repo",
		RunID:   "run-1",
		Source: domaintracker.SourceRef{Kind: domaintracker.SourceKindGitHubIssues, Location: "octo/repo"},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "github:octo/repo", Title: "GitHub", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{
				{WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "github:octo/repo", Title: "Issue 1", Status: domaintracker.StatusNotStarted, Metadata: map[string]any{"issue_reference": "octo/repo#1"}}},
				{WorkItem: domaintracker.WorkItem{ID: "task-2", BoardID: "github:octo/repo", Title: "Issue 2", Status: domaintracker.StatusNotStarted, Metadata: map[string]any{"issue_reference": "octo/repo#2"}}},
			},
		}},
	}}
	handler, err := NewIngestionAgentHandlerWithSupervisor(trackerService, supervisorService)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, err := json.Marshal(IngestionAgentPayload{
		RunID:      "run-1",
		TaskID:     "task-1",
		JobID:      "job-1",
		Prompt:     "ingest tracker board",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		BoardSources: []IngestionBoardSourcePayload{{
			BoardID:                  "board-1",
			Kind:                     "github_issues",
			Location:                 "octo/repo",
			AppliesToAllRepositories: true,
		}},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindIngestionAgent, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(supervisorService.issueOpenedCalls) != 2 {
		t.Fatalf("expected two issue-opened calls, got %d", len(supervisorService.issueOpenedCalls))
	}
	if supervisorService.issueOpenedCalls[0].source != "octo/repo" {
		t.Fatalf("expected source octo/repo, got %q", supervisorService.issueOpenedCalls[0].source)
	}
	if supervisorService.issueOpenedCalls[0].issueReference != "octo/repo#1" {
		t.Fatalf("expected first issue reference octo/repo#1, got %q", supervisorService.issueOpenedCalls[0].issueReference)
	}
}

func (service *fakeSupervisorSignalService) OnPRChecksEvaluated(ctx context.Context, correlation taskengine.CorrelationIDs, provider string, owner string, repository string, pullRequestNumber int, canMerge bool, reason string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	_ = provider
	_ = owner
	_ = repository
	_ = pullRequestNumber
	_ = canMerge
	_ = reason
	return domainsupervisor.Decision{}, nil
}

func (service *fakeSupervisorSignalService) OnPRMergeRequested(ctx context.Context, correlation taskengine.CorrelationIDs, provider string, owner string, repository string, pullRequestNumber int, mergeMethod string) (domainsupervisor.Decision, error) {
	_ = ctx
	_ = correlation
	_ = provider
	_ = owner
	_ = repository
	_ = pullRequestNumber
	_ = mergeMethod
	return domainsupervisor.Decision{}, nil
}

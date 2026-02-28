package worker

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainagent "agentic-orchestrator/internal/domain/agent"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakeAgentService struct {
	request domainagent.ExecutionRequest
	called  bool
	err     error
}

func (service *fakeAgentService) Execute(ctx context.Context, request domainagent.ExecutionRequest) error {
	_ = ctx
	service.request = request
	service.called = true
	return service.err
}

func TestAgentWorkflowHandlerDispatchesExecute(t *testing.T) {
	service := &fakeAgentService{}
	handler, err := NewAgentWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
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

	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !service.called {
		t.Fatalf("expected execute to be called")
	}
	if service.request.Session.SessionID != "session-1" {
		t.Fatalf("expected session id session-1, got %q", service.request.Session.SessionID)
	}
	if service.request.Session.Repository.Name != "repo" {
		t.Fatalf("expected repository repo, got %q", service.request.Session.Repository.Name)
	}
	if service.request.Metadata.CorrelationIDs.RunID != "run-1" {
		t.Fatalf("expected run correlation run-1, got %q", service.request.Metadata.CorrelationIDs.RunID)
	}
}

func TestAgentWorkflowHandlerReturnsServiceError(t *testing.T) {
	service := &fakeAgentService{err: errors.New("boom")}
	handler, err := NewAgentWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(AgentWorkflowPayload{
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
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err == nil {
		t.Fatalf("expected service error")
	}
}

func TestAgentWorkflowHandlerPropagatesResumeCheckpoint(t *testing.T) {
	service := &fakeAgentService{}
	handler, err := NewAgentWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(AgentWorkflowPayload{
		SessionID:             "session-1",
		Prompt:                "run analysis",
		Provider:              "github",
		Owner:                 "acme",
		Repository:            "repo",
		RunID:                 "run-1",
		TaskID:                "task-1",
		JobID:                 "job-1",
		IdempotencyKey:        "id-1",
		ResumeCheckpointStep:  "source_state",
		ResumeCheckpointToken: "id-1",
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.request.ResumeCheckpoint == nil {
		t.Fatalf("expected resume checkpoint to be populated")
	}
	if service.request.ResumeCheckpoint.Step != "source_state" {
		t.Fatalf("expected checkpoint step source_state, got %q", service.request.ResumeCheckpoint.Step)
	}
	if service.request.ResumeCheckpoint.Token != "id-1" {
		t.Fatalf("expected checkpoint token id-1, got %q", service.request.ResumeCheckpoint.Token)
	}
}

func TestAgentWorkflowHandlerPropagatesResumeCheckpointObject(t *testing.T) {
	service := &fakeAgentService{}
	handler, err := NewAgentWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(AgentWorkflowPayload{
		SessionID:      "session-1",
		Prompt:         "run analysis",
		Provider:       "github",
		Owner:          "acme",
		Repository:     "repo",
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
		ResumeCheckpoint: &taskengine.Checkpoint{
			Step:  "source_state",
			Token: "id-1",
		},
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.request.ResumeCheckpoint == nil {
		t.Fatalf("expected resume checkpoint to be populated")
	}
	if service.request.ResumeCheckpoint.Step != "source_state" {
		t.Fatalf("expected checkpoint step source_state, got %q", service.request.ResumeCheckpoint.Step)
	}
	if service.request.ResumeCheckpoint.Token != "id-1" {
		t.Fatalf("expected checkpoint token id-1, got %q", service.request.ResumeCheckpoint.Token)
	}
}

func TestAgentWorkflowHandlerTrimsLegacyResumeCheckpointFields(t *testing.T) {
	service := &fakeAgentService{}
	handler, err := NewAgentWorkflowHandler(service)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	payload, _ := json.Marshal(AgentWorkflowPayload{
		SessionID:             "session-1",
		Prompt:                "run analysis",
		Provider:              "github",
		Owner:                 "acme",
		Repository:            "repo",
		RunID:                 "run-1",
		TaskID:                "task-1",
		JobID:                 "job-1",
		IdempotencyKey:        "id-1",
		ResumeCheckpointStep:  " source_state ",
		ResumeCheckpointToken: " id-1 ",
	})
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.request.ResumeCheckpoint == nil {
		t.Fatalf("expected resume checkpoint to be populated")
	}
	if service.request.ResumeCheckpoint.Step != "source_state" || service.request.ResumeCheckpoint.Token != "id-1" {
		t.Fatalf("expected trimmed checkpoint values, got %+v", service.request.ResumeCheckpoint)
	}
}

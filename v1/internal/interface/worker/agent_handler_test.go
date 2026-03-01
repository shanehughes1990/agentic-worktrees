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

type fakeCheckpointStore struct {
	loadedCheckpoint *taskengine.Checkpoint
	loadErr          error
	savedCheckpoint  *taskengine.Checkpoint
	savedKey         string
	saveErr          error
}

func (store *fakeCheckpointStore) Save(ctx context.Context, idempotencyKey string, checkpoint taskengine.Checkpoint) error {
	_ = ctx
	store.savedKey = idempotencyKey
	store.savedCheckpoint = &checkpoint
	return store.saveErr
}

func (store *fakeCheckpointStore) Load(ctx context.Context, idempotencyKey string) (*taskengine.Checkpoint, error) {
	_ = ctx
	_ = idempotencyKey
	if store.loadErr != nil {
		return nil, store.loadErr
	}
	return store.loadedCheckpoint, nil
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

func TestAgentWorkflowHandlerLoadsPersistedCheckpointWhenAvailable(t *testing.T) {
	service := &fakeAgentService{}
	store := &fakeCheckpointStore{loadedCheckpoint: &taskengine.Checkpoint{Step: "source_state", Token: "id-1"}}
	handler, err := NewAgentWorkflowHandlerWithCheckpointStore(service, store)
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
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if service.request.ResumeCheckpoint == nil {
		t.Fatalf("expected checkpoint loaded from store")
	}
	if service.request.ResumeCheckpoint.Step != "source_state" || service.request.ResumeCheckpoint.Token != "id-1" {
		t.Fatalf("expected persisted checkpoint, got %+v", service.request.ResumeCheckpoint)
	}
}

func TestAgentWorkflowHandlerPersistsCheckpointAfterSuccess(t *testing.T) {
	service := &fakeAgentService{}
	store := &fakeCheckpointStore{}
	handler, err := NewAgentWorkflowHandlerWithCheckpointStore(service, store)
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
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.savedCheckpoint == nil {
		t.Fatalf("expected checkpoint to be persisted")
	}
	if store.savedKey != "id-1" {
		t.Fatalf("expected save key id-1, got %q", store.savedKey)
	}
	if store.savedCheckpoint.Step != "source_state" || store.savedCheckpoint.Token != "id-1" {
		t.Fatalf("expected source_state checkpoint, got %+v", store.savedCheckpoint)
	}
}

func TestAgentWorkflowHandlerDoesNotPersistCheckpointOnServiceError(t *testing.T) {
	service := &fakeAgentService{err: errors.New("boom")}
	store := &fakeCheckpointStore{}
	handler, err := NewAgentWorkflowHandlerWithCheckpointStore(service, store)
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
	if store.savedCheckpoint != nil {
		t.Fatalf("expected no persisted checkpoint on error")
	}
}


type fakeExecutionJournal struct {
	records []taskengine.ExecutionRecord
	err     error
}

func (journal *fakeExecutionJournal) Upsert(ctx context.Context, record taskengine.ExecutionRecord) error {
	_ = ctx
	if journal.err != nil {
		return journal.err
	}
	journal.records = append(journal.records, record)
	return nil
}

func (journal *fakeExecutionJournal) Load(ctx context.Context, runID string, taskID string, jobID string) (*taskengine.ExecutionRecord, error) {
	_ = ctx
	_ = runID
	_ = taskID
	_ = jobID
	return nil, nil
}

func TestAgentWorkflowHandlerRecordsExecutionJournalSuccess(t *testing.T) {
	service := &fakeAgentService{}
	journal := &fakeExecutionJournal{}
	handler, err := NewAgentWorkflowHandlerWithReliability(service, nil, journal)
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
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(journal.records) < 2 {
		t.Fatalf("expected at least 2 journal records, got %d", len(journal.records))
	}
	if journal.records[0].Status != taskengine.ExecutionStatusRunning {
		t.Fatalf("expected first status running, got %q", journal.records[0].Status)
	}
	if journal.records[len(journal.records)-1].Status != taskengine.ExecutionStatusSucceeded {
		t.Fatalf("expected last status succeeded, got %q", journal.records[len(journal.records)-1].Status)
	}
}

func TestAgentWorkflowHandlerRecordsExecutionJournalFailure(t *testing.T) {
	service := &fakeAgentService{err: errors.New("boom")}
	journal := &fakeExecutionJournal{}
	handler, err := NewAgentWorkflowHandlerWithReliability(service, nil, journal)
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
		t.Fatalf("expected error")
	}
	if len(journal.records) < 2 {
		t.Fatalf("expected at least 2 journal records, got %d", len(journal.records))
	}
	if journal.records[len(journal.records)-1].Status != taskengine.ExecutionStatusFailed {
		t.Fatalf("expected last status failed, got %q", journal.records[len(journal.records)-1].Status)
	}
}

func TestAgentWorkflowHandlerIgnoresExecutionJournalWriteErrors(t *testing.T) {
	service := &fakeAgentService{}
	journal := &fakeExecutionJournal{err: errors.New("journal unavailable")}
	handler, err := NewAgentWorkflowHandlerWithReliability(service, nil, journal)
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
	if err := handler.Handle(context.Background(), taskengine.Job{Kind: taskengine.JobKindAgentWorkflow, Payload: payload}); err != nil {
		t.Fatalf("expected journal errors to be ignored, got: %v", err)
	}
}

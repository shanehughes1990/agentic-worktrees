package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	"agentic-orchestrator/internal/application/taskengine"
	domainagent "agentic-orchestrator/internal/domain/agent"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AgentWorkflowPayload struct {
	SessionID             string                 `json:"session_id"`
	Prompt                string                 `json:"prompt"`
	Provider              string                 `json:"provider"`
	Owner                 string                 `json:"owner"`
	Repository            string                 `json:"repository"`
	RunID                 string                 `json:"run_id"`
	TaskID                string                 `json:"task_id"`
	JobID                 string                 `json:"job_id"`
	IdempotencyKey        string                 `json:"idempotency_key"`
	ResumeCheckpoint      *taskengine.Checkpoint `json:"resume_checkpoint,omitempty"`
	ResumeCheckpointStep  string                 `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string                 `json:"resume_checkpoint_token,omitempty"`
}

type agentService interface {
	Execute(ctx context.Context, request domainagent.ExecutionRequest) error
}

var _ agentService = (*applicationagent.Service)(nil)

type AgentWorkflowHandler struct {
	service          agentService
	checkpointStore  taskengine.CheckpointStore
	executionJournal taskengine.ExecutionJournal
}

func NewAgentWorkflowHandler(service agentService) (*AgentWorkflowHandler, error) {
	return NewAgentWorkflowHandlerWithCheckpointStore(service, nil)
}

func NewAgentWorkflowHandlerWithCheckpointStore(service agentService, checkpointStore taskengine.CheckpointStore) (*AgentWorkflowHandler, error) {
	return NewAgentWorkflowHandlerWithReliability(service, checkpointStore, nil)
}

func NewAgentWorkflowHandlerWithReliability(service agentService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*AgentWorkflowHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("agent service is required")
	}
	return &AgentWorkflowHandler{service: service, checkpointStore: checkpointStore, executionJournal: executionJournal}, nil
}

func (handler *AgentWorkflowHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload AgentWorkflowPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode agent workflow payload: %w", err)
	}
	request := domainagent.ExecutionRequest{
		Session: domainagent.SessionRef{
			SessionID: payload.SessionID,
			Repository: domainscm.Repository{
				Provider: payload.Provider,
				Owner:    payload.Owner,
				Name:     payload.Repository,
			},
		},
		Prompt: payload.Prompt,
		Metadata: domainagent.Metadata{
			CorrelationIDs: domainagent.CorrelationIDs{
				RunID:  payload.RunID,
				TaskID: payload.TaskID,
				JobID:  payload.JobID,
			},
			IdempotencyKey: payload.IdempotencyKey,
		},
	}

	step := "source_state"
	if err := handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusRunning, ""); err != nil {
		return err
	}

	retryCheckpoint := (taskengine.RetryCheckpointContract{
		ResumeCheckpoint:      payload.ResumeCheckpoint,
		ResumeCheckpointStep:  payload.ResumeCheckpointStep,
		ResumeCheckpointToken: payload.ResumeCheckpointToken,
	}).Checkpoint()

	idempotencyKey := strings.TrimSpace(payload.IdempotencyKey)
	resumeCheckpoint := retryCheckpoint
	if handler.checkpointStore != nil && idempotencyKey != "" {
		persistedCheckpoint, err := handler.checkpointStore.Load(ctx, idempotencyKey)
		if err != nil {
			_ = handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("load persisted checkpoint: %w", err)
		}
		if persistedCheckpoint != nil {
			resumeCheckpoint = persistedCheckpoint
		}
	}
	if resumeCheckpoint != nil {
		request.ResumeCheckpoint = &domainagent.Checkpoint{Step: resumeCheckpoint.Step, Token: resumeCheckpoint.Token}
	}

	if err := handler.service.Execute(ctx, request); err != nil {
		_ = handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
		return err
	}

	if handler.checkpointStore != nil && idempotencyKey != "" {
		if err := handler.checkpointStore.Save(ctx, idempotencyKey, taskengine.Checkpoint{Step: step, Token: idempotencyKey}); err != nil {
			_ = handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("persist completed checkpoint: %w", err)
		}
	}
	if err := handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusSucceeded, ""); err != nil {
		return err
	}
	return nil
}

func (handler *AgentWorkflowHandler) recordExecution(ctx context.Context, request domainagent.ExecutionRequest, kind taskengine.JobKind, step string, status taskengine.ExecutionStatus, errorMessage string) error {
	if handler == nil || handler.executionJournal == nil {
		return nil
	}
	record := taskengine.ExecutionRecord{
		RunID:          request.Metadata.CorrelationIDs.RunID,
		TaskID:         request.Metadata.CorrelationIDs.TaskID,
		JobID:          request.Metadata.CorrelationIDs.JobID,
		JobKind:        kind,
		IdempotencyKey: request.Metadata.IdempotencyKey,
		Step:           step,
		Status:         status,
		ErrorMessage:   strings.TrimSpace(errorMessage),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := handler.executionJournal.Upsert(ctx, record); err != nil {
		return fmt.Errorf("record execution journal: %w", err)
	}
	return nil
}

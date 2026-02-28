package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	"agentic-orchestrator/internal/application/taskengine"
	domainagent "agentic-orchestrator/internal/domain/agent"
	domainscm "agentic-orchestrator/internal/domain/scm"
	"context"
	"encoding/json"
	"fmt"
)

type AgentWorkflowPayload struct {
	SessionID             string `json:"session_id"`
	Prompt                string `json:"prompt"`
	Provider              string `json:"provider"`
	Owner                 string `json:"owner"`
	Repository            string `json:"repository"`
	RunID                 string `json:"run_id"`
	TaskID                string `json:"task_id"`
	JobID                 string `json:"job_id"`
	IdempotencyKey        string `json:"idempotency_key"`
	ResumeCheckpointStep  string `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string `json:"resume_checkpoint_token,omitempty"`
}

type agentService interface {
	Execute(ctx context.Context, request domainagent.ExecutionRequest) error
}

var _ agentService = (*applicationagent.Service)(nil)

type AgentWorkflowHandler struct {
	service agentService
}

func NewAgentWorkflowHandler(service agentService) (*AgentWorkflowHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("agent service is required")
	}
	return &AgentWorkflowHandler{service: service}, nil
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
	if payload.ResumeCheckpointStep != "" && payload.ResumeCheckpointToken != "" {
		request.ResumeCheckpoint = &domainagent.Checkpoint{
			Step:  payload.ResumeCheckpointStep,
			Token: payload.ResumeCheckpointToken,
		}
	}
	return handler.service.Execute(ctx, request)
}

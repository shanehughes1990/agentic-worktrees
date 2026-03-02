package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
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
	ProjectID             string                 `json:"project_id"`
	IdempotencyKey        string                 `json:"idempotency_key"`
	ResumeCheckpoint      *taskengine.Checkpoint `json:"resume_checkpoint,omitempty"`
	ResumeCheckpointStep  string                 `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string                 `json:"resume_checkpoint_token,omitempty"`
}

type AgentRuntimeService interface {
	Execute(ctx context.Context, request domainagent.ExecutionRequest) error
}

var _ AgentRuntimeService = (*applicationagent.Service)(nil)

type agentServiceFactoryFunc func(ctx context.Context, projectID string, scm applicationcontrolplane.ProjectSCM, repository applicationcontrolplane.ProjectRepository) (AgentRuntimeService, error)

type agentRuntimeResolver interface {
	Resolve(ctx context.Context, payload AgentWorkflowPayload) (AgentRuntimeService, domainscm.Repository, error)
}

type staticAgentRuntimeResolver struct {
	service AgentRuntimeService
}

func (resolver *staticAgentRuntimeResolver) Resolve(ctx context.Context, payload AgentWorkflowPayload) (AgentRuntimeService, domainscm.Repository, error) {
	_ = ctx
	if resolver == nil || resolver.service == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("agent service is required")
	}
	return resolver.service, domainscm.Repository{Provider: payload.Provider, Owner: payload.Owner, Name: payload.Repository}, nil
}

type projectSetupAgentRuntimeResolver struct {
	projectRepository projectSetupLookup
	serviceFactory    agentServiceFactoryFunc
}

func (resolver *projectSetupAgentRuntimeResolver) Resolve(ctx context.Context, payload AgentWorkflowPayload) (AgentRuntimeService, domainscm.Repository, error) {
	if resolver == nil || resolver.projectRepository == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("project setup repository is required")
	}
	if resolver.serviceFactory == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("agent service factory is required")
	}
	projectID := strings.TrimSpace(payload.ProjectID)
	if projectID == "" {
		return nil, domainscm.Repository{}, fmt.Errorf("project_id is required for agent workflow execution")
	}
	setup, err := resolver.projectRepository.GetProjectSetup(ctx, projectID)
	if err != nil {
		return nil, domainscm.Repository{}, fmt.Errorf("load project setup: %w", err)
	}
	if setup == nil {
		return nil, domainscm.Repository{}, fmt.Errorf("project setup not found for project_id %q", projectID)
	}
	repositoryConfig, err := primaryProjectRepository(setup.Repositories)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	scmConfig, err := projectSCMByID(setup.SCMs, repositoryConfig.SCMID)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	owner, repositoryName, err := ownerRepositoryFromURL(repositoryConfig.RepositoryURL)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	service, err := resolver.serviceFactory(ctx, projectID, scmConfig, repositoryConfig)
	if err != nil {
		return nil, domainscm.Repository{}, err
	}
	return service, domainscm.Repository{Provider: scmConfig.SCMProvider, Owner: owner, Name: repositoryName}, nil
}

type AgentWorkflowHandler struct {
	runtimeResolver   agentRuntimeResolver
	checkpointStore   taskengine.CheckpointStore
	executionJournal  taskengine.ExecutionJournal
	supervisorService supervisorSignalService
}

func NewAgentWorkflowHandler(service AgentRuntimeService) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandler(&staticAgentRuntimeResolver{service: service}, nil, nil, nil)
}

func NewAgentWorkflowHandlerWithCheckpointStore(service AgentRuntimeService, checkpointStore taskengine.CheckpointStore) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandler(&staticAgentRuntimeResolver{service: service}, checkpointStore, nil, nil)
}

func NewAgentWorkflowHandlerWithReliability(service AgentRuntimeService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandler(&staticAgentRuntimeResolver{service: service}, checkpointStore, executionJournal, nil)
}

func NewAgentWorkflowHandlerWithSupervisor(service AgentRuntimeService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandler(&staticAgentRuntimeResolver{service: service}, checkpointStore, executionJournal, supervisorService)
}

func NewAgentWorkflowHandlerWithProjectSetup(projectRepository projectSetupLookup, serviceFactory agentServiceFactoryFunc, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandler(&projectSetupAgentRuntimeResolver{projectRepository: projectRepository, serviceFactory: serviceFactory}, checkpointStore, executionJournal, supervisorService)
}

func newAgentWorkflowHandler(runtimeResolver agentRuntimeResolver, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	if runtimeResolver == nil {
		return nil, fmt.Errorf("agent runtime resolver is required")
	}
	return &AgentWorkflowHandler{runtimeResolver: runtimeResolver, checkpointStore: checkpointStore, executionJournal: executionJournal, supervisorService: supervisorService}, nil
}

func (handler *AgentWorkflowHandler) Handle(ctx context.Context, job taskengine.Job) error {
	var payload AgentWorkflowPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode agent workflow payload: %w", err)
	}
	idempotencyKey := strings.TrimSpace(payload.IdempotencyKey)
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
				RunID:     payload.RunID,
				TaskID:    payload.TaskID,
				JobID:     payload.JobID,
				ProjectID: payload.ProjectID,
			},
			IdempotencyKey: idempotencyKey,
		},
	}

	service, repository, err := handler.runtimeResolver.Resolve(ctx, payload)
	if err != nil {
		handler.safeRecordExecution(ctx, request, job.Kind, "source_state", taskengine.ExecutionStatusFailed, err.Error())
		return err
	}
	request.Session.Repository = repository

	step := "source_state"
	if record, err := handler.recordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusRunning, ""); err == nil {
		handler.safeSupervisorExecution(ctx, record)
	}

	retryCheckpoint := (taskengine.RetryCheckpointContract{
		ResumeCheckpoint:      payload.ResumeCheckpoint,
		ResumeCheckpointStep:  payload.ResumeCheckpointStep,
		ResumeCheckpointToken: payload.ResumeCheckpointToken,
	}).Checkpoint()

	resumeCheckpoint := retryCheckpoint
	if handler.checkpointStore != nil && idempotencyKey != "" {
		persistedCheckpoint, err := handler.checkpointStore.Load(ctx, idempotencyKey)
		if err != nil {
			handler.safeRecordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("load persisted checkpoint: %w", err)
		}
		if persistedCheckpoint != nil {
			resumeCheckpoint = persistedCheckpoint
		}
	}
	if resumeCheckpoint != nil {
		request.ResumeCheckpoint = &domainagent.Checkpoint{Step: resumeCheckpoint.Step, Token: resumeCheckpoint.Token}
	}

	if err := service.Execute(ctx, request); err != nil {
		handler.safeRecordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
		return err
	}

	if handler.checkpointStore != nil && idempotencyKey != "" {
		if err := handler.checkpointStore.Save(ctx, idempotencyKey, taskengine.Checkpoint{Step: step, Token: idempotencyKey}); err != nil {
			handler.safeRecordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, err.Error())
			return fmt.Errorf("persist completed checkpoint: %w", err)
		}
		handler.safeSupervisorCheckpoint(ctx, request.Metadata.CorrelationIDs, job.Kind, idempotencyKey, step)
	}
	handler.safeRecordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusSucceeded, "")
	return nil
}

func (handler *AgentWorkflowHandler) safeRecordExecution(ctx context.Context, request domainagent.ExecutionRequest, kind taskengine.JobKind, step string, status taskengine.ExecutionStatus, errorMessage string) {
	record, err := handler.recordExecution(ctx, request, kind, step, status, errorMessage)
	if err != nil {
		return
	}
	handler.safeSupervisorExecution(ctx, record)
}

func (handler *AgentWorkflowHandler) recordExecution(ctx context.Context, request domainagent.ExecutionRequest, kind taskengine.JobKind, step string, status taskengine.ExecutionStatus, errorMessage string) (taskengine.ExecutionRecord, error) {
	record := taskengine.ExecutionRecord{
		RunID:          request.Metadata.CorrelationIDs.RunID,
		TaskID:         request.Metadata.CorrelationIDs.TaskID,
		JobID:          request.Metadata.CorrelationIDs.JobID,
		ProjectID:      request.Metadata.CorrelationIDs.ProjectID,
		JobKind:        kind,
		IdempotencyKey: request.Metadata.IdempotencyKey,
		Step:           step,
		Status:         status,
		ErrorMessage:   strings.TrimSpace(errorMessage),
		UpdatedAt:      time.Now().UTC(),
	}
	if handler == nil || handler.executionJournal == nil {
		return record, nil
	}
	if err := handler.executionJournal.Upsert(ctx, record); err != nil {
		return taskengine.ExecutionRecord{}, fmt.Errorf("record execution journal: %w", err)
	}
	return record, nil
}

func (handler *AgentWorkflowHandler) safeSupervisorExecution(ctx context.Context, record taskengine.ExecutionRecord) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnExecution(ctx, record, 0, 0)
}

func (handler *AgentWorkflowHandler) safeSupervisorCheckpoint(ctx context.Context, correlation domainagent.CorrelationIDs, kind taskengine.JobKind, idempotencyKey string, step string) {
	if handler == nil || handler.supervisorService == nil {
		return
	}
	_, _ = handler.supervisorService.OnCheckpointSaved(ctx, taskengine.CorrelationIDs{RunID: correlation.RunID, TaskID: correlation.TaskID, JobID: correlation.JobID, ProjectID: correlation.ProjectID}, kind, idempotencyKey, step)
}

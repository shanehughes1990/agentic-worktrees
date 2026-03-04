package worker

import (
	applicationagent "agentic-orchestrator/internal/application/agent"
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domainagent "agentic-orchestrator/internal/domain/agent"
	domainscm "agentic-orchestrator/internal/domain/scm"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
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
	TrackerBoardID        string                 `json:"tracker_board_id,omitempty"`
	TrackerTaskID         string                 `json:"tracker_task_id,omitempty"`
	TrackerClaimID        string                 `json:"tracker_claim_id,omitempty"`
	ResumeCheckpoint      *taskengine.Checkpoint `json:"resume_checkpoint,omitempty"`
	ResumeCheckpointStep  string                 `json:"resume_checkpoint_step,omitempty"`
	ResumeCheckpointToken string                 `json:"resume_checkpoint_token,omitempty"`
}

type trackerTaskService interface {
	ClaimNextTask(ctx context.Context, request applicationtracker.ClaimNextTaskRequest) (applicationtracker.ClaimedTask, error)
	ApplyTaskResult(ctx context.Context, request applicationtracker.ApplyTaskResultRequest) (applicationtracker.AppliedTaskResult, error)
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
	projectRepository projectSetupLookup
	trackerService    trackerTaskService
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
	return newAgentWorkflowHandlerWithTracker(
		&projectSetupAgentRuntimeResolver{projectRepository: projectRepository, serviceFactory: serviceFactory},
		projectRepository,
		nil,
		checkpointStore,
		executionJournal,
		supervisorService,
	)
}

func NewAgentWorkflowHandlerWithProjectSetupAndTracker(projectRepository projectSetupLookup, trackerService trackerTaskService, serviceFactory agentServiceFactoryFunc, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandlerWithTracker(
		&projectSetupAgentRuntimeResolver{projectRepository: projectRepository, serviceFactory: serviceFactory},
		projectRepository,
		trackerService,
		checkpointStore,
		executionJournal,
		supervisorService,
	)
}

func newAgentWorkflowHandler(runtimeResolver agentRuntimeResolver, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	return newAgentWorkflowHandlerWithTracker(runtimeResolver, nil, nil, checkpointStore, executionJournal, supervisorService)
}

func newAgentWorkflowHandlerWithTracker(runtimeResolver agentRuntimeResolver, projectRepository projectSetupLookup, trackerService trackerTaskService, checkpointStore taskengine.CheckpointStore, executionJournal taskengine.ExecutionJournal, supervisorService supervisorSignalService) (*AgentWorkflowHandler, error) {
	if runtimeResolver == nil {
		return nil, fmt.Errorf("agent runtime resolver is required")
	}
	return &AgentWorkflowHandler{runtimeResolver: runtimeResolver, projectRepository: projectRepository, trackerService: trackerService, checkpointStore: checkpointStore, executionJournal: executionJournal, supervisorService: supervisorService}, nil
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

	boardID := strings.TrimSpace(payload.TrackerBoardID)
	if boardID == "" && handler.projectRepository != nil && strings.TrimSpace(payload.ProjectID) != "" {
		setup, setupErr := handler.projectRepository.GetProjectSetup(ctx, strings.TrimSpace(payload.ProjectID))
		if setupErr != nil {
			handler.safeRecordExecution(ctx, request, job.Kind, "source_state", taskengine.ExecutionStatusFailed, setupErr.Error())
			return fmt.Errorf("load project setup for tracker claim: %w", setupErr)
		}
		if setup != nil && len(setup.Boards) > 0 {
			boardID = strings.TrimSpace(setup.Boards[0].BoardID)
		}
	}

	claimID := strings.TrimSpace(payload.TrackerClaimID)
	claimedTaskID := strings.TrimSpace(payload.TrackerTaskID)
	if handler.trackerService != nil && strings.TrimSpace(payload.ProjectID) != "" && boardID != "" && claimID == "" {
		claimed, claimErr := handler.trackerService.ClaimNextTask(ctx, applicationtracker.ClaimNextTaskRequest{
			ProjectID: strings.TrimSpace(payload.ProjectID),
			BoardID:   boardID,
			AgentID:   strings.TrimSpace(payload.SessionID),
			LeaseTTL:  30 * time.Minute,
		})
		if claimErr != nil {
			handler.safeRecordExecution(ctx, request, job.Kind, "source_state", taskengine.ExecutionStatusFailed, claimErr.Error())
			return fmt.Errorf("claim next tracker task: %w", claimErr)
		}
		claimID = strings.TrimSpace(claimed.ClaimToken)
		claimedTaskID = strings.TrimSpace(string(claimed.Task.ID))
		request.Prompt = mergeTaskPrompt(request.Prompt, claimed.Task)
	}

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

	if handler.trackerService != nil && strings.TrimSpace(payload.ProjectID) != "" && boardID != "" && claimID != "" && claimedTaskID != "" {
		_, applyErr := handler.trackerService.ApplyTaskResult(ctx, applicationtracker.ApplyTaskResultRequest{
			ProjectID:      strings.TrimSpace(payload.ProjectID),
			BoardID:        boardID,
			ClaimToken:     claimID,
			TaskID:         claimedTaskID,
			NextState:      domaintracker.TaskStateCompleted,
			OutcomeStatus:  domaintracker.OutcomeStatusSuccess,
			OutcomeSummary: "agent workflow completed",
		})
		if applyErr != nil {
			handler.safeRecordExecution(ctx, request, job.Kind, step, taskengine.ExecutionStatusFailed, applyErr.Error())
			return fmt.Errorf("apply tracker task result: %w", applyErr)
		}
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

func mergeTaskPrompt(basePrompt string, task domaintracker.Task) string {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = strings.TrimSpace(string(task.ID))
	}
	details := strings.TrimSpace(task.Description)
	if details == "" {
		details = "No additional task description provided."
	}
	taskPrompt := fmt.Sprintf("Work item: %s\nTask ID: %s\nDetails: %s", title, strings.TrimSpace(string(task.ID)), details)
	if strings.TrimSpace(basePrompt) == "" {
		return taskPrompt
	}
	return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(basePrompt), taskPrompt)
}

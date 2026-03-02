package controlplane

import (
	applicationsupervisor "agentic-orchestrator/internal/application/supervisor"
	"agentic-orchestrator/internal/application/taskengine"
	domainsupervisor "agentic-orchestrator/internal/domain/supervisor"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type SessionSummary struct {
	RunID     string
	TaskCount int
	JobCount  int
	UpdatedAt time.Time
}

type WorkflowJob struct {
	RunID          string
	TaskID         string
	JobID          string
	ProjectID      string
	JobKind        taskengine.JobKind
	IdempotencyKey string
	QueueTaskID    string
	Queue          string
	Status         string
	Duplicate      bool
	EnqueuedAt     time.Time
	UpdatedAt      time.Time
}

type ExecutionHistoryRecord struct {
	RunID          string
	TaskID         string
	JobID          string
	ProjectID      string
	JobKind        taskengine.JobKind
	IdempotencyKey string
	Step           string
	Status         taskengine.ExecutionStatus
	ErrorMessage   string
	UpdatedAt      time.Time
}

type DeadLetterHistoryRecord struct {
	Queue      string
	TaskID     string
	JobKind    taskengine.JobKind
	Action     taskengine.DeadLetterAction
	LastError  string
	Reason     string
	Actor      string
	OccurredAt time.Time
}

type ProjectSetup struct {
	ProjectID       string
	ProjectName     string
	SCMProvider     string
	RepositoryURL   string
	TrackerProvider string
	TrackerLocation string
	TrackerBoardID  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpsertProjectSetupRequest struct {
	ProjectID       string
	ProjectName     string
	SCMProvider     string
	RepositoryURL   string
	TrackerProvider string
	TrackerLocation string
	TrackerBoardID  string
}

func (request UpsertProjectSetupRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(request.ProjectName) == "" {
		return fmt.Errorf("project_name is required")
	}
	if strings.TrimSpace(request.SCMProvider) != "github" {
		return fmt.Errorf("scm_provider must be github")
	}
	if strings.TrimSpace(request.RepositoryURL) == "" {
		return fmt.Errorf("repository_url is required")
	}
	if parsed, err := url.ParseRequestURI(strings.TrimSpace(request.RepositoryURL)); err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("repository_url must be a valid absolute URL")
	}
	if err := domaintracker.SourceKind(strings.TrimSpace(request.TrackerProvider)).Validate(); err != nil {
		return fmt.Errorf("tracker_provider: %w", err)
	}
	return nil
}

type CorrelationFilter struct {
	RunID  string
	TaskID string
	JobID  string
	ProjectID string
}

type QueryRepository interface {
	ListSessions(ctx context.Context, limit int) ([]SessionSummary, error)
	GetSession(ctx context.Context, runID string) (*SessionSummary, error)
	ListWorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]WorkflowJob, error)
	ListExecutionHistory(ctx context.Context, filter CorrelationFilter, limit int) ([]ExecutionHistoryRecord, error)
	ListDeadLetterHistory(ctx context.Context, queue string, limit int) ([]DeadLetterHistoryRecord, error)
}

type ProjectSetupRepository interface {
	ListProjectSetups(ctx context.Context, limit int) ([]ProjectSetup, error)
	GetProjectSetup(ctx context.Context, projectID string) (*ProjectSetup, error)
	UpsertProjectSetup(ctx context.Context, setup ProjectSetup) (*ProjectSetup, error)
	DeleteProjectSetup(ctx context.Context, projectID string) error
}

type ProjectCleanupManager interface {
	CleanupProjectArtifacts(ctx context.Context, setup ProjectSetup) error
}

type IngestionBoardSource struct {
	Kind     string
	Location string
	BoardID  string
	Config   map[string]any
}

type EnqueueIngestionWorkflowRequest struct {
	RunID          string
	TaskID         string
	JobID          string
	IdempotencyKey string
	Prompt         string
	ProjectID      string
	WorkflowID     string
	BoardSource    IngestionBoardSource
}

func (request EnqueueIngestionWorkflowRequest) Validate() error {
	if strings.TrimSpace(request.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(request.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(request.JobID) == "" {
		return fmt.Errorf("job_id is required")
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(request.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(request.WorkflowID) == "" {
		return fmt.Errorf("workflow_id is required")
	}
	if strings.TrimSpace(request.BoardSource.Kind) == "" {
		return fmt.Errorf("board_source.kind is required")
	}
	return nil
}

type ApproveIssueIntakeRequest struct {
	RunID          string
	TaskID         string
	JobID          string
	ProjectID      string
	Source         string
	IssueReference string
	ApprovedBy     string
}

func (request ApproveIssueIntakeRequest) Validate() error {
	if strings.TrimSpace(request.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(request.TaskID) == "" {
		return fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(request.JobID) == "" {
		return fmt.Errorf("job_id is required")
	}
	if strings.TrimSpace(request.IssueReference) == "" {
		return fmt.Errorf("issue_reference is required")
	}
	if strings.TrimSpace(request.ApprovedBy) == "" {
		return fmt.Errorf("approved_by is required")
	}
	return nil
}

type Service struct {
	scheduler         *taskengine.Scheduler
	supervisorService *applicationsupervisor.Service
	queryRepository   QueryRepository
	projectRepository ProjectSetupRepository
	deadLetterManager taskengine.DeadLetterManager
	cleanupManager    ProjectCleanupManager
}

func NewService(scheduler *taskengine.Scheduler, supervisorService *applicationsupervisor.Service, queryRepository QueryRepository, projectRepository ProjectSetupRepository, deadLetterManager taskengine.DeadLetterManager) (*Service, error) {
	if queryRepository == nil {
		return nil, fmt.Errorf("control-plane query repository is required")
	}
	if projectRepository == nil {
		return nil, fmt.Errorf("control-plane project repository is required")
	}
	return &Service{
		scheduler:         scheduler,
		supervisorService: supervisorService,
		queryRepository:   queryRepository,
		projectRepository: projectRepository,
		deadLetterManager: deadLetterManager,
	}, nil
}

func (service *Service) SetProjectCleanupManager(cleanupManager ProjectCleanupManager) {
	if service == nil {
		return
	}
	service.cleanupManager = cleanupManager
}

func (service *Service) Sessions(ctx context.Context, limit int) ([]SessionSummary, error) {
	return service.queryRepository.ListSessions(ctx, normalizeLimit(limit, 50, 250))
}

func (service *Service) Session(ctx context.Context, runID string) (*SessionSummary, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	return service.queryRepository.GetSession(ctx, runID)
}

func (service *Service) WorkflowJobs(ctx context.Context, runID string, taskID string, limit int) ([]WorkflowJob, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	return service.queryRepository.ListWorkflowJobs(ctx, runID, strings.TrimSpace(taskID), normalizeLimit(limit, 100, 500))
}

func (service *Service) ExecutionHistory(ctx context.Context, filter CorrelationFilter, limit int) ([]ExecutionHistoryRecord, error) {
	filter.RunID = strings.TrimSpace(filter.RunID)
	filter.TaskID = strings.TrimSpace(filter.TaskID)
	filter.JobID = strings.TrimSpace(filter.JobID)
	if filter.RunID == "" || filter.TaskID == "" || filter.JobID == "" {
		return nil, fmt.Errorf("run_id, task_id, and job_id are required")
	}
	return service.queryRepository.ListExecutionHistory(ctx, filter, normalizeLimit(limit, 100, 500))
}

func (service *Service) DeadLetterHistory(ctx context.Context, queue string, limit int) ([]DeadLetterHistoryRecord, error) {
	return service.queryRepository.ListDeadLetterHistory(ctx, strings.TrimSpace(queue), normalizeLimit(limit, 100, 500))
}

func (service *Service) ProjectSetups(ctx context.Context, limit int) ([]ProjectSetup, error) {
	if service == nil || service.projectRepository == nil {
		return nil, fmt.Errorf("project repository is not configured")
	}
	return service.projectRepository.ListProjectSetups(ctx, normalizeLimit(limit, 50, 250))
}

func (service *Service) ProjectSetup(ctx context.Context, projectID string) (*ProjectSetup, error) {
	if service == nil || service.projectRepository == nil {
		return nil, fmt.Errorf("project repository is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	return service.projectRepository.GetProjectSetup(ctx, projectID)
}

func (service *Service) UpsertProjectSetup(ctx context.Context, request UpsertProjectSetupRequest) (*ProjectSetup, error) {
	if service == nil || service.projectRepository == nil {
		return nil, fmt.Errorf("project repository is not configured")
	}
	request.ProjectID = strings.TrimSpace(request.ProjectID)
	request.ProjectName = strings.TrimSpace(request.ProjectName)
	request.SCMProvider = strings.ToLower(strings.TrimSpace(request.SCMProvider))
	request.RepositoryURL = strings.TrimSpace(request.RepositoryURL)
	request.TrackerProvider = strings.TrimSpace(request.TrackerProvider)
	request.TrackerLocation = strings.TrimSpace(request.TrackerLocation)
	request.TrackerBoardID = strings.TrimSpace(request.TrackerBoardID)
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return service.projectRepository.UpsertProjectSetup(ctx, ProjectSetup{
		ProjectID:       request.ProjectID,
		ProjectName:     request.ProjectName,
		SCMProvider:     request.SCMProvider,
		RepositoryURL:   request.RepositoryURL,
		TrackerProvider: request.TrackerProvider,
		TrackerLocation: request.TrackerLocation,
		TrackerBoardID:  request.TrackerBoardID,
	})
}

func (service *Service) DeleteProjectSetup(ctx context.Context, projectID string) error {
	if service == nil || service.projectRepository == nil {
		return fmt.Errorf("project repository is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	projectSetup, err := service.projectRepository.GetProjectSetup(ctx, projectID)
	if err != nil {
		return fmt.Errorf("load project setup: %w", err)
	}
	if projectSetup == nil {
		return fmt.Errorf("project setup not found")
	}
	if service.deadLetterManager != nil {
		if err := service.deadLetterManager.DeleteProjectTasks(ctx, projectID); err != nil {
			return fmt.Errorf("delete project tasks: %w", err)
		}
	}
	if service.cleanupManager != nil {
		if err := service.cleanupManager.CleanupProjectArtifacts(ctx, *projectSetup); err != nil {
			return fmt.Errorf("cleanup project artifacts: %w", err)
		}
	}
	if err := service.projectRepository.DeleteProjectSetup(ctx, projectID); err != nil {
		return fmt.Errorf("delete project setup: %w", err)
	}
	return nil
}

func (service *Service) EnqueueIngestionWorkflow(ctx context.Context, request EnqueueIngestionWorkflowRequest) (taskengine.EnqueueResult, error) {
	if service == nil || service.scheduler == nil {
		return taskengine.EnqueueResult{}, fmt.Errorf("task scheduler is not configured")
	}
	if err := request.Validate(); err != nil {
		return taskengine.EnqueueResult{}, err
	}
	payload, err := json.Marshal(map[string]any{
		"run_id":          strings.TrimSpace(request.RunID),
		"task_id":         strings.TrimSpace(request.TaskID),
		"job_id":          strings.TrimSpace(request.JobID),
		"idempotency_key": strings.TrimSpace(request.IdempotencyKey),
		"prompt":          strings.TrimSpace(request.Prompt),
		"project_id":      strings.TrimSpace(request.ProjectID),
		"workflow_id":     strings.TrimSpace(request.WorkflowID),
		"board_source": map[string]any{
			"kind":     strings.TrimSpace(request.BoardSource.Kind),
			"location": strings.TrimSpace(request.BoardSource.Location),
			"board_id": strings.TrimSpace(request.BoardSource.BoardID),
			"config":   request.BoardSource.Config,
		},
	})
	if err != nil {
		return taskengine.EnqueueResult{}, fmt.Errorf("encode ingestion workflow payload: %w", err)
	}
	result, err := service.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindIngestionAgent,
		Payload:        payload,
		IdempotencyKey: strings.TrimSpace(request.IdempotencyKey),
		CorrelationIDs: taskengine.CorrelationIDs{RunID: strings.TrimSpace(request.RunID), TaskID: strings.TrimSpace(request.TaskID), JobID: strings.TrimSpace(request.JobID), ProjectID: strings.TrimSpace(request.ProjectID)},
	})
	if err != nil {
		return taskengine.EnqueueResult{}, fmt.Errorf("enqueue ingestion workflow: %w", err)
	}
	return result, nil
}

func (service *Service) ApproveIssueIntake(ctx context.Context, request ApproveIssueIntakeRequest) (domainsupervisor.Decision, error) {
	if service == nil || service.supervisorService == nil {
		return domainsupervisor.Decision{}, fmt.Errorf("supervisor service is not configured")
	}
	if err := request.Validate(); err != nil {
		return domainsupervisor.Decision{}, err
	}
	decision, err := service.supervisorService.OnIssueApproved(ctx, taskengine.CorrelationIDs{RunID: strings.TrimSpace(request.RunID), TaskID: strings.TrimSpace(request.TaskID), JobID: strings.TrimSpace(request.JobID), ProjectID: strings.TrimSpace(request.ProjectID)}, strings.TrimSpace(request.Source), strings.TrimSpace(request.IssueReference), strings.TrimSpace(request.ApprovedBy))
	if err != nil {
		return domainsupervisor.Decision{}, fmt.Errorf("approve issue intake: %w", err)
	}
	return decision, nil
}

func (service *Service) RequeueDeadLetter(ctx context.Context, queue string, taskID string) error {
	if service == nil || service.deadLetterManager == nil {
		return fmt.Errorf("dead-letter manager is not configured")
	}
	if err := service.deadLetterManager.RequeueDeadLetter(ctx, strings.TrimSpace(queue), strings.TrimSpace(taskID)); err != nil {
		return fmt.Errorf("requeue dead letter: %w", err)
	}
	return nil
}

func normalizeLimit(limit int, fallback int, max int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}

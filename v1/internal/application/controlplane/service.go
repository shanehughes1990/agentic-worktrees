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

type ProjectRepository struct {
	RepositoryID string
	SCMProvider  string
	RepositoryURL string
	IsPrimary    bool
}

type ProjectBoard struct {
	BoardID                  string
	TrackerProvider          string
	TrackerLocation          string
	TrackerBoardID           string
	AppliesToAllRepositories bool
	RepositoryIDs            []string
}

type ProjectSetup struct {
	ProjectID   string
	ProjectName string
	Repositories []ProjectRepository
	Boards      []ProjectBoard
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpsertProjectSetupRequest struct {
	ProjectID    string
	ProjectName  string
	Repositories []ProjectRepository
	Boards       []ProjectBoard
}

var supportedTrackerProviders = map[string]struct{}{
	"local_json":    {},
	"github_issues": {},
}

func (request UpsertProjectSetupRequest) Validate() error {
	if strings.TrimSpace(request.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(request.ProjectName) == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(request.Repositories) == 0 {
		return fmt.Errorf("at least one repository is required")
	}
	for index, repository := range request.Repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		if repositoryID == "" {
			return fmt.Errorf("repositories[%d].repository_id is required", index)
		}
		if strings.TrimSpace(repository.SCMProvider) != "github" {
			return fmt.Errorf("repositories[%d].scm_provider must be github", index)
		}
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryURL == "" {
			return fmt.Errorf("repositories[%d].repository_url is required", index)
		}
		if parsed, err := url.ParseRequestURI(repositoryURL); err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("repositories[%d].repository_url must be a valid absolute URL", index)
		}
	}
	if len(request.Boards) != 1 {
		return fmt.Errorf("exactly one board is required")
	}
	for index, board := range request.Boards {
		if strings.TrimSpace(board.BoardID) == "" {
			return fmt.Errorf("boards[%d].board_id is required", index)
		}
		trackerProvider := strings.ToLower(strings.TrimSpace(board.TrackerProvider))
		if _, supported := supportedTrackerProviders[trackerProvider]; !supported {
			return fmt.Errorf("boards[%d].tracker_provider must be one of: local_json, github_issues", index)
		}
		if !board.AppliesToAllRepositories {
			return fmt.Errorf("boards[%d].applies_to_all_repositories must be true", index)
		}
		if len(board.RepositoryIDs) > 0 {
			return fmt.Errorf("boards[%d].repository_ids is not supported", index)
		}
		if strings.TrimSpace(board.TrackerLocation) == "" {
			return fmt.Errorf("boards[%d].tracker_location is required", index)
		}
	}
	return nil
}

type CorrelationFilter struct {
	RunID     string
	TaskID    string
	JobID     string
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
	BoardID                  string
	Kind                     string
	Location                 string
	ExternalBoardID          string
	AppliesToAllRepositories bool
	RepositoryIDs            []string
	Config                   map[string]any
}

type EnqueueIngestionWorkflowRequest struct {
	RunID          string
	TaskID         string
	JobID          string
	IdempotencyKey string
	Prompt         string
	ProjectID      string
	WorkflowID     string
	BoardSources   []IngestionBoardSource
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
	if len(request.BoardSources) != 1 {
		return fmt.Errorf("exactly one board_source is required")
	}
	for index, source := range request.BoardSources {
		if strings.TrimSpace(source.BoardID) == "" {
			return fmt.Errorf("board_sources[%d].board_id is required", index)
		}
		if err := domaintracker.SourceKind(strings.ToLower(strings.TrimSpace(source.Kind))).Validate(); err != nil {
			return fmt.Errorf("board_sources[%d].kind: %w", index, err)
		}
		if !source.AppliesToAllRepositories {
			return fmt.Errorf("board_sources[%d].applies_to_all_repositories must be true", index)
		}
		if len(source.RepositoryIDs) > 0 {
			return fmt.Errorf("board_sources[%d].repository_ids is not supported", index)
		}
		if strings.TrimSpace(source.Location) == "" {
			return fmt.Errorf("board_sources[%d].location is required", index)
		}
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
	for index := range request.Repositories {
		request.Repositories[index].RepositoryID = strings.TrimSpace(request.Repositories[index].RepositoryID)
		request.Repositories[index].SCMProvider = strings.ToLower(strings.TrimSpace(request.Repositories[index].SCMProvider))
		request.Repositories[index].RepositoryURL = strings.TrimSpace(request.Repositories[index].RepositoryURL)
	}
	for index := range request.Boards {
		request.Boards[index].BoardID = strings.TrimSpace(request.Boards[index].BoardID)
		request.Boards[index].TrackerProvider = strings.ToLower(strings.TrimSpace(request.Boards[index].TrackerProvider))
		request.Boards[index].TrackerLocation = strings.TrimSpace(request.Boards[index].TrackerLocation)
		request.Boards[index].TrackerBoardID = strings.TrimSpace(request.Boards[index].TrackerBoardID)
		request.Boards[index].AppliesToAllRepositories = true
		request.Boards[index].RepositoryIDs = nil
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return service.projectRepository.UpsertProjectSetup(ctx, ProjectSetup{
		ProjectID:    request.ProjectID,
		ProjectName:  request.ProjectName,
		Repositories: request.Repositories,
		Boards:       request.Boards,
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
	boardSources := make([]map[string]any, 0, len(request.BoardSources))
	for _, source := range request.BoardSources {
		boardSources = append(boardSources, map[string]any{
			"board_id":                     strings.TrimSpace(source.BoardID),
			"kind":                         strings.TrimSpace(source.Kind),
			"location":                     strings.TrimSpace(source.Location),
			"external_board_id":            strings.TrimSpace(source.ExternalBoardID),
			"applies_to_all_repositories":  source.AppliesToAllRepositories,
			"repository_ids":               source.RepositoryIDs,
			"config":                       source.Config,
		})
	}
	payload, err := json.Marshal(map[string]any{
		"run_id":          strings.TrimSpace(request.RunID),
		"task_id":         strings.TrimSpace(request.TaskID),
		"job_id":          strings.TrimSpace(request.JobID),
		"idempotency_key": strings.TrimSpace(request.IdempotencyKey),
		"prompt":          strings.TrimSpace(request.Prompt),
		"project_id":      strings.TrimSpace(request.ProjectID),
		"workflow_id":     strings.TrimSpace(request.WorkflowID),
		"board_sources":   boardSources,
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

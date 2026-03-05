package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
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
	RepositoryID  string
	SCMID         string
	RepositoryURL string
	IsPrimary     bool
}

type ProjectSCM struct {
	SCMID      string
	SCMProvider string
	SCMToken    string
}

type ProjectBoard struct {
	BoardID                  string
	TrackerProvider          string
	TaskboardName            string
	AppliesToAllRepositories bool
	RepositoryIDs            []string
}

type ProjectSetup struct {
	ProjectID    string
	ProjectName  string
	SCMs         []ProjectSCM
	Repositories []ProjectRepository
	Boards       []ProjectBoard
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProjectRepositoryBranches struct {
	RepositoryID  string
	RepositoryURL string
	DefaultBranch string
	Branches      []string
}

type ProjectRepositoryBranchCatalog interface {
	ListOriginBranches(ctx context.Context, projectID string, scm ProjectSCM, repository ProjectRepository) ([]string, string, error)
}

type UpsertProjectSetupRequest struct {
	ProjectID    string
	ProjectName  string
	SCMs         []ProjectSCM
	Repositories []ProjectRepository
	Boards       []ProjectBoard
}

var supportedTrackerProviders = map[string]struct{}{
	"internal": {},
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
	if len(request.SCMs) == 0 {
		return fmt.Errorf("at least one scm is required")
	}
	scmByID := make(map[string]struct{}, len(request.SCMs))
	for index, scm := range request.SCMs {
		scmID := strings.TrimSpace(scm.SCMID)
		if scmID == "" {
			return fmt.Errorf("scms[%d].scm_id is required", index)
		}
		if _, exists := scmByID[scmID]; exists {
			return fmt.Errorf("scms[%d].scm_id must be unique", index)
		}
		scmByID[scmID] = struct{}{}
		if strings.TrimSpace(scm.SCMProvider) != "github" {
			return fmt.Errorf("scms[%d].scm_provider must be github", index)
		}
	}
	for index, repository := range request.Repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		if repositoryID == "" {
			return fmt.Errorf("repositories[%d].repository_id is required", index)
		}
		scmID := strings.TrimSpace(repository.SCMID)
		if scmID == "" {
			return fmt.Errorf("repositories[%d].scm_id is required", index)
		}
		if _, exists := scmByID[scmID]; !exists {
			return fmt.Errorf("repositories[%d].scm_id must reference an existing scm", index)
		}
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryURL == "" {
			return fmt.Errorf("repositories[%d].repository_url is required", index)
		}
		if parsed, err := url.ParseRequestURI(repositoryURL); err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("repositories[%d].repository_url must be a valid absolute URL", index)
		}
	}
	if len(request.Boards) > 1 {
		return fmt.Errorf("at most one board is supported")
	}
	for index, board := range request.Boards {
		trackerProvider := strings.ToLower(strings.TrimSpace(board.TrackerProvider))
		if _, supported := supportedTrackerProviders[trackerProvider]; !supported {
			return fmt.Errorf("boards[%d].tracker_provider must be internal", index)
		}
		if !board.AppliesToAllRepositories {
			return fmt.Errorf("boards[%d].applies_to_all_repositories must be true", index)
		}
		if len(board.RepositoryIDs) > 0 {
			return fmt.Errorf("boards[%d].repository_ids is not supported", index)
		}
		if strings.TrimSpace(board.TaskboardName) == "" {
			return fmt.Errorf("boards[%d].taskboard_name is required", index)
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
	queryRepository   QueryRepository
	projectRepository ProjectSetupRepository
	repositoryBranchCatalog ProjectRepositoryBranchCatalog
	projectDocumentRepository ProjectDocumentRepository
	projectFileStore          ProjectFileStore
	projectCDNSigner          ProjectCDNSigner
	projectDocumentRootPrefix string
	projectDocumentRemoteStorageType string
	projectDocumentGoogleApplicationCredentialsPath string
	projectDocumentUploadWait time.Duration
	deadLetterManager taskengine.DeadLetterManager
	cleanupManager    ProjectCleanupManager
}

func NewService(scheduler *taskengine.Scheduler, queryRepository QueryRepository, projectRepository ProjectSetupRepository, deadLetterManager taskengine.DeadLetterManager) (*Service, error) {
	if queryRepository == nil {
		return nil, fmt.Errorf("control-plane query repository is required")
	}
	if projectRepository == nil {
		return nil, fmt.Errorf("control-plane project repository is required")
	}
	return &Service{
		scheduler:         scheduler,
		queryRepository:   queryRepository,
		projectRepository: projectRepository,
		projectDocumentRootPrefix: "projects",
		projectDocumentRemoteStorageType: "gcs",
		projectDocumentUploadWait: 5 * time.Second,
		deadLetterManager: deadLetterManager,
	}, nil
}

func (service *Service) SetProjectCleanupManager(cleanupManager ProjectCleanupManager) {
	if service == nil {
		return
	}
	service.cleanupManager = cleanupManager
}

func (service *Service) SetProjectRepositoryBranchCatalog(catalog ProjectRepositoryBranchCatalog) {
	if service == nil {
		return
	}
	service.repositoryBranchCatalog = catalog
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

func (service *Service) ProjectRepositoryBranches(ctx context.Context, projectID string) ([]ProjectRepositoryBranches, error) {
	if service == nil || service.projectRepository == nil {
		return nil, fmt.Errorf("project repository is not configured")
	}
	if service.repositoryBranchCatalog == nil {
		return nil, fmt.Errorf("project repository branch catalog is not configured")
	}
	cleanProjectID := strings.TrimSpace(projectID)
	if cleanProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	setup, err := service.projectRepository.GetProjectSetup(ctx, cleanProjectID)
	if err != nil {
		return nil, fmt.Errorf("load project setup: %w", err)
	}
	if setup == nil {
		return nil, fmt.Errorf("project setup not found")
	}
	scmByID := make(map[string]ProjectSCM, len(setup.SCMs))
	for _, scm := range setup.SCMs {
		scmID := strings.TrimSpace(scm.SCMID)
		if scmID == "" {
			continue
		}
		scmByID[scmID] = scm
	}
	result := make([]ProjectRepositoryBranches, 0, len(setup.Repositories))
	for _, repository := range setup.Repositories {
		scm, ok := scmByID[strings.TrimSpace(repository.SCMID)]
		if !ok {
			return nil, fmt.Errorf("repository %q references unknown scm_id %q", strings.TrimSpace(repository.RepositoryID), strings.TrimSpace(repository.SCMID))
		}
		branches, defaultBranch, listErr := service.repositoryBranchCatalog.ListOriginBranches(ctx, cleanProjectID, scm, repository)
		if listErr != nil {
			return nil, fmt.Errorf("list origin branches for repository %q: %w", strings.TrimSpace(repository.RepositoryID), listErr)
		}
		normalizedBranches := normalizeBranchList(branches)
		resolvedDefault := strings.TrimSpace(defaultBranch)
		if resolvedDefault == "" && len(normalizedBranches) > 0 {
			resolvedDefault = normalizedBranches[0]
		}
		result = append(result, ProjectRepositoryBranches{
			RepositoryID:  strings.TrimSpace(repository.RepositoryID),
			RepositoryURL: strings.TrimSpace(repository.RepositoryURL),
			DefaultBranch: resolvedDefault,
			Branches:      normalizedBranches,
		})
	}
	return result, nil
}

func normalizeBranchList(branches []string) []string {
	seen := make(map[string]struct{}, len(branches))
	normalized := make([]string, 0, len(branches))
	for _, branch := range branches {
		cleanBranch := strings.TrimSpace(branch)
		if cleanBranch == "" {
			continue
		}
		if _, exists := seen[cleanBranch]; exists {
			continue
		}
		seen[cleanBranch] = struct{}{}
		normalized = append(normalized, cleanBranch)
	}
	sort.Strings(normalized)
	for _, preferred := range []string{"main", "master"} {
		for index, branch := range normalized {
			if branch != preferred {
				continue
			}
			if index == 0 {
				return normalized
			}
			normalized = append([]string{branch}, append(normalized[:index], normalized[index+1:]...)...)
			return normalized
		}
	}
	return normalized
}

func (service *Service) UpsertProjectSetup(ctx context.Context, request UpsertProjectSetupRequest) (*ProjectSetup, error) {
	if service == nil || service.projectRepository == nil {
		return nil, fmt.Errorf("project repository is not configured")
	}
	request.ProjectID = strings.TrimSpace(request.ProjectID)
	request.ProjectName = strings.TrimSpace(request.ProjectName)
	for index := range request.SCMs {
		request.SCMs[index].SCMID = strings.TrimSpace(request.SCMs[index].SCMID)
		request.SCMs[index].SCMProvider = strings.ToLower(strings.TrimSpace(request.SCMs[index].SCMProvider))
		request.SCMs[index].SCMToken = strings.TrimSpace(request.SCMs[index].SCMToken)
	}
	for index := range request.Repositories {
		request.Repositories[index].RepositoryID = strings.TrimSpace(request.Repositories[index].RepositoryID)
		request.Repositories[index].SCMID = strings.TrimSpace(request.Repositories[index].SCMID)
		request.Repositories[index].RepositoryURL = strings.TrimSpace(request.Repositories[index].RepositoryURL)
	}
	for index := range request.Boards {
		request.Boards[index].TrackerProvider = strings.ToLower(strings.TrimSpace(request.Boards[index].TrackerProvider))
		request.Boards[index].TaskboardName = strings.TrimSpace(request.Boards[index].TaskboardName)
		request.Boards[index].BoardID = boardIDFromName(request.Boards[index].TaskboardName)
		request.Boards[index].AppliesToAllRepositories = true
		request.Boards[index].RepositoryIDs = nil
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return service.projectRepository.UpsertProjectSetup(ctx, ProjectSetup{
		ProjectID:    request.ProjectID,
		ProjectName:  request.ProjectName,
		SCMs:         request.SCMs,
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

var boardNameSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func boardIDFromName(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	if trimmed == "" {
		return ""
	}
	cleaned := boardNameSanitizer.ReplaceAllString(trimmed, "_")
	return strings.Trim(cleaned, "_")
}

func (service *Service) ApproveIssueIntake(ctx context.Context, request ApproveIssueIntakeRequest) (interface{}, error) {
	_ = ctx
	_ = request
	return nil, fmt.Errorf("supervisor feature removed - ApproveIssueIntake no longer supported")
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

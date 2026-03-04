package ingestion

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Document struct {
	ProjectID   string
	DocumentID  string
	FileName    string
	ContentType string
	ObjectPath  string
	CDNURL      string
	Status      string
}

type BoardStore interface {
	UpsertBoard(ctx context.Context, board domaintracker.Board) error
}

type ArtifactFetcher interface {
	FetchToPath(ctx context.Context, objectPath string, destinationPath string) error
}

type AgentRunner interface {
	GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string) error
}

type RepositorySynchronizer interface {
	Sync(ctx context.Context, projectID string, sandboxDir string, sourceBranch string, sourceRepositories []SourceRepository) error
}

type SourceRepository struct {
	RepositoryID  string
	RepositoryURL string
}

type ArtifactFetcherFunc func(ctx context.Context, objectPath string, destinationPath string) error

func (function ArtifactFetcherFunc) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	return function(ctx, objectPath, destinationPath)
}

type Request struct {
	RunID                     string
	ProjectID                 string
	BoardID                   string
	SelectedDocumentLocations []string
	SourceRepositories        []SourceRepository
	SourceBranch              string
	Model                     string
	SystemPrompt              string
	UserPrompt                string
}

func (request Request) Validate() error {
	if strings.TrimSpace(request.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if len(request.SelectedDocumentLocations) == 0 {
		return failures.WrapTerminal(errors.New("selected_document_locations are required"))
	}
	if strings.TrimSpace(request.SystemPrompt) == "" {
		return failures.WrapTerminal(errors.New("system_prompt is required"))
	}
	return nil
}

type Service struct {
	boardStore             BoardStore
	artifactFetcher        ArtifactFetcher
	agentRunner            AgentRunner
	repositorySynchronizer RepositorySynchronizer
}

const defaultModel = "gpt-5.3-codex"
const defaultSourceBranch = "main"

func NewService(boardStore BoardStore, artifactFetcher ArtifactFetcher, agentRunner AgentRunner, repositorySynchronizer RepositorySynchronizer) (*Service, error) {
	if boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("ingestion board store is required"))
	}
	if artifactFetcher == nil {
		return nil, failures.WrapTerminal(errors.New("ingestion artifact fetcher is required"))
	}
	if agentRunner == nil {
		return nil, failures.WrapTerminal(errors.New("ingestion agent runner is required"))
	}
	return &Service{
		boardStore:             boardStore,
		artifactFetcher:        artifactFetcher,
		agentRunner:            agentRunner,
		repositorySynchronizer: repositorySynchronizer,
	}, nil
}

func (service *Service) Execute(ctx context.Context, request Request) (domaintracker.Board, error) {
	if service == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("ingestion service is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	boardID := strings.TrimSpace(request.BoardID)
	if boardID == "" {
		boardID = defaultBoardID(request.ProjectID)
	}
	now := time.Now().UTC()
	tempDir, err := os.MkdirTemp("", "ingestion-sandbox-")
	if err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("create ingestion temp directory: %w", err))
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()
	sandboxDir := filepath.Join(tempDir, "sandbox")
	documentsDir := filepath.Join(sandboxDir, "documents")
	if err := os.MkdirAll(documentsDir, 0o755); err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("create ingestion sandbox directories: %w", err))
	}

	normalizedSourceRepositories := normalizeSourceRepositories(request.SourceRepositories)
	sourceBranch := strings.TrimSpace(request.SourceBranch)
	if sourceBranch == "" {
		sourceBranch = defaultSourceBranch
	}
	if len(normalizedSourceRepositories) > 0 {
		if service.repositorySynchronizer == nil {
			return domaintracker.Board{}, failures.WrapTerminal(errors.New("repository synchronizer is required when source_repositories are provided"))
		}
		if syncErr := service.repositorySynchronizer.Sync(ctx, strings.TrimSpace(request.ProjectID), sandboxDir, sourceBranch, normalizedSourceRepositories); syncErr != nil {
			return domaintracker.Board{}, ensureClassified(syncErr)
		}
	}

	hasDocument := false
	fetchedDocuments := make([]fetchedDocument, 0, len(request.SelectedDocumentLocations))
	for index, rawDocumentLocation := range request.SelectedDocumentLocations {
		documentLocation := strings.TrimSpace(rawDocumentLocation)
		if documentLocation == "" {
			continue
		}
		hasDocument = true
		fileName := filepath.Base(documentLocation)
		localFilePath := filepath.Join(documentsDir, fmt.Sprintf("%d_%s", index+1, sanitizeLocalFileName(fileName)))
		if fetchErr := service.artifactFetcher.FetchToPath(ctx, documentLocation, localFilePath); fetchErr != nil {
			return domaintracker.Board{}, ensureClassified(fetchErr)
		}
		fetchedDocuments = append(fetchedDocuments, fetchedDocument{RemoteLocation: documentLocation, LocalPath: localFilePath})
	}
	if !hasDocument {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("selected_document_locations must contain at least one non-empty value"))
	}

	documentDigest, decodeErr := decodeDocumentsForPrompt(fetchedDocuments)
	if decodeErr != nil {
		return domaintracker.Board{}, ensureClassified(fmt.Errorf("decode fetched documents for ingestion prompt: %w", decodeErr))
	}

	outputPath := filepath.Join(sandboxDir, "taskboard.json")
	composedPrompt := composeIngestionPrompt(request, boardID, outputPath, documentDigest, normalizedSourceRepositories, sourceBranch)
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = defaultModel
	}
	if runErr := service.agentRunner.GenerateTaskboard(ctx, sandboxDir, composedPrompt, outputPath, model); runErr != nil {
		return domaintracker.Board{}, ensureClassified(runErr)
	}

	boardJSON, err := os.ReadFile(outputPath)
	if err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("read generated taskboard output: %w", err))
	}

	var board domaintracker.Board
	if err := json.Unmarshal(boardJSON, &board); err != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("decode generated taskboard json: %w", err))
	}

	normalizeBoard(&board, request, boardID, now)
	if err := ensureRepositoryAssignments(board, normalizedSourceRepositories); err != nil {
		return domaintracker.Board{}, err
	}
	if err := board.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	if err := service.boardStore.UpsertBoard(ctx, board); err != nil {
		return domaintracker.Board{}, ensureClassified(fmt.Errorf("persist generated taskboard: %w", err))
	}

	return board, nil
}

type fetchedDocument struct {
	RemoteLocation string
	LocalPath      string
}

func composeIngestionPrompt(request Request, boardID string, outputPath string, decodedDocuments string, sourceRepositories []SourceRepository, sourceBranch string) string {
	segments := []string{
		strings.TrimSpace(request.SystemPrompt),
		"Taskboard synthesis contract:",
		"- You are a taskboard synthesis agent.",
		"- Build a canonical execution taskboard from the provided project documents.",
		"- Unless the user explicitly narrows scope, treat the taskboard as the complete project scope for this ingestion run.",
		"- When multiple repositories exist, break work down clearly by repository and ensure every task is explicitly attributable to a single target repository.",
		"- Extract epics and actionable tasks, preserve dependencies, avoid inventing unsupported facts, and output a deterministic plan suitable for execution orchestration.",
		"Ingestion execution instructions:",
		"- Generate a full-featured taskboard JSON from the decoded project documents.",
		"- Unless the user explicitly asks for partial scope, treat the generated taskboard as the complete scope for this ingestion run.",
		"- Decompose work into the smallest practical execution chunks where each epic and each task represents one concrete outcome.",
		"- Use explicit dependencies where supported by the source materials; treat depends_on as a required modeling tool, not an optional hint.",
		"- Whenever one task could block, invalidate, or conflict with another task, add an explicit depends_on relationship to serialize execution.",
		"- Keep dependencies acyclic and reference only existing IDs.",
		"- Put first-executable tasks early in each epic.tasks list; preserve topological ordering whenever possible.",
		"- Keep independent work parallel by avoiding unnecessary dependencies.",
		"- Do not create duplicate or near-duplicate tasks (including renamed variants with the same intent).",
		"- Do not create parallel sibling tasks that could implement the same behavior in different files/packages.",
		"- If two candidate tasks touch the same artifact, behavior, or acceptance outcome, merge them into one canonical task OR link them with explicit depends_on.",
		"- Prefer one implementation-owner task per behavior change; model additional checks as dependency-aware verification tasks.",
		"- Separate implementation from verification; verification tasks should validate one check per task and avoid bundled checks.",
		"- Ensure task titles are specific and disambiguated by concrete scope (artifact/package + action) so duplicate intent is obvious and prevented.",
		"- Do not invent unsupported requirements or implementation facts.",
		"- Synchronized source repositories (if present) are available under: " + filepath.Join(".", "repos") + "/<repository-name>",
		"- Use source branch \"" + strings.TrimSpace(sourceBranch) + "\" for source repository context.",
		"- Repository layout contract: local source cache is projects/{projectId}/repositories/{repository-name}, then copied into sandbox ./repos/{repository-name}.",
		"- For multi-repository projects, decompose work by repository and ensure every task is clearly mapped to exactly one target repository.",
		"- Include board metadata with repository scope and source branch: metadata.repositories[] and metadata.source_branch.",
		"- For every epic and task, include metadata.repository_id to identify target repository.",
		"- For each task, make repository ownership explicit in task title and description (or equivalent structured metadata if available).",
		"- Write valid JSON only to this exact output path: " + strings.TrimSpace(outputPath),
		"- Ensure board_id is \"" + strings.TrimSpace(boardID) + "\" and run_id is \"" + strings.TrimSpace(request.RunID) + "\".",
		"- Ensure all epic/task board_id values match board_id.",
		"- Required board fields: board_id, run_id, status, epics, created_at, updated_at.",
		"- Required task/epic fields: id, board_id, title, status.",
		"- Use status values only from: not-started, in-progress, completed, blocked.",
		"- For task outcomes, include outcome.status at minimum.",
		decodedDocuments,
	}
	if len(sourceRepositories) > 0 {
		repositoryLines := make([]string, 0, len(sourceRepositories))
		for _, repository := range sourceRepositories {
			repositoryID := strings.TrimSpace(repository.RepositoryID)
			repositoryURL := strings.TrimSpace(repository.RepositoryURL)
			repositoryFolder := deriveRepositoryFolderName(repositoryID, repositoryURL)
			repositoryLines = append(repositoryLines, "- repository_id="+repositoryID+" repository_url="+repositoryURL+" local_dir="+filepath.Join(".", "repos", repositoryFolder))
		}
		segments = append(segments, "Synchronized source repositories:", strings.Join(repositoryLines, "\n"))
	}
	if userPrompt := strings.TrimSpace(request.UserPrompt); userPrompt != "" {
		segments = append(segments, "User requirements:", userPrompt)
	}
	return strings.TrimSpace(strings.Join(segments, "\n\n"))
}

func deriveRepositoryFolderName(repositoryID string, repositoryURL string) string {
	trimmedID := strings.TrimSpace(repositoryID)
	trimmedURL := strings.TrimSpace(repositoryURL)
	if parsedURL, err := url.Parse(trimmedURL); err == nil {
		parts := strings.Split(strings.Trim(strings.TrimSpace(parsedURL.Path), "/"), "/")
		if len(parts) >= 2 {
			repositoryName := strings.TrimSpace(strings.TrimSuffix(parts[len(parts)-1], ".git"))
			if repositoryName != "" {
				return repositoryName
			}
		}
	}
	if trimmedID != "" {
		return trimmedID
	}
	return "repository"
}

func normalizeSourceRepositories(repositories []SourceRepository) []SourceRepository {
	normalized := make([]SourceRepository, 0, len(repositories))
	for _, repository := range repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		normalized = append(normalized, SourceRepository{RepositoryID: repositoryID, RepositoryURL: repositoryURL})
	}
	return normalized
}

func decodeDocumentsForPrompt(documents []fetchedDocument) (string, error) {
	var builder strings.Builder
	builder.WriteString("Decoded documents for ingestion:\n")

	const maxCharsPerDocument = 24000
	for _, document := range documents {
		cleanLocalPath := strings.TrimSpace(document.LocalPath)
		if cleanLocalPath == "" {
			continue
		}
		content, err := os.ReadFile(cleanLocalPath)
		if err != nil {
			return "", fmt.Errorf("read document %s: %w", cleanLocalPath, err)
		}

		builder.WriteString("\n---\n")
		builder.WriteString("Remote location: ")
		builder.WriteString(strings.TrimSpace(document.RemoteLocation))
		builder.WriteString("\nLocal path: ")
		builder.WriteString(cleanLocalPath)
		builder.WriteString("\n")

		if !utf8.Valid(content) {
			builder.WriteString("Decoded content: [non-UTF8/binary content omitted]\n")
			continue
		}

		text := strings.TrimSpace(string(content))
		if len(text) > maxCharsPerDocument {
			text = text[:maxCharsPerDocument]
			text = strings.TrimSpace(text) + "\n...[truncated]"
		}
		builder.WriteString("Decoded content:\n")
		builder.WriteString(text)
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String()), nil
}

func normalizeBoard(board *domaintracker.Board, request Request, boardID string, now time.Time) {
	if board == nil {
		return
	}
	repositoryScope := buildRepositoryScopeMetadata(request.SourceRepositories, request.SourceBranch)
	singleRepositoryID := ""
	if len(repositoryScope) == 1 {
		singleRepositoryID = strings.TrimSpace(fmt.Sprintf("%v", repositoryScope[0]["repository_id"]))
	}

	board.BoardID = strings.TrimSpace(boardID)
	board.RunID = strings.TrimSpace(request.RunID)
	if board.Status == "" {
		board.Status = domaintracker.StatusNotStarted
	} else {
		board.Status = normalizeStatus(board.Status)
	}
	board.Source = domaintracker.SourceRef{
		Kind:     domaintracker.SourceKindInternal,
		BoardID:  strings.TrimSpace(boardID),
		Location: compactSourceLocation(request.SelectedDocumentLocations),
		Metadata: map[string]any{
			"selected_document_locations": request.SelectedDocumentLocations,
			"source_branch":              strings.TrimSpace(request.SourceBranch),
			"repositories":               repositoryScope,
		},
	}
	if board.Metadata == nil {
		board.Metadata = map[string]any{}
	}
	board.Metadata["source_branch"] = strings.TrimSpace(request.SourceBranch)
	board.Metadata["repositories"] = repositoryScope
	if board.CreatedAt.IsZero() {
		board.CreatedAt = now
	}
	if board.UpdatedAt.IsZero() {
		board.UpdatedAt = now
	}

	for epicIndex := range board.Epics {
		epic := &board.Epics[epicIndex]
		if strings.TrimSpace(string(epic.ID)) == "" {
			epic.ID = domaintracker.WorkItemID(fmt.Sprintf("epic-%d", epicIndex+1))
		}
		epic.BoardID = board.BoardID
		if epic.Metadata == nil {
			epic.Metadata = map[string]any{}
		}
		epicRepositoryID := extractRepositoryID(epic.Metadata)
		if epicRepositoryID == "" && singleRepositoryID != "" {
			epicRepositoryID = singleRepositoryID
			epic.Metadata["repository_id"] = epicRepositoryID
		}
		epic.Status = normalizeStatus(epic.Status)
		epic.Priority = normalizePriority(epic.Priority)
		if epic.CreatedAt.IsZero() {
			epic.CreatedAt = now
		}
		if epic.UpdatedAt.IsZero() {
			epic.UpdatedAt = now
		}

		for taskIndex := range epic.Tasks {
			task := &epic.Tasks[taskIndex]
			if strings.TrimSpace(string(task.ID)) == "" {
				task.ID = domaintracker.WorkItemID(fmt.Sprintf("task-%d-%d", epicIndex+1, taskIndex+1))
			}
			task.BoardID = board.BoardID
			if task.Metadata == nil {
				task.Metadata = map[string]any{}
			}
			taskRepositoryID := extractRepositoryID(task.Metadata)
			if taskRepositoryID == "" {
				taskRepositoryID = epicRepositoryID
			}
			if taskRepositoryID == "" && singleRepositoryID != "" {
				taskRepositoryID = singleRepositoryID
			}
			if taskRepositoryID != "" {
				task.Metadata["repository_id"] = taskRepositoryID
			}
			task.Status = normalizeStatus(task.Status)
			task.Priority = normalizePriority(task.Priority)
			if task.CreatedAt.IsZero() {
				task.CreatedAt = now
			}
			if task.UpdatedAt.IsZero() {
				task.UpdatedAt = now
			}
			if task.Outcome != nil {
				task.Outcome.Status = strings.TrimSpace(task.Outcome.Status)
				if task.Outcome.Status == "" {
					task.Outcome.Status = string(task.Status)
				}
				if strings.TrimSpace(task.Outcome.Repository) == "" && taskRepositoryID != "" {
					task.Outcome.Repository = taskRepositoryID
				}
				if task.Outcome.UpdatedAt.IsZero() {
					task.Outcome.UpdatedAt = now
				}
			}
		}
	}
}

func buildRepositoryScopeMetadata(repositories []SourceRepository, sourceBranch string) []map[string]any {
	result := make([]map[string]any, 0, len(repositories))
	for _, repository := range repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		repositoryFolder := deriveRepositoryFolderName(repositoryID, repositoryURL)
		result = append(result, map[string]any{
			"repository_id":  repositoryID,
			"repository_url": repositoryURL,
			"source_branch":  strings.TrimSpace(sourceBranch),
			"local_dir":      filepath.Join(".", "repos", repositoryFolder),
		})
	}
	return result
}

func extractRepositoryID(metadata map[string]any) string {
	if metadata == nil {
		return ""
	}
	for _, key := range []string{"repository_id", "repository", "repo_id", "repo"} {
		if value, ok := metadata[key]; ok {
			trimmed := strings.TrimSpace(fmt.Sprintf("%v", value))
			if trimmed != "" {
				metadata["repository_id"] = trimmed
				return trimmed
			}
		}
	}
	return ""
}

func ensureRepositoryAssignments(board domaintracker.Board, sourceRepositories []SourceRepository) error {
	if len(sourceRepositories) == 0 {
		return nil
	}
	if len(sourceRepositories) == 1 {
		return nil
	}
	for _, epic := range board.Epics {
		epicRepositoryID := extractRepositoryID(epic.Metadata)
		if epicRepositoryID == "" {
			return failures.WrapTerminal(fmt.Errorf("epic %s is missing metadata.repository_id for multi-repository scope", epic.ID))
		}
		for _, task := range epic.Tasks {
			taskRepositoryID := extractRepositoryID(task.Metadata)
			if taskRepositoryID == "" {
				return failures.WrapTerminal(fmt.Errorf("task %s is missing metadata.repository_id for multi-repository scope", task.ID))
			}
		}
	}
	return nil
}

func compactSourceLocation(locations []string) string {
	normalized := make([]string, 0, len(locations))
	for _, location := range locations {
		clean := strings.TrimSpace(location)
		if clean == "" {
			continue
		}
		normalized = append(normalized, clean)
	}
	if len(normalized) == 0 {
		return "ingestion://documents"
	}
	joined := strings.Join(normalized, ";")
	if len(joined) > 1024 {
		return joined[:1024]
	}
	return joined
}

func normalizeStatus(status domaintracker.Status) domaintracker.Status {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case "", "todo", "queued", "not_started", "not started", "not-started":
		return domaintracker.StatusNotStarted
	case "doing", "in_progress", "in progress", "in-progress", "active":
		return domaintracker.StatusInProgress
	case "done", "complete", "completed":
		return domaintracker.StatusCompleted
	case "blocked":
		return domaintracker.StatusBlocked
	default:
		return status
	}
}

func normalizePriority(priority domaintracker.Priority) domaintracker.Priority {
	switch strings.ToLower(strings.TrimSpace(string(priority))) {
	case "", "none":
		return ""
	case "p0", "critical", "urgent", "highest", "high":
		return domaintracker.PriorityP0
	case "p1", "medium-high", "major":
		return domaintracker.PriorityP1
	case "p2", "medium", "normal":
		return domaintracker.PriorityP2
	case "p3", "low", "minor":
		return domaintracker.PriorityP3
	default:
		if numeric, err := strconv.Atoi(strings.TrimSpace(string(priority))); err == nil {
			switch {
			case numeric >= 3:
				return domaintracker.PriorityP0
			case numeric == 2:
				return domaintracker.PriorityP1
			case numeric == 1:
				return domaintracker.PriorityP2
			default:
				return domaintracker.PriorityP3
			}
		}
		return priority
	}
}

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}

var nonAlphaNumericSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func sanitizeLocalFileName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "document.txt"
	}
	trimmed = strings.ReplaceAll(trimmed, string(filepath.Separator), "_")
	trimmed = strings.ReplaceAll(trimmed, "/", "_")
	return trimmed
}

func defaultBoardID(projectID string) string {
	cleanProjectID := strings.ToLower(strings.TrimSpace(projectID))
	cleanProjectID = nonAlphaNumericSanitizer.ReplaceAllString(cleanProjectID, "_")
	cleanProjectID = strings.Trim(cleanProjectID, "_")
	if cleanProjectID == "" {
		cleanProjectID = "project"
	}
	return fmt.Sprintf("%s_ingestion", cleanProjectID)
}

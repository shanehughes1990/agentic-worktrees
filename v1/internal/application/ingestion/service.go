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
	GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string, runContext AgentRunContext) (AgentRunContext, error)
}

type AgentRunContext struct {
	SessionID string
	StreamID  string
}

type RepositorySynchronizer interface {
	Sync(ctx context.Context, projectID string, sandboxDir string, sourceBranch string, sourceRepositories []SourceRepository) error
}

type SourceRepository struct {
	RepositoryID  string
	RepositoryURL string
	SourceBranch  string
}

type ArtifactFetcherFunc func(ctx context.Context, objectPath string, destinationPath string) error

func (function ArtifactFetcherFunc) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	return function(ctx, objectPath, destinationPath)
}

type Request struct {
	RunID                     string
	JobID                     string
	ProjectID                 string
	BoardID                   string
	TaskboardName             string
	StreamID                  string
	SelectedDocumentLocations []string
	PreferSelectedDocuments   bool
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
	if strings.TrimSpace(request.SystemPrompt) == "" {
		return failures.WrapTerminal(errors.New("system_prompt is required"))
	}
	hasPrompt := strings.TrimSpace(request.UserPrompt) != ""
	hasSelectedDocuments := hasNonEmptyValues(request.SelectedDocumentLocations)
	if !hasPrompt && !hasSelectedDocuments {
		return failures.WrapTerminal(errors.New("user_prompt or selected_document_locations is required"))
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
const maxTaskboardValidationAttempts = 3

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
	for index := range normalizedSourceRepositories {
		if strings.TrimSpace(normalizedSourceRepositories[index].SourceBranch) == "" {
			normalizedSourceRepositories[index].SourceBranch = sourceBranch
		}
	}
	if len(normalizedSourceRepositories) > 0 {
		if service.repositorySynchronizer == nil {
			return domaintracker.Board{}, failures.WrapTerminal(errors.New("repository synchronizer is required when source_repositories are provided"))
		}
		if syncErr := service.repositorySynchronizer.Sync(ctx, strings.TrimSpace(request.ProjectID), sandboxDir, sourceBranch, normalizedSourceRepositories); syncErr != nil {
			return domaintracker.Board{}, ensureClassified(syncErr)
		}
	}

	fetchedDocuments := make([]fetchedDocument, 0, len(request.SelectedDocumentLocations))
	for index, rawDocumentLocation := range request.SelectedDocumentLocations {
		documentLocation := strings.TrimSpace(rawDocumentLocation)
		if documentLocation == "" {
			continue
		}
		fileName := filepath.Base(documentLocation)
		localFilePath := filepath.Join(documentsDir, fmt.Sprintf("%d_%s", index+1, sanitizeLocalFileName(fileName)))
		if fetchErr := service.artifactFetcher.FetchToPath(ctx, documentLocation, localFilePath); fetchErr != nil {
			return domaintracker.Board{}, ensureClassified(fetchErr)
		}
		fetchedDocuments = append(fetchedDocuments, fetchedDocument{RemoteLocation: documentLocation, LocalPath: localFilePath})
	}
	documentDigest := "No remote documents were selected for this ingestion run."
	if len(fetchedDocuments) > 0 {
		decodedDocuments, decodeErr := decodeDocumentsForPrompt(fetchedDocuments)
		if decodeErr != nil {
			return domaintracker.Board{}, ensureClassified(fmt.Errorf("decode fetched documents for ingestion prompt: %w", decodeErr))
		}
		documentDigest = decodedDocuments
	}

	outputPath := filepath.Join(sandboxDir, "taskboard.json")
	composedPrompt := composeIngestionPrompt(request, boardID, outputPath, documentDigest, normalizedSourceRepositories)
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = defaultModel
	}
	board, validationErr := service.generateAndValidateBoard(ctx, request, composedPrompt, sandboxDir, outputPath, model, boardID, now, normalizedSourceRepositories)
	if validationErr != nil {
		return domaintracker.Board{}, validationErr
	}

	if err := service.boardStore.UpsertBoard(ctx, board); err != nil {
		return domaintracker.Board{}, ensureClassified(fmt.Errorf("persist generated taskboard: %w", err))
	}

	return board, nil
}

func (service *Service) generateAndValidateBoard(
	ctx context.Context,
	request Request,
	basePrompt string,
	sandboxDir string,
	outputPath string,
	model string,
	boardID string,
	now time.Time,
	normalizedSourceRepositories []SourceRepository,
) (domaintracker.Board, error) {
	var lastValidationErr error
	feedback := ""
	runContext := AgentRunContext{StreamID: ingestionStreamID(request)}

	for attempt := 1; attempt <= maxTaskboardValidationAttempts; attempt++ {
		attemptPrompt := basePrompt
		attemptPrompt = strings.TrimSpace(attemptPrompt + "\n\nExecution continuity:\n- stream_id: " + strings.TrimSpace(runContext.StreamID) + "\n- session_id: " + strings.TrimSpace(runContext.SessionID) + "\n- Reuse this same stream/session for retries and continuation.")
		if strings.TrimSpace(feedback) != "" {
			attemptPrompt = strings.TrimSpace(basePrompt + "\n\nValidation failure feedback:\n" + feedback + "\n\nRegenerate the full JSON so it passes validation exactly.")
			attemptPrompt = strings.TrimSpace(attemptPrompt + "\n\nExecution continuity:\n- stream_id: " + strings.TrimSpace(runContext.StreamID) + "\n- session_id: " + strings.TrimSpace(runContext.SessionID) + "\n- Reuse this same stream/session for retries and continuation.")
		}

		nextRunContext, runErr := service.agentRunner.GenerateTaskboard(ctx, sandboxDir, attemptPrompt, outputPath, model, runContext)
		if runErr != nil {
			return domaintracker.Board{}, ensureClassified(runErr)
		}
		runContext = mergeRunContext(runContext, nextRunContext)

		boardJSON, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("read generated taskboard output: %w", readErr))
		}

		var board domaintracker.Board
		if err := json.Unmarshal(boardJSON, &board); err != nil {
			lastValidationErr = failures.WrapTerminal(fmt.Errorf("decode generated taskboard json: %w", err))
			feedback = lastValidationErr.Error()
			continue
		}
		if err := board.Validate(); err != nil {
			lastValidationErr = err
			feedback = err.Error()
			continue
		}

		normalizeBoard(&board, request, boardID, now, runContext)
		if err := ensureRepositoryAssignments(board, normalizedSourceRepositories); err != nil {
			lastValidationErr = err
			feedback = err.Error()
			continue
		}
		if err := board.Validate(); err != nil {
			lastValidationErr = err
			feedback = err.Error()
			continue
		}

		return board, nil
	}

	if lastValidationErr != nil {
		return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("taskboard generation failed validation after %d attempts: %w", maxTaskboardValidationAttempts, lastValidationErr))
	}
	return domaintracker.Board{}, failures.WrapTerminal(fmt.Errorf("taskboard generation failed validation after %d attempts", maxTaskboardValidationAttempts))
}

func mergeRunContext(current AgentRunContext, next AgentRunContext) AgentRunContext {
	merged := current
	if strings.TrimSpace(next.StreamID) != "" {
		merged.StreamID = strings.TrimSpace(next.StreamID)
	}
	if strings.TrimSpace(next.SessionID) != "" {
		merged.SessionID = strings.TrimSpace(next.SessionID)
	}
	if strings.TrimSpace(merged.StreamID) == "" {
		merged.StreamID = strings.TrimSpace(current.StreamID)
	}
	return merged
}

func ingestionStreamID(request Request) string {
	if clean := strings.TrimSpace(request.StreamID); clean != "" {
		return clean
	}
	if clean := strings.TrimSpace(request.JobID); clean != "" {
		return "ingestion-stream:" + clean
	}
	if clean := strings.TrimSpace(request.RunID); clean != "" {
		return "ingestion-stream:" + clean
	}
	return fmt.Sprintf("ingestion-stream:%d", time.Now().UTC().UnixNano())
}

type fetchedDocument struct {
	RemoteLocation string
	LocalPath      string
}

func composeIngestionPrompt(request Request, boardID string, outputPath string, decodedDocuments string, sourceRepositories []SourceRepository) string {
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
		"- Use the selected source branch for each repository when resolving source context.",
		"- Repository layout contract: local source cache is projects/{projectId}/repositories/{repository-name}, then copied into sandbox ./repos/{repository-name}.",
		"- For multi-repository projects, decompose work by repository and ensure every task is clearly mapped to exactly one target repository.",
		"- Include board metadata with repository scope and source branch: metadata.repositories[] and metadata.source_branch.",
		"- For every epic and task, include metadata.repository_id to identify target repository.",
		"- For each task, make repository ownership explicit in task title and description (or equivalent structured metadata if available).",
		"- Output must be plain-text JSON only (no markdown, no prose, no code fences).",
		"- Write valid JSON only to this exact output path: " + strings.TrimSpace(outputPath),
		"- The JSON must match the exact schema below; do not add extra fields.",
		"- Exact schema contract:\n{\n  \"board_id\": string,\n  \"run_id\": string,\n  \"name\"?: string,\n  \"state\": \"pending\" | \"active\" | \"completed\" | \"failed\",\n  \"epics\": [\n    {\n      \"id\": string,\n      \"board_id\": string,\n      \"title\": string,\n      \"objective\"?: string,\n      \"state\": \"planned\" | \"in_progress\" | \"completed\" | \"blocked\" | \"failed\",\n      \"rank\": number,\n      \"depends_on_epic_ids\"?: string[],\n      \"tasks\": [\n        {\n          \"id\": string,\n          \"board_id\": string,\n          \"epic_id\": string,\n          \"title\": string,\n          \"description\"?: string,\n          \"task_type\": string,\n          \"state\": \"planned\" | \"in_progress\" | \"completed\" | \"failed\" | \"no_work_needed\",\n          \"rank\": number,\n          \"depends_on_task_ids\"?: string[],\n          \"audit\"?: {\n            \"model_provider\": string,\n            \"model_name\": string,\n            \"model_version\"?: string,\n            \"model_run_id\"?: string,\n            \"prompt_fingerprint\"?: string,\n            \"input_tokens\"?: number,\n            \"output_tokens\"?: number,\n            \"started_at\"?: RFC3339 timestamp,\n            \"completed_at\"?: RFC3339 timestamp\n          },\n          \"outcome\"?: {\n            \"status\": \"success\" | \"partial\" | \"failed\",\n            \"summary\": string,\n            \"error_code\"?: string,\n            \"error_message\"?: string\n          }\n        }\n      ]\n    }\n  ],\n  \"created_at\": RFC3339 timestamp,\n  \"updated_at\": RFC3339 timestamp\n}",
		"- Ensure board_id is \"" + strings.TrimSpace(boardID) + "\" and run_id is \"" + strings.TrimSpace(request.RunID) + "\".",
		"- Ensure all epic/task board_id values match board_id.",
		"- Required board fields: board_id, run_id, state, epics, created_at, updated_at.",
		"- Required epic fields: id, board_id, title, state, rank.",
		"- Required task fields: id, board_id, epic_id, title, task_type, state, rank.",
		"- Board state values: pending, active, completed, failed.",
		"- Epic state values: planned, in_progress, completed, blocked, failed.",
		"- Task state values: planned, in_progress, completed, failed, no_work_needed.",
		"- For outcomes, include outcome.status and outcome.summary.",
		"Remote document guidance:",
		"- Selected remote documents are optional input.",
		"- If no selected remote documents are provided, infer context from synchronized repositories and explicit user requirements.",
		decodedDocuments,
	}
	if request.PreferSelectedDocuments {
		segments = append(segments,
			"Remote document preference:",
			"- Prefer selected remote documents as the primary planning context when resolving ambiguity.",
			"- Use repository context and user requirements to complement, not override, selected remote documents unless conflicts are explicit.",
		)
	}
	if len(sourceRepositories) > 0 {
		repositoryLines := make([]string, 0, len(sourceRepositories))
		for _, repository := range sourceRepositories {
			repositoryID := strings.TrimSpace(repository.RepositoryID)
			repositoryURL := strings.TrimSpace(repository.RepositoryURL)
			repositoryBranch := strings.TrimSpace(repository.SourceBranch)
			repositoryFolder := deriveRepositoryFolderName(repositoryID, repositoryURL)
			repositoryLines = append(repositoryLines, "- repository_id="+repositoryID+" repository_url="+repositoryURL+" source_branch="+repositoryBranch+" local_dir="+filepath.Join(".", "repos", repositoryFolder))
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
		sourceBranch := strings.TrimSpace(repository.SourceBranch)
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		normalized = append(normalized, SourceRepository{RepositoryID: repositoryID, RepositoryURL: repositoryURL, SourceBranch: sourceBranch})
	}
	return normalized
}

func hasNonEmptyValues(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
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

func normalizeBoard(board *domaintracker.Board, request Request, boardID string, now time.Time, runContext AgentRunContext) {
	if board == nil {
		return
	}
	board.BoardID = strings.TrimSpace(boardID)
	board.RunID = strings.TrimSpace(request.RunID)
	board.ProjectID = strings.TrimSpace(request.ProjectID)
	requestedName := strings.TrimSpace(request.TaskboardName)
	if requestedName != "" {
		board.Name = requestedName
	} else if strings.TrimSpace(board.Name) == "" {
		board.Name = board.BoardID
	}
	if board.State == "" {
		board.State = domaintracker.BoardStatePending
	}
	if board.CreatedAt.IsZero() {
		board.CreatedAt = now
	}
	board.UpdatedAt = now

	for epicIndex := range board.Epics {
		epic := &board.Epics[epicIndex]
		if strings.TrimSpace(string(epic.ID)) == "" {
			epic.ID = domaintracker.WorkItemID(fmt.Sprintf("epic-%d", epicIndex+1))
		}
		epic.BoardID = board.BoardID
		if strings.TrimSpace(epic.Title) == "" {
			epic.Title = fmt.Sprintf("Epic %d", epicIndex+1)
		}
		if epic.State == "" {
			epic.State = domaintracker.EpicStatePlanned
		}
		epic.Rank = normalizeRank(epic.Rank, epicIndex)
		if epic.CreatedAt.IsZero() {
			epic.CreatedAt = now
		}
		epic.UpdatedAt = now

		for taskIndex := range epic.Tasks {
			task := &epic.Tasks[taskIndex]
			if strings.TrimSpace(string(task.ID)) == "" {
				task.ID = domaintracker.WorkItemID(fmt.Sprintf("task-%d-%d", epicIndex+1, taskIndex+1))
			}
			task.BoardID = board.BoardID
			task.EpicID = epic.ID
			if strings.TrimSpace(task.Title) == "" {
				task.Title = fmt.Sprintf("Task %d.%d", epicIndex+1, taskIndex+1)
			}
			if strings.TrimSpace(task.TaskType) == "" {
				task.TaskType = "implementation"
			}
			if task.State == "" {
				task.State = domaintracker.TaskStatePlanned
			}
			task.Rank = normalizeRank(task.Rank, taskIndex)
			if task.CreatedAt.IsZero() {
				task.CreatedAt = now
			}
			task.UpdatedAt = now
			if strings.TrimSpace(task.Audit.AgentSessionID) == "" {
				task.Audit.AgentSessionID = strings.TrimSpace(runContext.SessionID)
			}
			if strings.TrimSpace(task.Audit.AgentStreamID) == "" {
				task.Audit.AgentStreamID = strings.TrimSpace(runContext.StreamID)
			}
			if strings.TrimSpace(task.Audit.ModelRunID) == "" {
				task.Audit.ModelRunID = strings.TrimSpace(runContext.StreamID)
			}
			if task.Outcome != nil {
				if task.Outcome.Status == "" {
					task.Outcome.Status = domaintracker.OutcomeStatusPartial
				}
				if strings.TrimSpace(task.Outcome.Summary) == "" {
					task.Outcome.Summary = "in progress"
				}
			}
		}
	}
}

func buildRepositoryScopeMetadata(repositories []SourceRepository) []map[string]any {
	result := make([]map[string]any, 0, len(repositories))
	for _, repository := range repositories {
		repositoryID := strings.TrimSpace(repository.RepositoryID)
		repositoryURL := strings.TrimSpace(repository.RepositoryURL)
		repositoryBranch := strings.TrimSpace(repository.SourceBranch)
		if repositoryID == "" || repositoryURL == "" {
			continue
		}
		repositoryFolder := deriveRepositoryFolderName(repositoryID, repositoryURL)
		result = append(result, map[string]any{
			"repository_id":  repositoryID,
			"repository_url": repositoryURL,
			"source_branch":  repositoryBranch,
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
	_ = board
	_ = sourceRepositories
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

func normalizeRank(rank int, index int) int {
	if rank < 0 {
		return index + 1
	}
	if rank == 0 {
		return index + 1
	}
	return rank
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

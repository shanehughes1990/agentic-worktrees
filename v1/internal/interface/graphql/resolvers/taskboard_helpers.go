package resolvers

import (
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var taskboardIDSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func (r *mutationResolver) loadBoardForMutation(ctx context.Context, projectID string, boardID string) (domaintracker.Board, string, error) {
	if r == nil || r.Resolver == nil || r.Resolver.TrackerService == nil {
		return domaintracker.Board{}, "", fmt.Errorf("tracker service is not configured")
	}
	cleanProjectID := strings.TrimSpace(projectID)
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanProjectID == "" || cleanBoardID == "" {
		return domaintracker.Board{}, "", fmt.Errorf("project_id and board_id are required")
	}
	board, err := r.Resolver.TrackerService.LoadBoard(ctx, cleanProjectID, cleanBoardID)
	if err != nil {
		return domaintracker.Board{}, "", fmt.Errorf("load taskboard: %w", err)
	}
	return board, cleanProjectID, nil
}

func isBoardEnded(state domaintracker.BoardState) bool {
	return state == domaintracker.BoardStateCompleted || state == domaintracker.BoardStateFailed
}

func taskboardReadOnlyError(board domaintracker.Board) models.GraphError {
	return models.GraphError{
		Code:    models.GraphErrorCodeConflict,
		Message: fmt.Sprintf("taskboard %q is ended and read-only", strings.TrimSpace(board.BoardID)),
	}
}

func toGraphTaskboard(board domaintracker.Board) *models.Taskboard {
	epics := make([]*models.TaskboardEpic, 0, len(board.Epics))
	for _, epic := range board.Epics {
		tasks := make([]*models.TaskboardTask, 0, len(epic.Tasks))
		for _, task := range epic.Tasks {
			taskAudits := make([]*models.TaskModelAudit, 0, len(task.Audits))
			for _, audit := range task.Audits {
				taskAudits = append(taskAudits, &models.TaskModelAudit{
					ModelProvider:     strings.TrimSpace(audit.ModelProvider),
					ModelName:         strings.TrimSpace(audit.ModelName),
					ModelVersion:      nilIfEmpty(audit.ModelVersion),
					ModelRunID:        nilIfEmpty(audit.ModelRunID),
					AgentSessionID:    nilIfEmpty(audit.AgentSessionID),
					AgentStreamID:     nilIfEmpty(audit.AgentStreamID),
					PromptFingerprint: nilIfEmpty(audit.PromptFingerprint),
					InputTokens:       intPtrToInt32Ptr(audit.InputTokens),
					OutputTokens:      intPtrToInt32Ptr(audit.OutputTokens),
					StartedAt:         audit.StartedAt,
					CompletedAt:       audit.CompletedAt,
				})
			}
			tasks = append(tasks, &models.TaskboardTask{
				ID:               strings.TrimSpace(string(task.ID)),
				BoardID:          strings.TrimSpace(task.BoardID),
				EpicID:           strings.TrimSpace(string(task.EpicID)),
				Title:            strings.TrimSpace(task.Title),
				Description:      nilIfEmpty(task.Description),
				RepositoryIDs:    append([]string(nil), task.RepositoryIDs...),
				Deliverables:     append([]string(nil), task.Deliverables...),
				TaskType:         strings.TrimSpace(task.TaskType),
				State:            strings.TrimSpace(string(task.State)),
				Rank:             int32(task.Rank),
				DependsOnTaskIDs: workItemIDsToStrings(task.DependsOnTaskIDs),
				Audits:           taskAudits,
			})
		}
		epics = append(epics, &models.TaskboardEpic{
			ID:               strings.TrimSpace(string(epic.ID)),
			BoardID:          strings.TrimSpace(epic.BoardID),
			Title:            strings.TrimSpace(epic.Title),
			Objective:        nilIfEmpty(epic.Objective),
			RepositoryIDs:    append([]string(nil), epic.RepositoryIDs...),
			Deliverables:     append([]string(nil), epic.Deliverables...),
			State:            strings.TrimSpace(string(epic.State)),
			Rank:             int32(epic.Rank),
			DependsOnEpicIDs: workItemIDsToStrings(epic.DependsOnEpicIDs),
			Tasks:            tasks,
		})
	}
	ingestionAudits := make([]*models.TaskModelAudit, 0, len(board.IngestionAudits))
	for _, audit := range board.IngestionAudits {
		ingestionAudits = append(ingestionAudits, &models.TaskModelAudit{
			ModelProvider:     strings.TrimSpace(audit.ModelProvider),
			ModelName:         strings.TrimSpace(audit.ModelName),
			ModelVersion:      nilIfEmpty(audit.ModelVersion),
			ModelRunID:        nilIfEmpty(audit.ModelRunID),
			AgentSessionID:    nilIfEmpty(audit.AgentSessionID),
			AgentStreamID:     nilIfEmpty(audit.AgentStreamID),
			PromptFingerprint: nilIfEmpty(audit.PromptFingerprint),
			InputTokens:       intPtrToInt32Ptr(audit.InputTokens),
			OutputTokens:      intPtrToInt32Ptr(audit.OutputTokens),
			StartedAt:         audit.StartedAt,
			CompletedAt:       audit.CompletedAt,
		})
	}
	name := strings.TrimSpace(board.Name)
	if name == "" {
		name = strings.TrimSpace(board.BoardID)
	}
	var ingestionDetails *models.TaskboardIngestionDetails
	if board.IngestionDetails != nil {
		ingestionDetails = &models.TaskboardIngestionDetails{
			FilesAdded: append([]string(nil), sanitizeStringList(board.IngestionDetails.FilesAdded)...),
			UserPrompt: strings.TrimSpace(board.IngestionDetails.UserPrompt),
		}
	}
	return &models.Taskboard{
		BoardID:         strings.TrimSpace(board.BoardID),
		ProjectID:       strings.TrimSpace(board.ProjectID),
		Name:            name,
		State:           strings.TrimSpace(string(board.State)),
		Epics:           epics,
		IngestionAudits: ingestionAudits,
		IngestionDetails: ingestionDetails,
		CreatedAt:       board.CreatedAt.UTC(),
		UpdatedAt:       board.UpdatedAt.UTC(),
	}
}

func intPtrToInt32Ptr(value *int) *int32 {
	if value == nil {
		return nil
	}
	converted := int32(*value)
	return &converted
}

func parseBoardState(value string) (domaintracker.BoardState, error) {
	state := domaintracker.BoardState(strings.TrimSpace(value))
	if err := state.Validate(); err != nil {
		return "", err
	}
	return state, nil
}

func parseEpicState(value string) (domaintracker.EpicState, error) {
	state := domaintracker.EpicState(strings.TrimSpace(value))
	if err := state.Validate(); err != nil {
		return "", err
	}
	return state, nil
}

func parseTaskState(value string) (domaintracker.TaskState, error) {
	state := domaintracker.TaskState(strings.TrimSpace(value))
	if err := state.Validate(); err != nil {
		return "", err
	}
	return state, nil
}

func newTaskboardID(name string) string {
	base := sanitizeName(name)
	if base == "" {
		base = "taskboard"
	}
	return fmt.Sprintf("%s_%d", base, time.Now().UTC().UnixNano())
}

func newEpicID(title string) string {
	base := sanitizeName(title)
	if base == "" {
		base = "epic"
	}
	return fmt.Sprintf("%s_%d", base, time.Now().UTC().UnixNano())
}

func newTaskID(title string) string {
	base := sanitizeName(title)
	if base == "" {
		base = "task"
	}
	return fmt.Sprintf("%s_%d", base, time.Now().UTC().UnixNano())
}

func sanitizeName(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}
	cleaned := taskboardIDSanitizer.ReplaceAllString(trimmed, "_")
	return strings.Trim(cleaned, "_")
}

func workItemIDsToStrings(values []domaintracker.WorkItemID) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		clean := strings.TrimSpace(string(value))
		if clean == "" {
			continue
		}
		items = append(items, clean)
	}
	return items
}

func stringsToWorkItemIDs(values []string) []domaintracker.WorkItemID {
	items := make([]domaintracker.WorkItemID, 0, len(values))
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" {
			continue
		}
		items = append(items, domaintracker.WorkItemID(clean))
	}
	return items
}

func derefStrings(value []string) []string {
	return value
}

func sanitizeStringList(values []string) []string {
	items := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		items = append(items, clean)
	}
	return items
}

func inferBoardRepositoryIDs(board domaintracker.Board) []string {
	seen := map[string]struct{}{}
	items := make([]string, 0)
	for _, epic := range board.Epics {
		for _, repositoryID := range sanitizeStringList(epic.RepositoryIDs) {
			if _, exists := seen[repositoryID]; exists {
				continue
			}
			seen[repositoryID] = struct{}{}
			items = append(items, repositoryID)
		}
		for _, task := range epic.Tasks {
			for _, repositoryID := range sanitizeStringList(task.RepositoryIDs) {
				if _, exists := seen[repositoryID]; exists {
					continue
				}
				seen[repositoryID] = struct{}{}
				items = append(items, repositoryID)
			}
		}
	}
	return items
}

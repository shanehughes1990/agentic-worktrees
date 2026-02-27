package jsonrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type Repository struct {
	taskboardsDirectory string
	workflowsDirectory  string
}

func NewRepository(rootDirectory string) (*Repository, error) {
	return NewRepositoryWithWorkflowDirectory(rootDirectory, rootDirectory)
}

func NewRepositoryWithWorkflowDirectory(taskboardsDirectory string, workflowsDirectory string) (*Repository, error) {
	cleanTaskboardsDirectory := strings.TrimSpace(taskboardsDirectory)
	if cleanTaskboardsDirectory == "" {
		return nil, fmt.Errorf("taskboards directory is required")
	}
	cleanWorkflowsDirectory := strings.TrimSpace(workflowsDirectory)
	if cleanWorkflowsDirectory == "" {
		return nil, fmt.Errorf("workflows directory is required")
	}

	if err := os.MkdirAll(cleanTaskboardsDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create taskboard directory: %w", err)
	}
	if err := os.MkdirAll(cleanWorkflowsDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create workflow directory: %w", err)
	}

	return &Repository{taskboardsDirectory: cleanTaskboardsDirectory, workflowsDirectory: cleanWorkflowsDirectory}, nil
}

func (repository *Repository) taskboardsDir() string {
	if repository == nil {
		return ""
	}
	return strings.TrimSpace(repository.taskboardsDirectory)
}

func (repository *Repository) workflowsDir() string {
	if repository == nil {
		return ""
}
	cleanWorkflowsDirectory := strings.TrimSpace(repository.workflowsDirectory)
	if cleanWorkflowsDirectory != "" {
		return cleanWorkflowsDirectory
}
	return strings.TrimSpace(repository.taskboardsDirectory)
}

func (repository *Repository) legacyWorkflowsDir() string {
	if repository == nil {
		return ""
}
	cleanTaskboardsDirectory := strings.TrimSpace(repository.taskboardsDirectory)
	cleanWorkflowsDirectory := strings.TrimSpace(repository.workflowsDirectory)
	if cleanTaskboardsDirectory == "" || cleanTaskboardsDirectory == cleanWorkflowsDirectory {
		return ""
	}
	return cleanTaskboardsDirectory
}

func (repository *Repository) GetByBoardID(ctx context.Context, boardID string) (*domaintaskboard.Board, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath, err := repository.filePathForBoard(boardID)
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read board file: %w", err)
	}

	board := &domaintaskboard.Board{}
	if err := json.Unmarshal(payload, board); err != nil {
		return nil, fmt.Errorf("unmarshal board json: %w", err)
	}

	if err := board.ValidateBasics(); err != nil {
		return nil, fmt.Errorf("invalid board data: %w", err)
	}
	if err := board.ValidateComplete(); err != nil {
		return nil, fmt.Errorf("incomplete board data: %w", err)
	}

	return board, nil
}

func (repository *Repository) ListBoardIDs(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(repository.taskboardsDir())
	if err != nil {
		return nil, fmt.Errorf("read taskboard directory: %w", err)
	}

	boardIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" || !strings.HasSuffix(name, ".json") {
			continue
		}
		if strings.HasPrefix(name, "workflow-") || strings.HasPrefix(name, "run-") || strings.HasPrefix(name, "job-") {
			continue
		}
		boardID := strings.TrimSuffix(name, ".json")
		if strings.TrimSpace(boardID) == "" {
			continue
		}
		boardIDs = append(boardIDs, boardID)
	}

	sort.Strings(boardIDs)
	return boardIDs, nil
}

func (repository *Repository) Save(ctx context.Context, board *domaintaskboard.Board) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if board == nil {
		return fmt.Errorf("board is required")
	}
	if err := board.ValidateBasics(); err != nil {
		return err
	}
	if err := board.ValidateComplete(); err != nil {
		return err
	}

	filePath, err := repository.filePathForBoard(board.BoardID)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal board json: %w", err)
	}

	if err := writeAtomically(filePath, payload); err != nil {
		return fmt.Errorf("write board file: %w", err)
	}

	return nil
}

func (repository *Repository) GetWorkflow(ctx context.Context, runID string) (*apptaskboard.IngestionWorkflow, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath, err := repository.filePathForWorkflow(runID)
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			legacyDirectory := repository.legacyWorkflowsDir()
			if legacyDirectory == "" {
				return nil, nil
			}
			legacyPath := filepath.Join(legacyDirectory, "workflow-"+strings.TrimSpace(runID)+".json")
			legacyPayload, legacyErr := os.ReadFile(legacyPath)
			if legacyErr != nil {
				if os.IsNotExist(legacyErr) {
					return nil, nil
				}
				return nil, fmt.Errorf("read workflow file: %w", legacyErr)
			}
			payload = legacyPayload
		} else {
			return nil, fmt.Errorf("read workflow file: %w", err)
		}
	}

	workflow := &apptaskboard.IngestionWorkflow{}
	if err := json.Unmarshal(payload, workflow); err != nil {
		return nil, fmt.Errorf("unmarshal workflow json: %w", err)
	}
	return workflow, nil
}

func (repository *Repository) ListWorkflows(ctx context.Context) ([]apptaskboard.IngestionWorkflow, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	pattern := filepath.Join(repository.workflowsDir(), "workflow-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("list workflow files: %w", err)
	}
	legacyPattern := ""
	if legacyDirectory := repository.legacyWorkflowsDir(); legacyDirectory != "" {
		legacyPattern = filepath.Join(legacyDirectory, "workflow-*.json")
		legacyMatches, legacyErr := filepath.Glob(legacyPattern)
		if legacyErr == nil {
			matches = append(matches, legacyMatches...)
		}
	}

	workflowsByRunID := make(map[string]apptaskboard.IngestionWorkflow, len(matches))
	for _, filePath := range matches {
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read workflow file %s: %w", filePath, err)
		}
		workflow := apptaskboard.IngestionWorkflow{}
		if err := json.Unmarshal(payload, &workflow); err != nil {
			return nil, fmt.Errorf("unmarshal workflow file %s: %w", filePath, err)
		}
		existingWorkflow, exists := workflowsByRunID[workflow.RunID]
		if !exists || workflow.UpdatedAt.After(existingWorkflow.UpdatedAt) {
			workflowsByRunID[workflow.RunID] = workflow
		}
	}

	workflows := make([]apptaskboard.IngestionWorkflow, 0, len(workflowsByRunID))
	for _, workflow := range workflowsByRunID {
		workflows = append(workflows, workflow)
	}

	sort.SliceStable(workflows, func(left, right int) bool {
		return workflows[left].UpdatedAt.After(workflows[right].UpdatedAt)
	})

	return workflows, nil
}

func (repository *Repository) SaveWorkflow(ctx context.Context, workflow *apptaskboard.IngestionWorkflow) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if workflow == nil {
		return fmt.Errorf("workflow is required")
	}
	cleanRunID := strings.TrimSpace(workflow.RunID)
	if cleanRunID == "" {
		return fmt.Errorf("run_id is required")
	}

	filePath, err := repository.filePathForWorkflow(cleanRunID)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow json: %w", err)
	}

	if err := writeAtomically(filePath, payload); err != nil {
		return fmt.Errorf("write workflow file: %w", err)
	}
	return nil
}

func (repository *Repository) GetRunState(ctx context.Context, runID string) (*apptaskboard.RunState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath, err := repository.filePathForRunState(runID)
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read run file: %w", err)
	}

	runState := &apptaskboard.RunState{}
	if err := json.Unmarshal(payload, runState); err != nil {
		return nil, fmt.Errorf("unmarshal run json: %w", err)
	}
	return runState, nil
}

func (repository *Repository) ListRunStates(ctx context.Context) ([]apptaskboard.RunState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(repository.workflowsDir())
	if err != nil {
		return nil, fmt.Errorf("read taskboard directory: %w", err)
	}

	runStates := make([]apptaskboard.RunState, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !strings.HasPrefix(name, "run-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		filePath := filepath.Join(repository.workflowsDir(), name)
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read run file %s: %w", filePath, err)
		}
		runState := apptaskboard.RunState{}
		if err := json.Unmarshal(payload, &runState); err != nil {
			return nil, fmt.Errorf("unmarshal run file %s: %w", filePath, err)
		}
		runStates = append(runStates, runState)
	}

	sort.SliceStable(runStates, func(left, right int) bool {
		if runStates[left].UpdatedAt.Equal(runStates[right].UpdatedAt) {
			return runStates[left].RunID > runStates[right].RunID
		}
		return runStates[left].UpdatedAt.After(runStates[right].UpdatedAt)
	})

	return runStates, nil
}

func (repository *Repository) SaveRunState(ctx context.Context, runState *apptaskboard.RunState) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if runState == nil {
		return fmt.Errorf("run state is required")
	}
	cleanRunID := strings.TrimSpace(runState.RunID)
	if cleanRunID == "" {
		return fmt.Errorf("run_id is required")
	}
	runState.Normalize(cleanRunID)

	filePath, err := repository.filePathForRunState(cleanRunID)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(runState, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run json: %w", err)
	}

	if err := writeAtomically(filePath, payload); err != nil {
		return fmt.Errorf("write run file: %w", err)
	}
	return nil
}

func (repository *Repository) GetJobState(ctx context.Context, runID string, jobID string) (*apptaskboard.JobState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath, err := repository.filePathForJobState(runID, jobID)
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read job file: %w", err)
	}

	jobState := &apptaskboard.JobState{}
	if err := json.Unmarshal(payload, jobState); err != nil {
		return nil, fmt.Errorf("unmarshal job json: %w", err)
	}
	return jobState, nil
}

func (repository *Repository) ListJobStatesByRunID(ctx context.Context, runID string) ([]apptaskboard.JobState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	if strings.Contains(cleanRunID, "/") || strings.Contains(cleanRunID, "\\") {
		return nil, fmt.Errorf("run_id cannot contain path separators")
	}

	entries, err := os.ReadDir(repository.workflowsDir())
	if err != nil {
		return nil, fmt.Errorf("read taskboard directory: %w", err)
	}

	namePrefix := "job-" + cleanRunID + "-"
	jobStates := make([]apptaskboard.JobState, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if !strings.HasPrefix(name, namePrefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		filePath := filepath.Join(repository.workflowsDir(), name)
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read job file %s: %w", filePath, err)
		}
		jobState := apptaskboard.JobState{}
		if err := json.Unmarshal(payload, &jobState); err != nil {
			return nil, fmt.Errorf("unmarshal job file %s: %w", filePath, err)
		}
		jobStates = append(jobStates, jobState)
	}

	sort.SliceStable(jobStates, func(left, right int) bool {
		if jobStates[left].UpdatedAt.Equal(jobStates[right].UpdatedAt) {
			return jobStates[left].JobID > jobStates[right].JobID
		}
		return jobStates[left].UpdatedAt.After(jobStates[right].UpdatedAt)
	})

	return jobStates, nil
}

func (repository *Repository) SaveJobState(ctx context.Context, jobState *apptaskboard.JobState) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if jobState == nil {
		return fmt.Errorf("job state is required")
	}
	cleanRunID := strings.TrimSpace(jobState.RunID)
	if cleanRunID == "" {
		return fmt.Errorf("run_id is required")
	}
	cleanJobID := strings.TrimSpace(jobState.JobID)
	if cleanJobID == "" {
		return fmt.Errorf("job_id is required")
	}
	jobState.Normalize(cleanRunID, cleanJobID)

	filePath, err := repository.filePathForJobState(cleanRunID, cleanJobID)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(jobState, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal job json: %w", err)
	}

	if err := writeAtomically(filePath, payload); err != nil {
		return fmt.Errorf("write job file: %w", err)
	}
	return nil
}

func (repository *Repository) filePathForBoard(boardID string) (string, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return "", fmt.Errorf("board_id is required")
	}
	if strings.Contains(cleanBoardID, "/") || strings.Contains(cleanBoardID, "\\") {
		return "", fmt.Errorf("board_id cannot contain path separators")
	}
	return filepath.Join(repository.taskboardsDir(), cleanBoardID+".json"), nil
}

func (repository *Repository) filePathForWorkflow(runID string) (string, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	if strings.Contains(cleanRunID, "/") || strings.Contains(cleanRunID, "\\") {
		return "", fmt.Errorf("run_id cannot contain path separators")
	}
	return filepath.Join(repository.workflowsDir(), "workflow-"+cleanRunID+".json"), nil
}

func (repository *Repository) filePathForRunState(runID string) (string, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	if strings.Contains(cleanRunID, "/") || strings.Contains(cleanRunID, "\\") {
		return "", fmt.Errorf("run_id cannot contain path separators")
	}
	return filepath.Join(repository.workflowsDir(), "run-"+cleanRunID+".json"), nil
}

func (repository *Repository) filePathForJobState(runID string, jobID string) (string, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	if strings.Contains(cleanRunID, "/") || strings.Contains(cleanRunID, "\\") {
		return "", fmt.Errorf("run_id cannot contain path separators")
	}

	cleanJobID := strings.TrimSpace(jobID)
	if cleanJobID == "" {
		return "", fmt.Errorf("job_id is required")
	}
	if strings.Contains(cleanJobID, "/") || strings.Contains(cleanJobID, "\\") {
		return "", fmt.Errorf("job_id cannot contain path separators")
	}

	return filepath.Join(repository.workflowsDir(), "job-"+cleanRunID+"-"+cleanJobID+".json"), nil
}

func writeAtomically(targetPath string, payload []byte) error {
	temporaryPath := targetPath + ".tmp"
	if err := os.WriteFile(temporaryPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(temporaryPath, targetPath)
}

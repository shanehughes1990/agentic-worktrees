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
	rootDirectory string
}

func NewRepository(rootDirectory string) (*Repository, error) {
	cleanDirectory := strings.TrimSpace(rootDirectory)
	if cleanDirectory == "" {
		return nil, fmt.Errorf("root directory is required")
	}

	if err := os.MkdirAll(cleanDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("create taskboard directory: %w", err)
	}

	return &Repository{rootDirectory: cleanDirectory}, nil
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
			return nil, nil
		}
		return nil, fmt.Errorf("read workflow file: %w", err)
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

	pattern := filepath.Join(repository.rootDirectory, "workflow-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("list workflow files: %w", err)
	}

	workflows := make([]apptaskboard.IngestionWorkflow, 0, len(matches))
	for _, filePath := range matches {
		payload, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read workflow file %s: %w", filePath, err)
		}
		workflow := apptaskboard.IngestionWorkflow{}
		if err := json.Unmarshal(payload, &workflow); err != nil {
			return nil, fmt.Errorf("unmarshal workflow file %s: %w", filePath, err)
		}
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

func (repository *Repository) filePathForBoard(boardID string) (string, error) {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return "", fmt.Errorf("board_id is required")
	}
	if strings.Contains(cleanBoardID, "/") || strings.Contains(cleanBoardID, "\\") {
		return "", fmt.Errorf("board_id cannot contain path separators")
	}
	return filepath.Join(repository.rootDirectory, cleanBoardID+".json"), nil
}

func (repository *Repository) filePathForWorkflow(runID string) (string, error) {
	cleanRunID := strings.TrimSpace(runID)
	if cleanRunID == "" {
		return "", fmt.Errorf("run_id is required")
	}
	if strings.Contains(cleanRunID, "/") || strings.Contains(cleanRunID, "\\") {
		return "", fmt.Errorf("run_id cannot contain path separators")
	}
	return filepath.Join(repository.rootDirectory, "workflow-"+cleanRunID+".json"), nil
}

func writeAtomically(targetPath string, payload []byte) error {
	temporaryPath := targetPath + ".tmp"
	if err := os.WriteFile(temporaryPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(temporaryPath, targetPath)
}

package jsonrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	filePath, err := repository.filePathForBoard(board.BoardID)
	if err != nil {
		return err
	}

	payload, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal board json: %w", err)
	}

	if err := os.WriteFile(filePath, payload, 0o644); err != nil {
		return fmt.Errorf("write board file: %w", err)
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

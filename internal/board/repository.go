package board

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shanehughes1990/agentic-worktrees/internal/domain"
)

type Repository struct {
	path string
}

func NewRepository(path string) *Repository {
	return &Repository{path: path}
}

func (r *Repository) Path() string {
	return r.path
}

func (r *Repository) Write(board domain.Board) error {
	board = normalize(board)

	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create board dir: %w", err)
	}

	payload, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	payload = append(payload, '\n')

	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return fmt.Errorf("write tmp board file: %w", err)
	}
	if err := os.Rename(tmpPath, r.path); err != nil {
		return fmt.Errorf("rename board file: %w", err)
	}
	return nil
}

func (r *Repository) Read() (domain.Board, error) {
	payload, err := os.ReadFile(r.path)
	if err != nil {
		return domain.Board{}, fmt.Errorf("read board file: %w", err)
	}
	var board domain.Board
	if err := json.Unmarshal(payload, &board); err != nil {
		return domain.Board{}, fmt.Errorf("unmarshal board: %w", err)
	}
	board = normalize(board)
	if board.SchemaVersion > domain.CurrentSchemaVersion {
		return domain.Board{}, fmt.Errorf("unsupported board schema version %d", board.SchemaVersion)
	}
	return board, nil
}

func normalize(board domain.Board) domain.Board {
	if board.SchemaVersion <= 0 {
		board.SchemaVersion = domain.CurrentSchemaVersion
	}
	if board.Tasks == nil {
		board.Tasks = make([]domain.Task, 0)
	}
	for index := range board.Tasks {
		if board.Tasks[index].Status == "" {
			board.Tasks[index].Status = domain.TaskStatusPending
		}
		if board.Tasks[index].Lane == "" {
			board.Tasks[index].Lane = "lane-a"
		}
	}
	return board
}

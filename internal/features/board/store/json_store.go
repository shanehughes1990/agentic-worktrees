package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
)

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (s *JSONStore) Write(board boarddomain.Board) error {
	if err := board.Validate(); err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create board dir: %w", err)
	}

	payload, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	payload = append(payload, '\n')

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return fmt.Errorf("write board tmp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename board tmp file: %w", err)
	}
	return nil
}

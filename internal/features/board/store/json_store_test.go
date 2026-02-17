package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
)

func TestWriteBoard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.json")
	store := NewJSONStore(path)

	board := boarddomain.Board{
		SchemaVersion: 1,
		SourceScope:   "docs",
		GeneratedAt:   time.Now().UTC(),
		Epics: []boarddomain.Epic{{
			ID:    "epic-001",
			Title: "first",
			Tasks: []boarddomain.Task{{
				ID:     "task-001",
				Title:  "task",
				Lane:   "lane-a",
				Status: boarddomain.TaskStatusPending,
			}},
		}},
	}

	if err := store.Write(board); err != nil {
		t.Fatalf("write board: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected board file written: %v", err)
	}
}

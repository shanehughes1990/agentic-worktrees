package board

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shanehughes1990/agentic-worktrees/internal/domain"
)

func TestRepositoryWriteReadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "board.json")
	repo := NewRepository(path)

	input := domain.Board{
		SchemaVersion: 1,
		SourceScope:   "scope.md",
		Tasks: []domain.Task{{
			ID:     "task-001",
			Title:  "Do thing",
			Lane:   "lane-a",
			Status: domain.TaskStatusPending,
		}},
	}

	if err := repo.Write(input); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output, err := repo.Read()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if output.SchemaVersion != 1 {
		t.Fatalf("unexpected schema version: %d", output.SchemaVersion)
	}
	if len(output.Tasks) != 1 {
		t.Fatalf("expected one task, got %d", len(output.Tasks))
	}
}

func TestRepositoryNormalizeOlderSchema(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "board.json")
	payload := []byte(`{"schema_version":0,"tasks":[{"id":"task-1","title":"x"}]}`)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	repo := NewRepository(path)
	board, err := repo.Read()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if board.SchemaVersion != 1 {
		t.Fatalf("expected normalized schema version 1, got %d", board.SchemaVersion)
	}
	if board.Tasks[0].Status == "" {
		t.Fatalf("expected normalized status")
	}
}

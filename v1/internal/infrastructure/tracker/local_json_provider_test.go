package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalJSONProviderSyncBoardReadsCanonicalBoard(t *testing.T) {
	root := t.TempDir()
	boardPath := filepath.Join(root, "board-1.json")
	payload := []byte(`{
  "board_id": "board-1",
  "run_id": "run-1",
  "status": "in-progress",
  "epics": [
    {
      "id": "epic-1",
      "board_id": "board-1",
      "title": "Bootstrap",
      "status": "in-progress",
      "tasks": [
        {
          "id": "task-1",
          "board_id": "board-1",
          "title": "Read board",
          "status": "in-progress"
        }
      ]
    }
  ],
  "created_at": "2026-02-25T03:15:00Z",
  "updated_at": "2026-02-25T03:30:00Z"
}`)
	if err := os.WriteFile(boardPath, payload, 0o644); err != nil {
		t.Fatalf("write board file: %v", err)
	}

	provider, err := NewLocalJSONProvider(root)
	if err != nil {
		t.Fatalf("new local json provider: %v", err)
	}

	board, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindLocalJSON,
			Location: "board-1.json",
		},
	})
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if board.BoardID != "board-1" {
		t.Fatalf("expected board id board-1, got %q", board.BoardID)
	}
	if board.Source.Kind != domaintracker.SourceKindLocalJSON {
		t.Fatalf("expected source kind local_json, got %q", board.Source.Kind)
	}
	if err := board.Validate(); err != nil {
		t.Fatalf("expected canonical board validation, got %v", err)
	}
}

func TestLocalJSONProviderSyncBoardResolvesRelativePathFromBaseDirectory(t *testing.T) {
	root := t.TempDir()
	taskboardDirectory := filepath.Join(root, "taskboards")
	if err := os.MkdirAll(taskboardDirectory, 0o755); err != nil {
		t.Fatalf("mkdir taskboards: %v", err)
	}
	boardPath := filepath.Join(taskboardDirectory, "board-1.json")
	payload := []byte(`{
  "board_id": "board-1",
  "run_id": "run-1",
  "status": "completed",
  "epics": [
    {
      "id": "epic-1",
      "board_id": "board-1",
      "title": "Done",
      "status": "completed",
      "tasks": [
        {
          "id": "task-1",
          "board_id": "board-1",
          "title": "Done",
          "status": "completed"
        }
      ]
    }
  ],
  "created_at": "2026-02-25T03:15:00Z",
  "updated_at": "2026-02-25T03:30:00Z"
}`)
	if err := os.WriteFile(boardPath, payload, 0o644); err != nil {
		t.Fatalf("write board file: %v", err)
	}

	provider, err := NewLocalJSONProvider(root)
	if err != nil {
		t.Fatalf("new local json provider: %v", err)
	}

	board, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindLocalJSON,
			Location: filepath.Join("taskboards", "board-1.json"),
		},
	})
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if board.UpdatedAt.Before(time.Date(2026, time.February, 25, 3, 30, 0, 0, time.UTC)) {
		t.Fatalf("expected parsed updated_at timestamp, got %v", board.UpdatedAt)
	}
}

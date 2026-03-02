package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestPostgresInternalProviderLoadsBoardFromSnapshot(t *testing.T) {
	db := newTrackerDB(t)
	provider, err := NewPostgresInternalProvider(db)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	now := time.Now().UTC()
	seedBoard := domaintracker.Board{
		BoardID: "board-1",
		RunID:   "project-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-1",
			BoardID:  "board-1",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{{
				WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintracker.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	payload, err := json.Marshal(seedBoard)
	if err != nil {
		t.Fatalf("marshal seed board: %v", err)
	}
	if err := db.Create(&boardSnapshotRecord{RunID: "project-1", BoardID: "board-1", SourceKind: "internal", SourceRef: "board-1", Payload: payload}).Error; err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	board, err := provider.SyncBoard(context.Background(), applicationtracker.ProviderSyncRequest{
		RunID:      "run-1",
		ProjectID:  "project-1",
		WorkflowID: "workflow-1",
		Source:     domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1", BoardID: "board-1"},
	})
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if board.RunID != "run-1" {
		t.Fatalf("expected run_id run-1, got %q", board.RunID)
	}
	if board.Source.Kind != domaintracker.SourceKindInternal {
		t.Fatalf("expected internal source kind, got %q", board.Source.Kind)
	}
}

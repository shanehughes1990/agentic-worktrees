package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeUpstreamProvider struct {
	board domaintracker.Board
	err   error
}

func (provider *fakeUpstreamProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	_ = ctx
	_ = request
	if provider.err != nil {
		return domaintracker.Board{}, provider.err
	}
	return provider.board, nil
}

func newTrackerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func sampleBoard() domaintracker.Board {
	now := time.Now().UTC()
	return domaintracker.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindLocalJSON,
			Location: "board-1.json",
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
}

func TestPostgresBoardSnapshotProviderPersistsBoard(t *testing.T) {
	db := newTrackerDB(t)
	provider, err := NewPostgresBoardSnapshotProvider(db, &fakeUpstreamProvider{board: sampleBoard()})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	request := applicationtracker.ProviderSyncRequest{RunID: "run-1", ProjectID: "project-1", WorkflowID: "workflow-1", Source: domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"}}
	board, err := provider.SyncBoard(context.Background(), request)
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if board.BoardID != "board-1" {
		t.Fatalf("unexpected board id %q", board.BoardID)
	}
	var count int64
	if err := db.Model(&boardSnapshotRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&count).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 snapshot, got %d", count)
	}
}

func TestPostgresBoardSnapshotProviderPropagatesUpstreamError(t *testing.T) {
	db := newTrackerDB(t)
	provider, err := NewPostgresBoardSnapshotProvider(db, &fakeUpstreamProvider{err: errors.New("upstream failed")})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	request := applicationtracker.ProviderSyncRequest{RunID: "run-1", ProjectID: "project-1", WorkflowID: "workflow-1", Source: domaintracker.SourceRef{Kind: domaintracker.SourceKindLocalJSON, Location: "board-1.json"}}
	if _, err := provider.SyncBoard(context.Background(), request); err == nil {
		t.Fatalf("expected upstream error")
	}
}

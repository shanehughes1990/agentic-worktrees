package tracker

import (
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTrackerDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func sampleCanonicalBoard() domaintracker.Board {
	now := time.Now().UTC()
	return domaintracker.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindGitHubIssues,
			Location: "octo/repo",
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

func TestPostgresTaskboardStoreUpsertBoardPersistsSnapshotAndNormalizedTables(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresTaskboardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	var snapshotCount int64
	if err := db.Model(&boardSnapshotRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&snapshotCount).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("expected 1 snapshot record, got %d", snapshotCount)
	}

	var boardCount int64
	if err := db.Model(&trackerBoardRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&boardCount).Error; err != nil {
		t.Fatalf("count tracker boards: %v", err)
	}
	if boardCount != 1 {
		t.Fatalf("expected 1 tracker board record, got %d", boardCount)
	}

	var taskCount int64
	if err := db.Model(&trackerTaskRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&taskCount).Error; err != nil {
		t.Fatalf("count tracker tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("expected 1 tracker task record, got %d", taskCount)
	}
}

func TestPostgresTaskboardStoreLoadBoardFromSnapshot(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresTaskboardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	board.RunID = "project-1"
	board.Source.Kind = domaintracker.SourceKindInternal
	board.Source.Location = "board-1"
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	loaded, err := store.LoadBoard(context.Background(), "project-1", "board-1")
	if err != nil {
		t.Fatalf("load board: %v", err)
	}
	if loaded.BoardID != "board-1" {
		t.Fatalf("expected board id board-1, got %q", loaded.BoardID)
	}
	if loaded.Source.Kind != domaintracker.SourceKindInternal {
		t.Fatalf("expected internal source, got %q", loaded.Source.Kind)
	}
}

func TestPostgresTaskboardStoreAllowsBlankInitialBoard(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresTaskboardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	now := time.Now().UTC()
	board := domaintracker.Board{
		BoardID: "board-seed",
		RunID:   "project-seed",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-seed",
			BoardID:  "board-seed",
		},
		Status:    domaintracker.StatusNotStarted,
		Epics:     []domaintracker.Epic{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert blank board: %v", err)
	}

	loaded, err := store.LoadBoard(context.Background(), "project-seed", "board-seed")
	if err != nil {
		t.Fatalf("load blank board: %v", err)
	}
	if len(loaded.Epics) != 0 {
		t.Fatalf("expected no epics, got %d", len(loaded.Epics))
	}
}

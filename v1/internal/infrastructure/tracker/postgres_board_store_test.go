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
		Name:    "Board",
		State:   domaintracker.BoardStateActive,
		Epics: []domaintracker.Epic{{
			ID:      domaintracker.WorkItemID("epic-1"),
			BoardID: "board-1",
			Title:   "Epic",
			State:   domaintracker.EpicStateInProgress,
			Rank:    1,
			Tasks: []domaintracker.Task{{
				ID:       domaintracker.WorkItemID("task-1"),
				BoardID:  "board-1",
				EpicID:   domaintracker.WorkItemID("epic-1"),
				Title:    "Task",
				TaskType: "implementation",
				State:    domaintracker.TaskStatePlanned,
				Rank:     1,
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPostgresBoardStoreUpsertAndLoadBoard(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	loaded, err := store.LoadBoard(context.Background(), "run-1", "board-1")
	if err != nil {
		t.Fatalf("load board: %v", err)
	}
	if loaded.BoardID != "board-1" {
		t.Fatalf("expected board id board-1, got %q", loaded.BoardID)
	}
	if loaded.State != domaintracker.BoardStateActive {
		t.Fatalf("expected active board state, got %q", loaded.State)
	}
	if len(loaded.Epics) != 1 || len(loaded.Epics[0].Tasks) != 1 {
		t.Fatalf("expected one epic/one task, got %d epics and %d tasks", len(loaded.Epics), len(loaded.Epics[0].Tasks))
	}
}

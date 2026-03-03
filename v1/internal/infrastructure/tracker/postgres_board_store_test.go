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
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-1",
			BoardID:  "board-1",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{ {
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{ {
				WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintracker.StatusInProgress},
			} },
		} },
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sampleBoardWithTwoTasks() domaintracker.Board {
	now := time.Now().UTC()
	return domaintracker.Board{
		BoardID: "board-2",
		RunID:   "project-2",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-2",
			BoardID:  "board-2",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{ {
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-2", Title: "Epic", Status: domaintracker.StatusInProgress},
			Tasks: []domaintracker.Task{
				{WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-2", Title: "Task 1", Status: domaintracker.StatusNotStarted}},
				{WorkItem: domaintracker.WorkItem{ID: "task-2", BoardID: "board-2", Title: "Task 2", Status: domaintracker.StatusNotStarted}},
			},
		} },
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPostgresBoardStoreUpsertBoardPersistsSnapshot(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
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
}

func TestPostgresBoardStoreLoadBoardFromSnapshot(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
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

func TestPostgresBoardStoreAllowsBlankInitialBoard(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
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

func TestPostgresBoardStoreClaimNextTaskClaimsDistinctTasks(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleBoardWithTwoTasks()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	_, taskOne, claimOne, revisionOne, err := store.ClaimNextTask(context.Background(), "project-2", "board-2", "worker-1")
	if err != nil {
		t.Fatalf("claim next task one: %v", err)
	}
	_, taskTwo, claimTwo, revisionTwo, err := store.ClaimNextTask(context.Background(), "project-2", "board-2", "worker-2")
	if err != nil {
		t.Fatalf("claim next task two: %v", err)
	}

	if string(taskOne.ID) == string(taskTwo.ID) {
		t.Fatalf("expected distinct claimed tasks, got %q and %q", taskOne.ID, taskTwo.ID)
	}
	if claimOne == claimTwo || claimOne == "" || claimTwo == "" {
		t.Fatalf("expected unique non-empty claim ids, got %q and %q", claimOne, claimTwo)
	}
	if revisionTwo <= revisionOne {
		t.Fatalf("expected increasing snapshot revision, got %d then %d", revisionOne, revisionTwo)
	}
}

func TestPostgresBoardStoreApplyTaskResultUpdatesOnlyClaimedTask(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleBoardWithTwoTasks()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	_, claimedTask, claimID, _, err := store.ClaimNextTask(context.Background(), "project-2", "board-2", "worker-1")
	if err != nil {
		t.Fatalf("claim next task: %v", err)
	}

	updatedBoard, revision, err := store.ApplyTaskResult(
		context.Background(),
		"project-2",
		"board-2",
		claimID,
		string(claimedTask.ID),
		domaintracker.StatusCompleted,
		domaintracker.TaskOutcome{Status: "completed", Reason: "merged", UpdatedAt: time.Now().UTC()},
	)
	if err != nil {
		t.Fatalf("apply task result: %v", err)
	}
	if revision <= 0 {
		t.Fatalf("expected positive committed revision, got %d", revision)
	}

	statusByTaskID := map[string]domaintracker.Status{}
	for _, epic := range updatedBoard.Epics {
		for _, task := range epic.Tasks {
			statusByTaskID[string(task.ID)] = task.Status
		}
	}
	if statusByTaskID[string(claimedTask.ID)] != domaintracker.StatusCompleted {
		t.Fatalf("expected claimed task to be completed, got %q", statusByTaskID[string(claimedTask.ID)])
	}
}

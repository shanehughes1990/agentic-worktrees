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
			ID:            domaintracker.WorkItemID("epic-1"),
			BoardID:       "board-1",
			Title:         "Epic",
			RepositoryIDs: []string{"repo-1"},
			Deliverables:  []string{"Epic brief finalized"},
			State:         domaintracker.EpicStateInProgress,
			Rank:          1,
			Tasks: []domaintracker.Task{{
				ID:            domaintracker.WorkItemID("task-1"),
				BoardID:       "board-1",
				EpicID:        domaintracker.WorkItemID("epic-1"),
				Title:         "Task",
				RepositoryIDs: []string{"repo-1"},
				Deliverables:  []string{"README section updated"},
				TaskType:      "implementation",
				State:         domaintracker.TaskStatePlanned,
				Rank:          1,
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

func TestPostgresBoardStoreUpsertAndLoadBoardWithStringIDs(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	board.BoardID = "default"
	board.RunID = "ingest-1772598849983546759"
	board.Epics[0].BoardID = board.BoardID
	board.Epics[0].Tasks[0].BoardID = board.BoardID

	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board with string ids: %v", err)
	}

	loaded, err := store.LoadBoard(context.Background(), board.RunID, board.BoardID)
	if err != nil {
		t.Fatalf("load board with string ids: %v", err)
	}
	if loaded.BoardID != board.BoardID {
		t.Fatalf("expected board id %q, got %q", board.BoardID, loaded.BoardID)
	}
	if loaded.RunID != board.RunID {
		t.Fatalf("expected run id %q, got %q", board.RunID, loaded.RunID)
	}
}

func TestPostgresBoardStoreUpsertReplacesEpicAndTaskAssociations(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("initial upsert board: %v", err)
	}

	updated := board
	updated.Epics = []domaintracker.Epic{{
		ID:            domaintracker.WorkItemID("epic-2"),
		BoardID:       board.BoardID,
		Title:         "Epic 2",
		RepositoryIDs: []string{"repo-1"},
		Deliverables:  []string{"Epic 2 brief finalized"},
		State:         domaintracker.EpicStatePlanned,
		Rank:          1,
		Tasks: []domaintracker.Task{{
			ID:            domaintracker.WorkItemID("task-2"),
			BoardID:       board.BoardID,
			EpicID:        domaintracker.WorkItemID("epic-2"),
			Title:         "Task 2",
			RepositoryIDs: []string{"repo-1"},
			Deliverables:  []string{"Task 2 implementation completed"},
			TaskType:      "implementation",
			State:         domaintracker.TaskStatePlanned,
			Rank:          1,
		}},
	}}
	if err := store.UpsertBoard(context.Background(), updated); err != nil {
		t.Fatalf("second upsert board: %v", err)
	}

	loaded, err := store.LoadBoard(context.Background(), board.RunID, board.BoardID)
	if err != nil {
		t.Fatalf("load board: %v", err)
	}
	if len(loaded.Epics) != 1 {
		t.Fatalf("expected one epic after replacement upsert, got %d", len(loaded.Epics))
	}
	if string(loaded.Epics[0].ID) != "epic-2" {
		t.Fatalf("expected replaced epic id epic-2, got %s", loaded.Epics[0].ID)
	}
	if len(loaded.Epics[0].Tasks) != 1 || string(loaded.Epics[0].Tasks[0].ID) != "task-2" {
		t.Fatalf("expected replaced task task-2, got %+v", loaded.Epics[0].Tasks)
	}
}

func TestPostgresBoardStoreApplyTaskResult(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	now := time.Now().UTC()
	claimToken := "claim-agent-1"
	if err := db.Model(&projectBoardTaskRecord{}).
		Where("id = ? AND board_id = ?", "task-1", board.BoardID).
		Updates(map[string]any{
			"state":               string(domaintracker.TaskStateInProgress),
			"claimed_by_agent_id": "agent-1",
			"claimed_at":          now,
			"claim_expires_at":    now.Add(2 * time.Minute),
			"claim_token":         claimToken,
			"attempt_count":       1,
			"updated_at":          now,
		}).Error; err != nil {
		t.Fatalf("seed claimed task state: %v", err)
	}

	updatedBoard, err := store.ApplyTaskResult(
		context.Background(),
		board.RunID,
		board.BoardID,
		claimToken,
		"task-1",
		domaintracker.TaskStateCompleted,
		domaintracker.TaskOutcome{Status: domaintracker.OutcomeStatusSuccess, Summary: "completed"},
	)
	if err != nil {
		t.Fatalf("apply task result: %v", err)
	}
	if len(updatedBoard.Epics) != 1 || len(updatedBoard.Epics[0].Tasks) != 1 {
		t.Fatalf("expected single task board after result application, got %+v", updatedBoard)
	}
	resolvedTask := updatedBoard.Epics[0].Tasks[0]
	if resolvedTask.State != domaintracker.TaskStateCompleted {
		t.Fatalf("expected completed task state, got %s", resolvedTask.State)
	}
	if resolvedTask.Outcome == nil || resolvedTask.Outcome.Status != domaintracker.OutcomeStatusSuccess {
		t.Fatalf("expected success outcome to persist, got %+v", resolvedTask.Outcome)
	}
	if resolvedTask.ClaimToken != "" || resolvedTask.ClaimedByAgentID != "" {
		t.Fatalf("expected claim fields cleared after result application, got token=%q agent=%q", resolvedTask.ClaimToken, resolvedTask.ClaimedByAgentID)
	}
}

func TestPostgresBoardStorePersistsBoardEpicTaskSetTogether(t *testing.T) {
	db := newTrackerDB(t)
	store, err := NewPostgresBoardStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	board := sampleCanonicalBoard()
	if err := store.UpsertBoard(context.Background(), board); err != nil {
		t.Fatalf("upsert board: %v", err)
	}

	var boardCount int64
	if err := db.Table("project_boards").Where("id = ? AND project_id = ?", board.BoardID, board.RunID).Count(&boardCount).Error; err != nil {
		t.Fatalf("count project boards: %v", err)
	}
	if boardCount != 1 {
		t.Fatalf("expected one persisted board, got %d", boardCount)
	}

	var epicCount int64
	if err := db.Table("project_board_epics").Where("board_id = ?", board.BoardID).Count(&epicCount).Error; err != nil {
		t.Fatalf("count board epics: %v", err)
	}
	if epicCount != int64(len(board.Epics)) {
		t.Fatalf("expected %d persisted epics, got %d", len(board.Epics), epicCount)
	}

	var taskCount int64
	if err := db.Table("project_board_tasks").Where("board_id = ?", board.BoardID).Count(&taskCount).Error; err != nil {
		t.Fatalf("count board tasks: %v", err)
	}
	expectedTaskCount := int64(0)
	for _, epic := range board.Epics {
		expectedTaskCount += int64(len(epic.Tasks))
	}
	if taskCount != expectedTaskCount {
		t.Fatalf("expected %d persisted tasks, got %d", expectedTaskCount, taskCount)
	}
}

package tracker

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeNormalizedUpstreamProvider struct {
	board domaintracker.Board
	err   error
}

func (provider *fakeNormalizedUpstreamProvider) SyncBoard(ctx context.Context, request applicationtracker.ProviderSyncRequest) (domaintracker.Board, error) {
	_ = ctx
	_ = request
	if provider.err != nil {
		return domaintracker.Board{}, provider.err
	}
	return provider.board, nil
}

func normalizedSampleBoard() domaintracker.Board {
	now := time.Now().UTC()
	return domaintracker.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Title:   "Roadmap",
		Goal:    "Ship slices",
		Source: domaintracker.SourceRef{
			Kind:     domaintracker.SourceKindInternal,
			Location: "board-1",
			BoardID:  "board-1",
		},
		Status: domaintracker.StatusInProgress,
		Epics: []domaintracker.Epic{{
			WorkItem: domaintracker.WorkItem{ID: "epic-1", BoardID: "board-1", Title: "Epic", Status: domaintracker.StatusInProgress, Metadata: map[string]any{"stream": "core"}},
			Tasks: []domaintracker.Task{{
				WorkItem: domaintracker.WorkItem{ID: "task-1", BoardID: "board-1", Title: "Task", Status: domaintracker.StatusInProgress, Metadata: map[string]any{"issue_reference": "octo/repo#1"}},
				Outcome:  &domaintracker.TaskOutcome{Status: "completed", Reason: "done", TaskBranch: "task-1", Worktree: "/tmp/worktree", ResumeSessionID: "session-1", UpdatedAt: now},
			}},
		}},
		Metadata:  map[string]any{"owner": "platform"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPostgresNormalizedProviderPersistsBoardHierarchy(t *testing.T) {
	db := newTrackerDB(t)
	provider, err := NewPostgresNormalizedProvider(db, &fakeNormalizedUpstreamProvider{board: normalizedSampleBoard()})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	request := applicationtracker.ProviderSyncRequest{RunID: "run-1", ProjectID: "project-1", WorkflowID: "workflow-1", Source: domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"}}
	board, err := provider.SyncBoard(context.Background(), request)
	if err != nil {
		t.Fatalf("sync board: %v", err)
	}
	if board.BoardID != "board-1" {
		t.Fatalf("unexpected board id %q", board.BoardID)
	}
	var boards int64
	if err := db.Model(&trackerBoardRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&boards).Error; err != nil {
		t.Fatalf("count normalized boards: %v", err)
	}
	if boards != 1 {
		t.Fatalf("expected 1 normalized board, got %d", boards)
	}
	var epics int64
	if err := db.Model(&trackerEpicRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&epics).Error; err != nil {
		t.Fatalf("count normalized epics: %v", err)
	}
	if epics != 1 {
		t.Fatalf("expected 1 normalized epic, got %d", epics)
	}
	var tasks int64
	if err := db.Model(&trackerTaskRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&tasks).Error; err != nil {
		t.Fatalf("count normalized tasks: %v", err)
	}
	if tasks != 1 {
		t.Fatalf("expected 1 normalized task, got %d", tasks)
	}
	var outcomes int64
	if err := db.Model(&trackerTaskOutcomeRecord{}).Where("run_id = ? AND board_id = ?", "run-1", "board-1").Count(&outcomes).Error; err != nil {
		t.Fatalf("count normalized outcomes: %v", err)
	}
	if outcomes != 1 {
		t.Fatalf("expected 1 normalized outcome, got %d", outcomes)
	}
}

func TestPostgresNormalizedProviderPropagatesUpstreamError(t *testing.T) {
	db := newTrackerDB(t)
	provider, err := NewPostgresNormalizedProvider(db, &fakeNormalizedUpstreamProvider{err: errors.New("upstream failed")})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	request := applicationtracker.ProviderSyncRequest{RunID: "run-1", ProjectID: "project-1", WorkflowID: "workflow-1", Source: domaintracker.SourceRef{Kind: domaintracker.SourceKindInternal, Location: "board-1"}}
	if _, err := provider.SyncBoard(context.Background(), request); err == nil {
		t.Fatalf("expected upstream error")
	}
}

package tracker

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"testing"
	"time"
)

type fakeBoardStore struct {
	board     domaintracker.Board
	task      domaintracker.Task
	claimToken string
	err       error
}

func (store *fakeBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	_ = ctx
	store.board = board
	return store.err
}

func (store *fakeBoardStore) ClaimNextTask(ctx context.Context, projectID string, boardID string, agentID string, leaseTTL time.Duration) (domaintracker.Board, domaintracker.Task, string, error) {
	_, _, _, _ = projectID, boardID, agentID, leaseTTL
	_ = ctx
	return store.board, store.task, store.claimToken, store.err
}

func (store *fakeBoardStore) ApplyTaskResult(ctx context.Context, projectID string, boardID string, claimToken string, taskID string, nextState domaintracker.TaskState, outcome domaintracker.TaskOutcome) (domaintracker.Board, error) {
	_, _, _, _, _, _ = projectID, boardID, claimToken, taskID, nextState, outcome
	_ = ctx
	return store.board, store.err
}

func (store *fakeBoardStore) ListBoards(ctx context.Context, projectID string) ([]domaintracker.Board, error) {
	_, _ = ctx, projectID
	if store.err != nil {
		return nil, store.err
	}
	return []domaintracker.Board{store.board}, nil
}

func (store *fakeBoardStore) LoadBoard(ctx context.Context, projectID string, boardID string) (domaintracker.Board, error) {
	_, _, _ = ctx, projectID, boardID
	return store.board, store.err
}

func (store *fakeBoardStore) DeleteBoard(ctx context.Context, projectID string, boardID string) error {
	_, _, _ = ctx, projectID, boardID
	return store.err
}

func sampleBoard() domaintracker.Board {
	now := time.Now().UTC()
	return domaintracker.Board{
		BoardID: "board-1",
		RunID:   "project-1",
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

func TestClaimNextTaskValidatesRequiredFields(t *testing.T) {
	service, err := NewTaskMutationService(&fakeBoardStore{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.ClaimNextTask(context.Background(), ClaimNextTaskRequest{})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestApplyTaskResultValidatesRequiredFields(t *testing.T) {
	service, err := NewTaskMutationService(&fakeBoardStore{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.ApplyTaskResult(context.Background(), ApplyTaskResultRequest{})
	if !failures.IsClass(err, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation error, got %q (%v)", failures.ClassOf(err), err)
	}
}

func TestUpsertBoardClassifiesUnknownStoreErrorsAsTransient(t *testing.T) {
	service, err := NewTaskMutationService(&fakeBoardStore{err: errors.New("db unavailable")})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.UpsertBoard(context.Background(), sampleBoard()); !failures.IsClass(err, failures.ClassTransient) {
		t.Fatalf("expected transient error classification, got %q (%v)", failures.ClassOf(err), err)
	}
}

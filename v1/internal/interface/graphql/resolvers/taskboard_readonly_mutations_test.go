package resolvers

import (
	applicationtracker "agentic-orchestrator/internal/application/tracker"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type taskboardMutationFakeBoardStore struct {
	board       domaintracker.Board
	upsertCalls int
	deleteCalls int
}

func (store *taskboardMutationFakeBoardStore) UpsertBoard(_ context.Context, board domaintracker.Board) error {
	store.upsertCalls++
	store.board = board
	return nil
}

func (store *taskboardMutationFakeBoardStore) ListBoards(_ context.Context, projectID string) ([]domaintracker.Board, error) {
	if strings.TrimSpace(projectID) == strings.TrimSpace(store.board.ProjectID) {
		return []domaintracker.Board{store.board}, nil
	}
	return nil, nil
}

func (store *taskboardMutationFakeBoardStore) LoadBoard(_ context.Context, projectID string, boardID string) (domaintracker.Board, error) {
	if strings.TrimSpace(projectID) != strings.TrimSpace(store.board.ProjectID) || strings.TrimSpace(boardID) != strings.TrimSpace(store.board.BoardID) {
		return domaintracker.Board{}, errors.New("board not found")
	}
	return store.board, nil
}

func (store *taskboardMutationFakeBoardStore) DeleteBoard(_ context.Context, projectID string, boardID string) error {
	if strings.TrimSpace(projectID) != strings.TrimSpace(store.board.ProjectID) || strings.TrimSpace(boardID) != strings.TrimSpace(store.board.BoardID) {
		return errors.New("board not found")
	}
	store.deleteCalls++
	return nil
}

func (store *taskboardMutationFakeBoardStore) ClaimNextTask(context.Context, string, string, string, time.Duration) (domaintracker.Board, domaintracker.Task, string, error) {
	return domaintracker.Board{}, domaintracker.Task{}, "", errors.New("not implemented")
}

func (store *taskboardMutationFakeBoardStore) ApplyTaskResult(context.Context, string, string, string, string, domaintracker.TaskState, domaintracker.TaskOutcome) (domaintracker.Board, error) {
	return domaintracker.Board{}, errors.New("not implemented")
}

func newTaskboardMutationResolverWithBoard(t *testing.T, board domaintracker.Board) (*mutationResolver, *taskboardMutationFakeBoardStore) {
	t.Helper()
	store := &taskboardMutationFakeBoardStore{board: board}
	trackerService, err := applicationtracker.NewTaskMutationService(store)
	if err != nil {
		t.Fatalf("new tracker service: %v", err)
	}
	resolver := NewResolver(nil, nil, nil, nil, trackerService)
	return &mutationResolver{resolver}, store
}

func TestTaskboardMutationsRejectEndedBoardsAsReadOnly(t *testing.T) {
	now := time.Now().UTC()
	endedBoard := domaintracker.Board{
		BoardID:    "board-1",
		RunID:      "project-1",
		ProjectID:  "project-1",
		Name:       "Board One",
		State:      domaintracker.BoardStateCompleted,
		CreatedAt:  now,
		UpdatedAt:  now,
		Epics:      []domaintracker.Epic{},
		IngestionAudits: []domaintracker.TaskModelAudit{},
	}

	testCases := []struct {
		name string
		run  func(*mutationResolver) (any, error)
	}{
		{
			name: "update taskboard",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.UpdateTaskboard(context.Background(), models.UpdateTaskboardInput{ProjectID: "project-1", BoardID: "board-1", Name: "Renamed", State: string(domaintracker.BoardStateActive)})
			},
		},
		{
			name: "delete taskboard",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.DeleteTaskboard(context.Background(), models.DeleteTaskboardInput{ProjectID: "project-1", BoardID: "board-1"})
			},
		},
		{
			name: "create epic",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.CreateTaskboardEpic(context.Background(), models.CreateTaskboardEpicInput{ProjectID: "project-1", BoardID: "board-1", Title: "Epic 1", State: string(domaintracker.EpicStatePlanned), Rank: 1})
			},
		},
		{
			name: "update epic",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.UpdateTaskboardEpic(context.Background(), models.UpdateTaskboardEpicInput{ProjectID: "project-1", BoardID: "board-1", EpicID: "epic-1", Title: "Epic 1", State: string(domaintracker.EpicStateInProgress), Rank: 1})
			},
		},
		{
			name: "delete epic",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.DeleteTaskboardEpic(context.Background(), models.DeleteTaskboardEpicInput{ProjectID: "project-1", BoardID: "board-1", EpicID: "epic-1"})
			},
		},
		{
			name: "create task",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.CreateTaskboardTask(context.Background(), models.CreateTaskboardTaskInput{ProjectID: "project-1", BoardID: "board-1", EpicID: "epic-1", Title: "Task 1", TaskType: "implementation", State: string(domaintracker.TaskStatePlanned), Rank: 1})
			},
		},
		{
			name: "update task",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.UpdateTaskboardTask(context.Background(), models.UpdateTaskboardTaskInput{ProjectID: "project-1", BoardID: "board-1", EpicID: "epic-1", TaskID: "task-1", Title: "Task 1", TaskType: "implementation", State: string(domaintracker.TaskStateInProgress), Rank: 1})
			},
		},
		{
			name: "delete task",
			run: func(resolver *mutationResolver) (any, error) {
				return resolver.DeleteTaskboardTask(context.Background(), models.DeleteTaskboardTaskInput{ProjectID: "project-1", BoardID: "board-1", TaskID: "task-1"})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolver, store := newTaskboardMutationResolverWithBoard(t, endedBoard)
			result, err := testCase.run(resolver)
			if err != nil {
				t.Fatalf("mutation error = %v", err)
			}
			graphErr, ok := result.(models.GraphError)
			if !ok {
				t.Fatalf("expected GraphError, got %T", result)
			}
			if graphErr.Code != models.GraphErrorCodeConflict {
				t.Fatalf("expected CONFLICT, got %s", graphErr.Code)
			}
			if !strings.Contains(strings.ToLower(graphErr.Message), "read-only") {
				t.Fatalf("expected read-only error message, got %q", graphErr.Message)
			}
			if store.upsertCalls != 0 {
				t.Fatalf("expected no upsert writes, got %d", store.upsertCalls)
			}
			if store.deleteCalls != 0 {
				t.Fatalf("expected no delete writes, got %d", store.deleteCalls)
			}
		})
	}
}

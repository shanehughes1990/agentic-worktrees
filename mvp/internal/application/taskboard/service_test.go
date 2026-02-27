package taskboard

import (
	"context"
	"errors"
	"testing"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

type stubRepository struct {
	board *domaintaskboard.Board
	err   error
	saved bool
}

func (repository *stubRepository) ListBoardIDs(context.Context) ([]string, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	if repository.board == nil {
		return []string{}, nil
	}
	return []string{repository.board.BoardID}, nil
}

func (repository *stubRepository) GetByBoardID(context.Context, string) (*domaintaskboard.Board, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	return repository.board, nil
}

func (repository *stubRepository) Save(context.Context, *domaintaskboard.Board) error {
	repository.saved = true
	return repository.err
}

func TestServiceGetNextTask(t *testing.T) {
	now := time.Now().UTC()
	repository := &stubRepository{board: &domaintaskboard.Board{
		BoardID: "b1", RunID: "r1", Status: domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "b1", Title: "epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "b1", Title: "task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	service := NewService(repository)
	nextTask, err := service.GetNextTask(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextTask == nil || nextTask.ID != "t1" {
		t.Fatalf("expected t1 next, got %#v", nextTask)
	}
}

func TestServiceMarkTaskCompletedSavesBoard(t *testing.T) {
	now := time.Now().UTC()
	repository := &stubRepository{board: &domaintaskboard.Board{
		BoardID: "b1", RunID: "r1", Status: domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "b1", Title: "epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "b1", Title: "task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	service := NewService(repository)
	if err := service.MarkTaskCompleted(context.Background(), "b1", "t1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repository.saved {
		t.Fatalf("expected repository save to be called")
	}
}

func TestServiceGetNextTaskRepositoryError(t *testing.T) {
	repository := &stubRepository{err: errors.New("boom")}
	service := NewService(repository)
	if _, err := service.GetNextTask(context.Background(), "b1"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceCheckpointTaskResumeSessionPersistsSessionID(t *testing.T) {
	now := time.Now().UTC()
	repository := &stubRepository{board: &domaintaskboard.Board{
		BoardID: "b1", RunID: "r1", Status: domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "b1", Title: "epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "b1", Title: "task", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	service := NewService(repository)
	if err := service.CheckpointTaskResumeSession(context.Background(), "b1", "t1", "session-123"); err != nil {
		t.Fatalf("unexpected checkpoint error: %v", err)
	}
	if !repository.saved {
		t.Fatalf("expected repository save on checkpoint")
	}
	if repository.board.Epics[0].Tasks[0].Outcome == nil || repository.board.Epics[0].Tasks[0].Outcome.ResumeSessionID != "session-123" {
		t.Fatalf("expected resume session checkpoint to persist, got %#v", repository.board.Epics[0].Tasks[0].Outcome)
	}
}

func TestServiceCheckpointTaskResumeSessionNoopForEmptySession(t *testing.T) {
	now := time.Now().UTC()
	repository := &stubRepository{board: &domaintaskboard.Board{
		BoardID: "b1", RunID: "r1", Status: domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "b1", Title: "epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "b1", Title: "task", Status: domaintaskboard.StatusInProgress},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}}

	service := NewService(repository)
	if err := service.CheckpointTaskResumeSession(context.Background(), "b1", "t1", ""); err != nil {
		t.Fatalf("unexpected checkpoint error: %v", err)
	}
	if repository.saved {
		t.Fatalf("expected no save for empty session checkpoint")
	}
}

package jsonrepo

import (
	"context"
	"testing"
	"time"

	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
)

func TestRepositorySaveAndLoadBoard(t *testing.T) {
	repository, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	now := time.Now().UTC()
	board := &domaintaskboard.Board{
		BoardID: "board-1",
		RunID:   "run-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "board-1", Title: "Task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repository.Save(context.Background(), board); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	loaded, err := repository.GetByBoardID(context.Background(), "board-1")
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if loaded == nil || loaded.BoardID != "board-1" {
		t.Fatalf("expected loaded board-1, got %#v", loaded)
	}
}

func TestRepositoryWorkflowList(t *testing.T) {
	repository, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	workflow := &apptaskboard.IngestionWorkflow{RunID: "run-1", Status: apptaskboard.WorkflowStatusQueued}
	workflow.Normalize("run-1")
	if err := repository.SaveWorkflow(context.Background(), workflow); err != nil {
		t.Fatalf("unexpected save workflow error: %v", err)
	}

	workflows, err := repository.ListWorkflows(context.Background())
	if err != nil {
		t.Fatalf("unexpected list workflows error: %v", err)
	}
	if len(workflows) != 1 || workflows[0].RunID != "run-1" {
		t.Fatalf("expected workflow run-1, got %#v", workflows)
	}
}

func TestRepositorySaveRejectsIncompleteBoard(t *testing.T) {
	repository, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	now := time.Now().UTC()
	board := &domaintaskboard.Board{
		BoardID: "board-2",
		RunID:   "run-2",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "board-2", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks:    nil,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repository.Save(context.Background(), board); err == nil {
		t.Fatalf("expected save to fail for incomplete board")
	}

	loaded, err := repository.GetByBoardID(context.Background(), "board-2")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected no persisted board for failed save, got %#v", loaded)
	}
}

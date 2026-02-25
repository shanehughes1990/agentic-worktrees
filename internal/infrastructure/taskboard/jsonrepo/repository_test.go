package jsonrepo

import (
	"context"
	"os"
	"path/filepath"
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

func TestRepositoryListBoardIDs(t *testing.T) {
	repository, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	now := time.Now().UTC()
	board := &domaintaskboard.Board{
		BoardID: "board-list-1",
		RunID:   "run-list-1",
		Status:  domaintaskboard.StatusInProgress,
		Epics: []domaintaskboard.Epic{{
			WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "board-list-1", Title: "Epic", Status: domaintaskboard.StatusInProgress},
			Tasks: []domaintaskboard.Task{{
				WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "board-list-1", Title: "Task", Status: domaintaskboard.StatusNotStarted},
			}},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repository.Save(context.Background(), board); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	workflow := &apptaskboard.IngestionWorkflow{RunID: "run-list-1", Status: apptaskboard.WorkflowStatusQueued}
	workflow.Normalize("run-list-1")
	if err := repository.SaveWorkflow(context.Background(), workflow); err != nil {
		t.Fatalf("unexpected save workflow error: %v", err)
	}
	runState := &apptaskboard.RunState{RunID: "run-list-1", Status: apptaskboard.WorkflowStatusRunning}
	if err := repository.SaveRunState(context.Background(), runState); err != nil {
		t.Fatalf("unexpected save run state error: %v", err)
	}
	jobState := &apptaskboard.JobState{RunID: "run-list-1", JobID: "job-1", Status: "running", Attempt: 1}
	if err := repository.SaveJobState(context.Background(), jobState); err != nil {
		t.Fatalf("unexpected save job state error: %v", err)
	}

	boardIDs, err := repository.ListBoardIDs(context.Background())
	if err != nil {
		t.Fatalf("unexpected list board ids error: %v", err)
	}
	if len(boardIDs) != 1 || boardIDs[0] != "board-list-1" {
		t.Fatalf("expected board-list-1, got %#v", boardIDs)
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

func TestRepositoryStoresWorkflowFilesInConfiguredWorkflowsDirectory(t *testing.T) {
	taskboardsDirectory := filepath.Join(t.TempDir(), "taskboards")
	workflowsDirectory := filepath.Join(t.TempDir(), "workflows")

	repository, err := NewRepositoryWithWorkflowDirectory(taskboardsDirectory, workflowsDirectory)
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	workflow := &apptaskboard.IngestionWorkflow{RunID: "run-split-1", Status: apptaskboard.WorkflowStatusQueued}
	workflow.Normalize("run-split-1")
	if err := repository.SaveWorkflow(context.Background(), workflow); err != nil {
		t.Fatalf("unexpected save workflow error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workflowsDirectory, "workflow-run-split-1.json")); err != nil {
		t.Fatalf("expected workflow file in workflows directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(taskboardsDirectory, "workflow-run-split-1.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no workflow file in taskboards directory, got err=%v", err)
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

func TestRepositorySaveAndLoadRunState(t *testing.T) {
	repository, err := NewRepository(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	runState := &apptaskboard.RunState{
		RunID:   "run-state-1",
		BoardID: "board-1",
		Status:  apptaskboard.WorkflowStatusRunning,
		Message: "started",
		Checkpoints: []apptaskboard.RunCheckpoint{
			{Name: "ingestion_started", Timestamp: time.Now().UTC()},
		},
	}
	if err := repository.SaveRunState(context.Background(), runState); err != nil {
		t.Fatalf("unexpected save run state error: %v", err)
	}

	loaded, err := repository.GetRunState(context.Background(), "run-state-1")
	if err != nil {
		t.Fatalf("unexpected load run state error: %v", err)
	}
	if loaded == nil || loaded.RunID != "run-state-1" {
		t.Fatalf("expected run-state-1, got %#v", loaded)
	}

	runs, err := repository.ListRunStates(context.Background())
	if err != nil {
		t.Fatalf("unexpected list run states error: %v", err)
	}
	if len(runs) != 1 || runs[0].RunID != "run-state-1" {
		t.Fatalf("expected listed run-state-1, got %#v", runs)
	}
}

func TestRepositorySaveAndLoadJobState(t *testing.T) {
	root := t.TempDir()
	repository, err := NewRepository(root)
	if err != nil {
		t.Fatalf("unexpected repository creation error: %v", err)
	}

	jobState := &apptaskboard.JobState{
		RunID:        "run-job-1",
		JobID:        "job-1",
		TaskID:       "task-1",
		Attempt:      2,
		Status:       "failed",
		FailureClass: "transient",
		ResultRef:    "result://job-1",
	}
	if err := repository.SaveJobState(context.Background(), jobState); err != nil {
		t.Fatalf("unexpected save job state error: %v", err)
	}

	loaded, err := repository.GetJobState(context.Background(), "run-job-1", "job-1")
	if err != nil {
		t.Fatalf("unexpected load job state error: %v", err)
	}
	if loaded == nil || loaded.JobID != "job-1" || loaded.RunID != "run-job-1" {
		t.Fatalf("expected job run-job-1/job-1, got %#v", loaded)
	}

	jobs, err := repository.ListJobStatesByRunID(context.Background(), "run-job-1")
	if err != nil {
		t.Fatalf("unexpected list job states error: %v", err)
	}
	if len(jobs) != 1 || jobs[0].JobID != "job-1" {
		t.Fatalf("expected listed job-1, got %#v", jobs)
	}

	if _, err := os.Stat(filepath.Join(root, "job-run-job-1-job-1.json.tmp")); !os.IsNotExist(err) {
		t.Fatalf("expected no temp file after atomic write, got err=%v", err)
	}
}

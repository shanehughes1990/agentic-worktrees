package taskboard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
	filesystemsource "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/taskboard/filesystemsource"
)

type fakeDispatcher struct {
	enqueue func(ctx context.Context, job IngestionJob) (string, error)
}

func (dispatcher *fakeDispatcher) EnqueueIngestion(ctx context.Context, job IngestionJob) (string, error) {
	if dispatcher.enqueue != nil {
		return dispatcher.enqueue(ctx, job)
	}
	return "task-1", nil
}

type pollingRepository struct {
	boards    map[string]*domaintaskboard.Board
	workflows map[string]*IngestionWorkflow
}

func newPollingRepository() *pollingRepository {
	return &pollingRepository{
		boards:    make(map[string]*domaintaskboard.Board),
		workflows: make(map[string]*IngestionWorkflow),
	}
}

func (repository *pollingRepository) GetByBoardID(_ context.Context, boardID string) (*domaintaskboard.Board, error) {
	return repository.boards[boardID], nil
}

func (repository *pollingRepository) ListBoardIDs(_ context.Context) ([]string, error) {
	boardIDs := make([]string, 0, len(repository.boards))
	for boardID := range repository.boards {
		boardIDs = append(boardIDs, boardID)
	}
	return boardIDs, nil
}

func (repository *pollingRepository) Save(_ context.Context, board *domaintaskboard.Board) error {
	repository.boards[board.BoardID] = board
	return nil
}

func (repository *pollingRepository) GetWorkflow(_ context.Context, runID string) (*IngestionWorkflow, error) {
	return repository.workflows[runID], nil
}

func (repository *pollingRepository) ListWorkflows(_ context.Context) ([]IngestionWorkflow, error) {
	workflows := make([]IngestionWorkflow, 0, len(repository.workflows))
	for _, workflow := range repository.workflows {
		workflows = append(workflows, *workflow)
	}
	return workflows, nil
}

func (repository *pollingRepository) SaveWorkflow(_ context.Context, workflow *IngestionWorkflow) error {
	repository.workflows[workflow.RunID] = workflow
	return nil
}

func TestIngestDirectoryReturnsBoardAndRunID(t *testing.T) {
	repository := newPollingRepository()
	sourceAdapter := filesystemsource.NewAdapter()
	dispatcher := &fakeDispatcher{
		enqueue: func(_ context.Context, job IngestionJob) (string, error) {
			repository.boards[job.RunID] = &domaintaskboard.Board{BoardID: job.RunID, RunID: job.RunID}
			return "task-1", nil
		},
	}
	service := NewIngestionService(dispatcher, repository, repository, sourceAdapter, sourceAdapter, "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := service.IngestDirectory(ctx, ".")
	if err != nil {
		t.Fatalf("unexpected ingestion error: %v", err)
	}
	if result.BoardID == "" || result.RunID == "" {
		t.Fatalf("expected board id and run id, got %#v", result)
	}
}

func TestListWorkflows(t *testing.T) {
	repository := newPollingRepository()
	repository.workflows["run-1"] = &IngestionWorkflow{RunID: "run-1", Status: WorkflowStatusQueued}
	sourceAdapter := filesystemsource.NewAdapter()
	service := NewIngestionService(&fakeDispatcher{}, repository, repository, sourceAdapter, sourceAdapter, "")

	workflows, err := service.ListWorkflows(context.Background())
	if err != nil {
		t.Fatalf("unexpected list workflows error: %v", err)
	}
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
}

func TestIngestSupportsFileSource(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "scope.md")
	if err := os.WriteFile(filePath, []byte("scope"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	repository := newPollingRepository()
	sourceAdapter := filesystemsource.NewAdapter()
	dispatcher := &fakeDispatcher{
		enqueue: func(_ context.Context, job IngestionJob) (string, error) {
			repository.boards[job.RunID] = &domaintaskboard.Board{BoardID: job.RunID, RunID: job.RunID}
			if job.WorkingDirectory != filepath.Dir(filePath) {
				t.Fatalf("expected working directory to be source file parent directory, got %s", job.WorkingDirectory)
			}
			return "task-1", nil
		},
	}
	service := NewIngestionService(dispatcher, repository, repository, sourceAdapter, sourceAdapter, "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := service.Ingest(ctx, IngestRequest{SourcePath: filePath, SourceType: IngestionSourceTypeFile})
	if err != nil {
		t.Fatalf("unexpected ingestion error: %v", err)
	}
	if result.BoardID == "" || result.RunID == "" {
		t.Fatalf("expected board id and run id, got %#v", result)
	}
}

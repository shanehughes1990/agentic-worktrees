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

type recordingFilesystemSourceAdapter struct {
	delegate    *filesystemsource.Adapter
	listed      []domaintaskboard.SourceMetadata
	listOptions []domaintaskboard.SourceListOptions
	read        []domaintaskboard.SourceIdentity
}

func newRecordingFilesystemSourceAdapter() *recordingFilesystemSourceAdapter {
	return &recordingFilesystemSourceAdapter{
		delegate: filesystemsource.NewAdapter(),
	}
}

func (adapter *recordingFilesystemSourceAdapter) List(ctx context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	adapter.listed = append(adapter.listed, source)
	adapter.listOptions = append(adapter.listOptions, options)
	return adapter.delegate.List(ctx, source, options)
}

func (adapter *recordingFilesystemSourceAdapter) Read(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
	adapter.read = append(adapter.read, source)
	return adapter.delegate.Read(ctx, source)
}

func (adapter *recordingFilesystemSourceAdapter) ResolveWorkingDirectory(ctx context.Context, source domaintaskboard.SourceIdentity) (string, error) {
	return adapter.delegate.ResolveWorkingDirectory(ctx, source)
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

func TestIngestDirectoryUsesSourceAbstractionForFilesystemSource(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "scope.md")
	if err := os.WriteFile(filePath, []byte("scope"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	repository := newPollingRepository()
	sourceAdapter := newRecordingFilesystemSourceAdapter()
	dispatcher := &fakeDispatcher{
		enqueue: func(_ context.Context, job IngestionJob) (string, error) {
			repository.boards[job.RunID] = &domaintaskboard.Board{BoardID: job.RunID, RunID: job.RunID}
			return "task-1", nil
		},
	}
	service := NewIngestionService(dispatcher, repository, repository, sourceAdapter, sourceAdapter, "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := service.IngestDirectory(ctx, dir)
	if err != nil {
		t.Fatalf("unexpected ingestion error: %v", err)
	}
	if result.BoardID == "" || result.RunID == "" {
		t.Fatalf("expected board id and run id, got %#v", result)
	}
	if len(sourceAdapter.listed) == 0 {
		t.Fatalf("expected ingestion to list source documents through source lister abstraction")
	}
	if sourceAdapter.listed[0].Identity.Kind != domaintaskboard.SourceKindFolder || sourceAdapter.listed[0].Identity.Locator != dir {
		t.Fatalf("expected first list call to target source folder %q, got %#v", dir, sourceAdapter.listed[0].Identity)
	}
	if len(sourceAdapter.listOptions) == 0 || sourceAdapter.listOptions[0].WalkDepth != -1 {
		t.Fatalf("expected ingest directory walk depth to flow through source list options, got %#v", sourceAdapter.listOptions)
	}
	if len(sourceAdapter.read) == 0 {
		t.Fatalf("expected ingestion to read source documents through source reader abstraction")
	}
	readObserved := false
	for _, identity := range sourceAdapter.read {
		if identity.Kind == domaintaskboard.SourceKindFile && identity.Locator == filePath {
			readObserved = true
			break
		}
	}
	if !readObserved {
		t.Fatalf("expected source reader to read %q, got %#v", filePath, sourceAdapter.read)
	}
}

package taskboard

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

type fakeSourceProvider struct {
	list                    func(ctx context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error)
	read                    func(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error)
	resolveWorkingDirectory func(ctx context.Context, source domaintaskboard.SourceIdentity) (string, error)
}

func (provider *fakeSourceProvider) List(ctx context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
	if provider.list != nil {
		return provider.list(ctx, source, options)
	}
	return nil, nil
}

func (provider *fakeSourceProvider) Read(ctx context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
	if provider.read != nil {
		return provider.read(ctx, source)
	}
	return nil, nil
}

func (provider *fakeSourceProvider) ResolveWorkingDirectory(ctx context.Context, source domaintaskboard.SourceIdentity) (string, error) {
	if provider.resolveWorkingDirectory != nil {
		return provider.resolveWorkingDirectory(ctx, source)
	}
	return "", nil
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

func TestIngestOrchestratesWithFakeSourceProvider(t *testing.T) {
	repository := newPollingRepository()
	readSources := make([]domaintaskboard.SourceIdentity, 0, 2)
	provider := &fakeSourceProvider{
		list: func(_ context.Context, source domaintaskboard.SourceMetadata, options domaintaskboard.SourceListOptions) ([]domaintaskboard.SourceListEntry, error) {
			if source.Identity.Kind != domaintaskboard.SourceKindFolder || source.Identity.Locator != "provider://scope" {
				t.Fatalf("unexpected source metadata: %#v", source)
			}
			if options.WalkDepth != 3 || len(options.IgnorePaths) != 1 || options.IgnorePaths[0] != "ignored" {
				t.Fatalf("unexpected list options: %#v", options)
			}
			if len(options.IgnoreExtensions) != 1 || options.IgnoreExtensions[0] != ".tmp" {
				t.Fatalf("unexpected list options: %#v", options)
			}
			return []domaintaskboard.SourceListEntry{
				{
					Identity: domaintaskboard.SourceIdentity{
						Kind:    domaintaskboard.SourceKindFile,
						Locator: "provider://scope/a.md",
					},
					RelativePath: "a.md",
				},
				{
					Identity: domaintaskboard.SourceIdentity{
						Kind:    domaintaskboard.SourceKindFile,
						Locator: "provider://scope/b.md",
					},
					RelativePath: "b.md",
				},
			}, nil
		},
		read: func(_ context.Context, source domaintaskboard.SourceIdentity) ([]byte, error) {
			readSources = append(readSources, source)
			switch source.Locator {
			case "provider://scope/a.md":
				return []byte("first"), nil
			case "provider://scope/b.md":
				return []byte("second"), nil
			default:
				t.Fatalf("unexpected source read: %#v", source)
			}
			return nil, nil
		},
		resolveWorkingDirectory: func(_ context.Context, source domaintaskboard.SourceIdentity) (string, error) {
			if source.Kind != domaintaskboard.SourceKindFolder || source.Locator != "provider://scope" {
				t.Fatalf("unexpected source identity for working directory: %#v", source)
			}
			return "/virtual/worktree", nil
		},
	}

	var capturedJob IngestionJob
	dispatcher := &fakeDispatcher{
		enqueue: func(_ context.Context, job IngestionJob) (string, error) {
			capturedJob = job
			repository.boards[job.RunID] = &domaintaskboard.Board{BoardID: job.RunID, RunID: job.RunID}
			return "task-1", nil
		},
	}
	service := NewIngestionService(dispatcher, repository, repository, provider, provider, "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := service.Ingest(ctx, IngestRequest{
		SourcePath: "provider://scope",
		SourceType: IngestionSourceTypeFolder,
		Folder: FolderTraversalOptions{
			WalkDepth:        3,
			IgnorePaths:      []string{"ignored"},
			IgnoreExtensions: []string{".tmp"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected ingestion error: %v", err)
	}
	if result.BoardID == "" || result.RunID == "" {
		t.Fatalf("expected board id and run id, got %#v", result)
	}
	if len(readSources) != 2 {
		t.Fatalf("expected 2 source reads, got %d", len(readSources))
	}
	if capturedJob.WorkingDirectory != "/virtual/worktree" {
		t.Fatalf("expected provider working directory, got %q", capturedJob.WorkingDirectory)
	}
	if !strings.Contains(capturedJob.Prompt, "path: a.md") || !strings.Contains(capturedJob.Prompt, "content:\nfirst") {
		t.Fatalf("expected prompt to include normalized first document, got %q", capturedJob.Prompt)
	}
	if !strings.Contains(capturedJob.Prompt, "path: b.md") || !strings.Contains(capturedJob.Prompt, "content:\nsecond") {
		t.Fatalf("expected prompt to include normalized second document, got %q", capturedJob.Prompt)
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

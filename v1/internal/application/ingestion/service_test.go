package ingestion

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeBoardStore struct {
	board domaintracker.Board
}

func (store *fakeBoardStore) UpsertBoard(ctx context.Context, board domaintracker.Board) error {
	_ = ctx
	store.board = board
	return nil
}

type fakeArtifactFetcher struct {
	err             error
	fetchedObject   []string
	destinationPath []string
}

func (fetcher *fakeArtifactFetcher) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	_ = ctx
	if fetcher.err != nil {
		return fetcher.err
	}
	fetcher.fetchedObject = append(fetcher.fetchedObject, objectPath)
	fetcher.destinationPath = append(fetcher.destinationPath, destinationPath)
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destinationPath, []byte("doc content for taskboard"), 0o644)
}

type fakeAgentRunner struct {
	err          error
	prompt       string
	sandboxDir   string
	outputPath   string
	model        string
	writtenBoard string
	writtenBoards []string
	callCount     int
}

func (runner *fakeAgentRunner) GenerateTaskboard(ctx context.Context, sandboxDir string, prompt string, outputPath string, model string, runContext AgentRunContext) (AgentRunContext, error) {
	_ = ctx
	if runner.err != nil {
		return runContext, runner.err
	}
	runner.prompt = prompt
	runner.sandboxDir = sandboxDir
	runner.outputPath = outputPath
	runner.model = model
	runner.callCount++
	if len(runner.writtenBoards) > 0 {
		index := runner.callCount - 1
		if index >= len(runner.writtenBoards) {
			index = len(runner.writtenBoards) - 1
		}
		runner.writtenBoard = runner.writtenBoards[index]
	}
	if strings.TrimSpace(runner.writtenBoard) == "" {
		runner.writtenBoard = `{
		"board_id": "temporary-board",
		"run_id": "temporary-run",
		"state": "pending",
		"epics": [{
			"id": "epic-1",
			"board_id": "temporary-board",
			"title": "Epic one",
			"state": "planned",
			"rank": 1,
			"tasks": [{
				"id": "task-1",
				"board_id": "temporary-board",
				"epic_id": "epic-1",
				"title": "Task one",
				"task_type": "implementation",
				"state": "planned",
				"rank": 1
			}]
		}],
		"created_at": "2026-03-03T10:00:00Z",
		"updated_at": "2026-03-03T10:00:00Z"
	}`
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return runContext, err
	}
	if err := os.WriteFile(outputPath, []byte(runner.writtenBoard), 0o644); err != nil {
		return runContext, err
	}
	if strings.TrimSpace(runContext.StreamID) == "" {
		runContext.StreamID = "test-stream"
	}
	if strings.TrimSpace(runContext.SessionID) == "" {
		runContext.SessionID = "test-session"
	}
	return runContext, nil
}

type fakeRepositorySynchronizer struct {
	err                error
	projectID          string
	sandboxDir         string
	sourceBranch       string
	sourceRepositories []SourceRepository
}

func (synchronizer *fakeRepositorySynchronizer) Sync(ctx context.Context, projectID string, sandboxDir string, sourceBranch string, sourceRepositories []SourceRepository) error {
	_ = ctx
	if synchronizer.err != nil {
		return synchronizer.err
	}
	synchronizer.projectID = projectID
	synchronizer.sandboxDir = sandboxDir
	synchronizer.sourceBranch = sourceBranch
	synchronizer.sourceRepositories = append([]SourceRepository(nil), sourceRepositories...)
	return nil
}

func TestServiceExecuteGeneratesAndPersistsBoard(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{}
	repositorySynchronizer := &fakeRepositorySynchronizer{}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, repositorySynchronizer)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	board, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SourceRepositories:        []SourceRepository{{RepositoryID: "repo-1", RepositoryURL: "https://github.com/acme/source-repo.git"}},
		SourceBranch:              "develop",
		Model:                     "gpt-5.3-codex",
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if len(artifactFetcher.fetchedObject) != 1 {
		t.Fatalf("expected one fetched object, got %d", len(artifactFetcher.fetchedObject))
	}
	if !strings.Contains(agentRunner.prompt, "You are an ingestion planner.") {
		t.Fatalf("expected system prompt in composed prompt, got %q", agentRunner.prompt)
	}
	if !strings.Contains(agentRunner.prompt, "Create a delivery taskboard.") {
		t.Fatalf("expected user prompt in composed prompt, got %q", agentRunner.prompt)
	}
	if !strings.Contains(agentRunner.prompt, "treat depends_on as a required modeling tool") {
		t.Fatalf("expected strict depends_on guidance in composed prompt")
	}
	if !strings.Contains(agentRunner.prompt, "Do not create duplicate or near-duplicate tasks") {
		t.Fatalf("expected anti-duplicate guidance in composed prompt")
	}
	if !strings.Contains(agentRunner.prompt, "one implementation-owner task per behavior change") {
		t.Fatalf("expected single-owner behavior guidance in composed prompt")
	}
	if !strings.Contains(agentRunner.prompt, "verification tasks should validate one check per task") {
		t.Fatalf("expected one-check-per-task verification guidance in composed prompt")
	}
	if agentRunner.model != "gpt-5.3-codex" {
		t.Fatalf("expected model gpt-5.3-codex, got %q", agentRunner.model)
	}
	if repositorySynchronizer.sourceBranch != "develop" {
		t.Fatalf("expected source branch develop, got %q", repositorySynchronizer.sourceBranch)
	}
	if repositorySynchronizer.projectID != "project-1" {
		t.Fatalf("expected project id project-1, got %q", repositorySynchronizer.projectID)
	}
	if len(repositorySynchronizer.sourceRepositories) != 1 || repositorySynchronizer.sourceRepositories[0].RepositoryURL != "https://github.com/acme/source-repo.git" {
		t.Fatalf("unexpected synchronized repositories: %#v", repositorySynchronizer.sourceRepositories)
	}
	if strings.TrimSpace(board.BoardID) != "project_1_ingestion" {
		t.Fatalf("expected normalized board id, got %q", board.BoardID)
	}
	if board.RunID != "run-1" {
		t.Fatalf("expected board run_id to match request, got %q", board.RunID)
	}
	if board.State != domaintracker.BoardStatePending {
		t.Fatalf("expected board state pending, got %q", board.State)
	}
	if board.Epics[0].State != domaintracker.EpicStatePlanned {
		t.Fatalf("expected epic state planned, got %q", board.Epics[0].State)
	}
	if board.Epics[0].Tasks[0].State != domaintracker.TaskStatePlanned {
		t.Fatalf("expected task state planned, got %q", board.Epics[0].Tasks[0].State)
	}
	if boardStore.board.BoardID != board.BoardID {
		t.Fatalf("expected persisted board id %q, got %q", board.BoardID, boardStore.board.BoardID)
	}
}

func TestServiceExecuteClassifiesFetchFailuresAsTransient(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{err: errors.New("gcs timeout")}
	service, err := NewService(boardStore, artifactFetcher, &fakeAgentRunner{}, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if !failures.IsClass(executeErr, failures.ClassTransient) {
		t.Fatalf("expected transient failure class, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
}

func TestServiceExecuteCleansTemporaryArtifacts(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	service, err := NewService(boardStore, artifactFetcher, &fakeAgentRunner{}, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if len(artifactFetcher.destinationPath) != 1 {
		t.Fatalf("expected one destination path, got %d", len(artifactFetcher.destinationPath))
	}
	if _, statErr := os.Stat(artifactFetcher.destinationPath[0]); !os.IsNotExist(statErr) {
		t.Fatalf("expected temporary artifact to be removed, stat err=%v", statErr)
	}
}

func TestServiceExecuteClassifiesRunnerFailuresAsTransient(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	runnerFailure := fmt.Errorf("copilot cli timeout")
	agentRunner := &fakeAgentRunner{err: runnerFailure}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if !failures.IsClass(executeErr, failures.ClassTransient) {
		t.Fatalf("expected transient failure class, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
}

func TestServiceExecuteDefaultsModelAndSourceBranch(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{}
	repositorySynchronizer := &fakeRepositorySynchronizer{}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, repositorySynchronizer)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SourceRepositories:        []SourceRepository{{RepositoryID: "repo-1", RepositoryURL: "https://github.com/acme/source-repo.git"}},
		SystemPrompt:              "You are an ingestion planner.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if agentRunner.model != "gpt-5.3-codex" {
		t.Fatalf("expected default model gpt-5.3-codex, got %q", agentRunner.model)
	}
	if repositorySynchronizer.sourceBranch != "main" {
		t.Fatalf("expected default source branch main, got %q", repositorySynchronizer.sourceBranch)
	}
}

func TestServiceExecuteRetriesUntilValidationPasses(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{writtenBoards: []string{
		`{"board_id":"temporary-board","run_id":"temporary-run","state":"pending","epics":[{"id":"epic-1","board_id":"temporary-board","title":"Epic one","state":"planned","rank":1,"tasks":[{"id":"task-1","board_id":"temporary-board","epic_id":"epic-1","title":"Task one","state":"planned","rank":1}]}],"created_at":"2026-03-03T10:00:00Z","updated_at":"2026-03-03T10:00:00Z"}`,
		`{"board_id":"temporary-board","run_id":"temporary-run","state":"pending","epics":[{"id":"epic-1","board_id":"temporary-board","title":"Epic one","state":"planned","rank":1,"tasks":[{"id":"task-1","board_id":"temporary-board","epic_id":"epic-1","title":"Task one","task_type":"implementation","state":"planned","rank":1}]}],"created_at":"2026-03-03T10:00:00Z","updated_at":"2026-03-03T10:00:00Z"}`,
	}}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if agentRunner.callCount != 2 {
		t.Fatalf("expected 2 generation attempts, got %d", agentRunner.callCount)
	}
}

func TestServiceExecuteFailsAfterMaxValidationAttempts(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{writtenBoard: `{"board_id":"temporary-board","run_id":"temporary-run","state":"pending","epics":[{"id":"epic-1","board_id":"temporary-board","title":"Epic one","state":"planned","rank":1,"tasks":[{"id":"task-1","board_id":"temporary-board","epic_id":"epic-1","title":"Task one","state":"planned","rank":1}]}],"created_at":"2026-03-03T10:00:00Z","updated_at":"2026-03-03T10:00:00Z"}`}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		SystemPrompt:              "You are an ingestion planner.",
		UserPrompt:                "Create a delivery taskboard.",
	})
	if !failures.IsClass(executeErr, failures.ClassTerminal) {
		t.Fatalf("expected terminal validation failure, got %q (%v)", failures.ClassOf(executeErr), executeErr)
	}
	if agentRunner.callCount != maxTaskboardValidationAttempts {
		t.Fatalf("expected %d generation attempts, got %d", maxTaskboardValidationAttempts, agentRunner.callCount)
	}
}

func TestServiceExecuteAllowsPromptOnlyWithoutSelectedDocuments(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:        "run-1",
		ProjectID:    "project-1",
		SystemPrompt: "You are an ingestion planner.",
		UserPrompt:   "Build taskboard from repository context only.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if len(artifactFetcher.fetchedObject) != 0 {
		t.Fatalf("expected no fetched documents for prompt-only ingestion, got %d", len(artifactFetcher.fetchedObject))
	}
}

func TestServiceExecuteAllowsDocsOnlyWithoutUserPrompt(t *testing.T) {
	boardStore := &fakeBoardStore{}
	artifactFetcher := &fakeArtifactFetcher{}
	agentRunner := &fakeAgentRunner{}
	service, err := NewService(boardStore, artifactFetcher, agentRunner, &fakeRepositorySynchronizer{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, executeErr := service.Execute(context.Background(), Request{
		RunID:                     "run-1",
		ProjectID:                 "project-1",
		SelectedDocumentLocations: []string{"projects/project-1/documents/doc-1/spec.md"},
		PreferSelectedDocuments:   true,
		SystemPrompt:              "You are an ingestion planner.",
	})
	if executeErr != nil {
		t.Fatalf("execute ingestion: %v", executeErr)
	}
	if len(artifactFetcher.fetchedObject) != 1 {
		t.Fatalf("expected one fetched document for docs-only ingestion, got %d", len(artifactFetcher.fetchedObject))
	}
	if !strings.Contains(agentRunner.prompt, "Prefer selected remote documents as the primary planning context") {
		t.Fatalf("expected selected-document preference guidance in prompt")
	}
}

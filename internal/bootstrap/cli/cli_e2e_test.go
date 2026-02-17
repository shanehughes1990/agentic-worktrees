package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
	infradatabase "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/database"
	interfacefilesystem "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/filesystem"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
	infraqueue "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue"
	infrarepositories "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/repositories"
	interfaceclicommands "github.com/shanehughes1990/agentic-worktrees/internal/interface/cli/commands"
	"github.com/shanehughes1990/agentic-worktrees/tests/containers"
)

func TestE2EGenerateTaskBoardQueuePipelineWithRedisContainer(t *testing.T) {
	ctx := context.Background()

	redisContainer, releaseRedis, err := containers.AcquireRedis(ctx)
	if err != nil {
		t.Skipf("redis testcontainer unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = releaseRedis(context.Background())
	})

	appLogger, err := infralogger.New("debug", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	databaseURL := "sqlite:///" + filepath.Join(t.TempDir(), "e2e.db")
	dbClient, err := infradatabase.NewGormClient(appLogger, databaseURL)
	if err != nil {
		t.Fatalf("new gorm client: %v", err)
	}

	boardRepository, err := infrarepositories.NewSQLiteBoardRepository(dbClient.DB())
	if err != nil {
		t.Fatalf("new sqlite board repository: %v", err)
	}
	persistCommand, err := application.NewPersistGenerateTaskBoardResultCommand(boardRepository)
	if err != nil {
		t.Fatalf("new persist command: %v", err)
	}
	persistHandler, err := interfaceclicommands.NewPersistGeneratedBoardResultHandler(persistCommand)
	if err != nil {
		t.Fatalf("new persist handler: %v", err)
	}

	queueName := fmt.Sprintf("e2e-generate-%d", time.Now().UnixNano())
	resultQueueName := fmt.Sprintf("e2e-result-%d", time.Now().UnixNano())
	queueClient, err := infraqueue.NewAsynqClient(redisContainer.Addr())
	if err != nil {
		t.Fatalf("new asynq client: %v", err)
	}
	t.Cleanup(func() {
		_ = queueClient.Close()
	})

	generateClient, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, queueName)
	if err != nil {
		t.Fatalf("new generate client: %v", err)
	}
	resultPublisher, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, resultQueueName)
	if err != nil {
		t.Fatalf("new result publisher: %v", err)
	}

	server, err := infraqueue.NewAsynqServerWithQueues(redisContainer.Addr(), map[string]int{queueName: 1, resultQueueName: 1}, 4)
	if err != nil {
		t.Fatalf("new asynq server: %v", err)
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoard, func(handlerCtx context.Context, task *asynq.Task) error {
		if task == nil {
			return fmt.Errorf("task cannot be nil")
		}
		var payload application.GenerateTaskBoardPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		boardJSON := fmt.Sprintf(`{"id":"board-e2e-1","title":"E2E Board","epics":[{"id":"epic-1","title":"Epic 1","description":"E2E Epic","dependencies":[],"tasks":[{"id":"task-1","title":"Task 1","description":"E2E Task","status":"pending","dependencies":[]}]}],"created_at":"%s","updated_at":"%s"}`,
			time.Now().UTC().Format(time.RFC3339),
			time.Now().UTC().Format(time.RFC3339),
		)

		_, err := resultPublisher.EnqueueGenerateTaskBoardResult(handlerCtx, application.GenerateTaskBoardResultMessage{
			Metadata:  payload.Metadata,
			BoardJSON: boardJSON,
		})
		return err
	})
	mux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoardResult, persistHandler.Handle)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(mux)
	}()
	t.Cleanup(func() {
		server.Shutdown()
		select {
		case runErr := <-errCh:
			if runErr != nil {
				t.Logf("server shutdown: %v", runErr)
			}
		case <-time.After(2 * time.Second):
		}
	})

	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", "..", ".."))
	docsPath := filepath.Join(repoRoot, "docs")

	prepareCommand, err := application.NewPrepareGenerateTaskBoardCommand(interfacefilesystem.NewDocumentationLoader())
	if err != nil {
		t.Fatalf("new prepare command: %v", err)
	}
	prepared, err := prepareCommand.Execute(ctx, application.GenerateTaskBoardInput{
		RootDirectory: docsPath,
		MaxDepth:      1,
		Prompt:        "Generate board",
		Model:         "gpt-5.3-codex",
	})
	if err != nil {
		t.Fatalf("prepare payload: %v", err)
	}
	prepared.Metadata = domainservices.AgentRequestMetadata{
		RunID: "run-e2e-1",
		JobID: "job-e2e-1",
		Model: "gpt-5.3-codex",
	}

	if _, err := generateClient.EnqueueGenerateTaskBoard(ctx, prepared); err != nil {
		t.Fatalf("enqueue generate task: %v", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for persisted board")
		}
		_, err := boardRepository.GetByID(ctx, "board-e2e-1")
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
}

type fakeWorkerRunner struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func (r *fakeWorkerRunner) Run(ctx context.Context) error {
	if r == nil || r.server == nil || r.mux == nil {
		return fmt.Errorf("fake worker runner is not initialized")
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.server.Run(r.mux)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		r.server.Shutdown()
		select {
		case err := <-errCh:
			if err == nil {
				return nil
			}
			return err
		case <-time.After(5 * time.Second):
			return nil
		}
	}
}

func TestE2ERuntimeRunGenerateTaskBoardWithRedisContainer(t *testing.T) {
	ctx := context.Background()

	redisContainer, releaseRedis, err := containers.AcquireRedis(ctx)
	if err != nil {
		t.Skipf("redis testcontainer unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = releaseRedis(context.Background())
	})

	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "..", "..", ".."))
	docsPath := filepath.Join(repoRoot, "docs")
	databasePath := filepath.Join(t.TempDir(), "runtime-e2e.db")
	databaseDSN := "sqlite:///" + databasePath
	queueName := fmt.Sprintf("runtime-generate-%d", time.Now().UnixNano())
	resultQueueName := fmt.Sprintf("runtime-result-%d", time.Now().UnixNano())

	t.Setenv("REDIS_ADDR", redisContainer.Addr())
	t.Setenv("ASYNQ_QUEUE", queueName)
	t.Setenv("ASYNQ_RESULT_QUEUE", resultQueueName)
	t.Setenv("ASYNQ_CONCURRENCY", "4")
	t.Setenv("ASYNQ_WAIT_TIMEOUT", "45s")
	t.Setenv("COPILOT_START_TIMEOUT", "5s")
	t.Setenv("DATABASE_DSN", databaseDSN)
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")

	runtimeInstance, err := New()
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	appLogger, err := infralogger.New("debug", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	dbClient, err := infradatabase.NewGormClient(appLogger, databaseDSN)
	if err != nil {
		t.Fatalf("new gorm client: %v", err)
	}
	boardRepository, err := infrarepositories.NewSQLiteBoardRepository(dbClient.DB())
	if err != nil {
		t.Fatalf("new sqlite board repository: %v", err)
	}
	persistCommand, err := application.NewPersistGenerateTaskBoardResultCommand(boardRepository)
	if err != nil {
		t.Fatalf("new persist command: %v", err)
	}
	persistHandler, err := interfaceclicommands.NewPersistGeneratedBoardResultHandler(persistCommand)
	if err != nil {
		t.Fatalf("new persist handler: %v", err)
	}

	queueClient, err := infraqueue.NewAsynqClient(redisContainer.Addr())
	if err != nil {
		t.Fatalf("new asynq client: %v", err)
	}
	t.Cleanup(func() {
		_ = queueClient.Close()
	})
	resultPublisher, err := infraqueue.NewAsynqGenerateTaskBoardClient(queueClient, resultQueueName)
	if err != nil {
		t.Fatalf("new result publisher: %v", err)
	}
	workerServer, err := infraqueue.NewAsynqServerWithQueues(redisContainer.Addr(), map[string]int{queueName: 1, resultQueueName: 1}, 4)
	if err != nil {
		t.Fatalf("new worker server: %v", err)
	}
	workerMux := asynq.NewServeMux()
	workerMux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoard, func(handlerCtx context.Context, task *asynq.Task) error {
		if task == nil {
			return fmt.Errorf("task cannot be nil")
		}
		var payload application.GenerateTaskBoardPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		boardJSON := fmt.Sprintf(`{"id":"board-runtime-e2e-1","title":"Runtime E2E Board","epics":[{"id":"epic-1","title":"Epic 1","description":"Runtime E2E Epic","dependencies":[],"tasks":[{"id":"task-1","title":"Task 1","description":"Runtime E2E Task","status":"pending","dependencies":[]}]}],"created_at":"%s","updated_at":"%s"}`,
			time.Now().UTC().Format(time.RFC3339),
			time.Now().UTC().Format(time.RFC3339),
		)
		_, err := resultPublisher.EnqueueGenerateTaskBoardResult(handlerCtx, application.GenerateTaskBoardResultMessage{
			Metadata:  payload.Metadata,
			BoardJSON: boardJSON,
		})
		return err
	})
	workerMux.HandleFunc(application.AsynqTaskTypeGenerateTaskBoardResult, persistHandler.Handle)

	runtimeInstance.workerRunner = &fakeWorkerRunner{server: workerServer, mux: workerMux}
	if runtimeInstance.workerErrorCh == nil {
		runtimeInstance.workerErrorCh = make(chan error, 1)
	}

	originalArgs := os.Args
	os.Args = []string{
		"cli",
		"generate-task-board",
		"--ROOT_DIRECTORY=" + docsPath,
		"--MAX_DEPTH=1",
		"--PROMPT=Generate board",
		"--MODEL=gpt-5.3-codex",
	}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	if err := runtimeInstance.Run(ctx); err != nil {
		t.Fatalf("runtime run: %v", err)
	}

	storedBoard, err := boardRepository.GetByID(ctx, "board-runtime-e2e-1")
	if err != nil {
		t.Fatalf("get persisted board: %v", err)
	}
	if storedBoard.ID != "board-runtime-e2e-1" {
		t.Fatalf("unexpected board id: %s", storedBoard.ID)
	}
}

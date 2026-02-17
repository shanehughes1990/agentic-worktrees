package consumer

import (
	"context"
	errorspkg "errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
	checkpointstore "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/store"
	queuedomain "github.com/shanehughes1990/agentic-worktrees/internal/features/queue/domain"
	sharederrors "github.com/shanehughes1990/agentic-worktrees/internal/shared/errors"
)

type fakeIngestionService struct {
	board boarddomain.Board
	err   error
}

func (f fakeIngestionService) Plan(_ context.Context, _ string) (boarddomain.Board, error) {
	if f.err != nil {
		return boarddomain.Board{}, f.err
	}
	return f.board, nil
}

func TestProcessPlanBoardTaskSuccess(t *testing.T) {
	checkpointPath := filepath.Join(t.TempDir(), "checkpoints.json")
	handler := &Handler{
		IngestionService: fakeIngestionService{board: boarddomain.Board{
			SchemaVersion: 1,
			SourceScope:   "docs",
			GeneratedAt:   time.Now().UTC(),
			Epics: []boarddomain.Epic{{
				ID: "epic-001", Title: "first", Tasks: []boarddomain.Task{{
					ID: "task-001", Title: "task", Lane: "lane-a", Status: boarddomain.TaskStatusPending,
				}},
			}},
		}},
		CheckpointStore: checkpointstore.NewJSONStore(checkpointPath),
	}

	task, err := queuedomain.NewPlanBoardTask(queuedomain.PlanBoardPayload{
		RunID:     "run-1",
		TaskID:    "task-1",
		ScopePath: "docs",
		OutPath:   filepath.Join(t.TempDir(), "board.json"),
	}, "default")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	if err := handler.ProcessPlanBoardTask(context.Background(), task); err != nil {
		t.Fatalf("process task: %v", err)
	}
}

func TestProcessPlanBoardTaskTerminalErrorSkipsRetry(t *testing.T) {
	handler := &Handler{
		IngestionService: fakeIngestionService{err: sharederrors.Terminal("adk", errorspkg.New("bad request"))},
	}

	task, err := queuedomain.NewPlanBoardTask(queuedomain.PlanBoardPayload{
		RunID:     "run-1",
		TaskID:    "task-1",
		ScopePath: "docs",
		OutPath:   filepath.Join(t.TempDir(), "board.json"),
	}, "default")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}

	err = handler.ProcessPlanBoardTask(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errorspkg.Is(err, asynq.SkipRetry) {
		t.Fatalf("expected asynq.SkipRetry, got %v", err)
	}
}

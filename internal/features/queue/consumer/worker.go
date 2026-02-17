package consumer

import (
	"context"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	boarddomain "github.com/shanehughes1990/agentic-worktrees/internal/features/board/domain"
	boardstore "github.com/shanehughes1990/agentic-worktrees/internal/features/board/store"
	checkpointdomain "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/domain"
	checkpointstore "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/store"
	queuedomain "github.com/shanehughes1990/agentic-worktrees/internal/features/queue/domain"
	sharederrors "github.com/shanehughes1990/agentic-worktrees/internal/shared/errors"
)

type Planner interface {
	Plan(ctx context.Context, scopePath string) (boarddomain.Board, error)
}

type Handler struct {
	IngestionService Planner
	CheckpointStore  *checkpointstore.JSONStore
	Logger           *logrus.Logger
}

func (h *Handler) ProcessPlanBoardTask(ctx context.Context, task *asynq.Task) error {
	payload, err := queuedomain.ParsePlanBoardPayload(task)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	h.logFields(payload).Info("processing plan board task")
	h.checkpoint(payload, "executing_adk", "")

	board, err := h.IngestionService.Plan(ctx, payload.ScopePath)
	if err != nil {
		class := sharederrors.ClassOf(err)
		h.checkpoint(payload, "failed", err.Error())
		h.logFields(payload).WithError(err).Error("plan board task failed")
		if class == sharederrors.ClassTerminal {
			return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
		}
		return err
	}

	writer := boardstore.NewJSONStore(payload.OutPath)
	if err := writer.Write(board); err != nil {
		h.checkpoint(payload, "failed", err.Error())
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	h.checkpoint(payload, "succeeded", "")
	h.logFields(payload).Info("plan board task succeeded")
	return nil
}

func (h *Handler) checkpoint(payload queuedomain.PlanBoardPayload, state string, errText string) {
	if h.CheckpointStore == nil {
		return
	}
	_ = h.CheckpointStore.Append(checkpointdomain.Record{
		RunID:  payload.RunID,
		TaskID: payload.TaskID,
		JobID:  payload.IdempotencyKey,
		State:  state,
		Error:  strings.TrimSpace(errText),
	})
}

func (h *Handler) logFields(payload queuedomain.PlanBoardPayload) *logrus.Entry {
	logger := h.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return logger.WithFields(logrus.Fields{
		"run_id":  payload.RunID,
		"task_id": payload.TaskID,
		"job_id":  payload.IdempotencyKey,
	})
}

type Worker struct {
	server  *asynq.Server
	handler *Handler
}

func NewWorker(redisAddr string, queueName string, concurrency int, handler *Handler) (*Worker, error) {
	if strings.TrimSpace(redisAddr) == "" {
		return nil, fmt.Errorf("redis address cannot be empty")
	}
	if strings.TrimSpace(queueName) == "" {
		queueName = "default"
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if handler == nil || handler.IngestionService == nil {
		return nil, fmt.Errorf("handler and ingestion service are required")
	}

	server := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			queueName: 1,
		},
	})

	return &Worker{server: server, handler: handler}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(queuedomain.TypePlanBoard, w.handler.ProcessPlanBoardTask)

	errCh := make(chan error, 1)
	go func() {
		errCh <- w.server.Run(mux)
	}()

	select {
	case <-ctx.Done():
		w.server.Shutdown()
		return nil
	case err := <-errCh:
		return err
	}
}

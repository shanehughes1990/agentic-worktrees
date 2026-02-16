package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hibiken/asynq"

	"github.com/shanehughes1990/agentic-worktrees/internal/app"
	"github.com/shanehughes1990/agentic-worktrees/internal/gitops"
	"github.com/shanehughes1990/agentic-worktrees/internal/runstate"
)

type Worker struct {
	server      *asynq.Server
	client      *asynq.Client
	audit       *app.AuditSink
	checkpoints *runstate.Store
	git         *gitops.Manager
	metrics     *Metrics
	queueName   string
}

func NewWorker(redisAddr string, queueName string, concurrency int, audit *app.AuditSink, checkpoints *runstate.Store, git *gitops.Manager, metrics *Metrics) *Worker {
	if concurrency < 1 {
		concurrency = 1
	}
	if queueName == "" {
		queueName = "default"
	}
	if metrics == nil {
		metrics = NewMetrics()
	}
	if checkpoints == nil {
		checkpoints = runstate.NewStore("state/checkpoints.json")
	}
	if git == nil {
		git = gitops.NewManager(".worktrees")
	}

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				queueName: 1,
			},
		},
	)
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	return &Worker{server: server, client: client, audit: audit, checkpoints: checkpoints, git: git, metrics: metrics, queueName: queueName}
}

func (w *Worker) Run(ctx context.Context) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypePrepareWorktree, w.handlePrepareWorktree)
	mux.HandleFunc(TypeExecuteAgent, w.handleExecuteAgent)
	mux.HandleFunc(TypeValidate, w.handleValidate)
	mux.HandleFunc(TypeOpenOrUpdatePR, w.handleOpenOrUpdatePR)
	mux.HandleFunc(TypeRebaseAndMerge, w.handleRebaseAndMerge)
	mux.HandleFunc(TypeCleanup, w.handleCleanup)

	go func() {
		<-ctx.Done()
		w.server.Shutdown()
		_ = w.client.Close()
	}()

	return w.server.Run(mux)
}

func (w *Worker) handlePrepareWorktree(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypePrepareWorktree, TypeExecuteAgent)
}

func (w *Worker) handleExecuteAgent(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypeExecuteAgent, TypeValidate)
}

func (w *Worker) handleValidate(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypeValidate, TypeOpenOrUpdatePR)
}

func (w *Worker) handleOpenOrUpdatePR(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypeOpenOrUpdatePR, TypeRebaseAndMerge)
}

func (w *Worker) handleRebaseAndMerge(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypeRebaseAndMerge, TypeCleanup)
}

func (w *Worker) handleCleanup(ctx context.Context, task *asynq.Task) error {
	return w.processPhase(ctx, task, TypeCleanup, "")
}

func (w *Worker) processPhase(ctx context.Context, task *asynq.Task, current string, next string) error {
	payload, err := ParseLifecycleTaskPayload(task)
	if err != nil {
		w.metrics.MarkFailure()
		return err
	}

	w.metrics.MarkStart()
	if err := w.writeCheckpoint(payload, current, "running"); err != nil {
		w.metrics.MarkFailure()
		return err
	}

	startedData, _ := json.Marshal(map[string]string{
		"phase":    current,
		"worktree": payload.WorktreeName,
		"prompt":   payload.Prompt,
	})
	if w.audit != nil {
		if err := w.audit.Write(ctx, app.AuditEvent{Type: "task.phase.started", TaskID: payload.TaskID, Data: startedData}); err != nil {
			w.metrics.MarkFailure()
			return fmt.Errorf("audit start event: %w", err)
		}
	}

	switch current {
	case TypePrepareWorktree:
		if _, err := w.git.EnsureWorktree(payload.WorktreeName); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
	case TypeExecuteAgent:
		if err := appendWorktreeLog(payload.WorktreeName, fmt.Sprintf("%s execute_agent task=%s", time.Now().Format(time.RFC3339), payload.TaskID)); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
		time.Sleep(500 * time.Millisecond)
	case TypeValidate:
		if err := appendWorktreeLog(payload.WorktreeName, fmt.Sprintf("%s validate task=%s", time.Now().Format(time.RFC3339), payload.TaskID)); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
	case TypeOpenOrUpdatePR:
		if err := appendWorktreeLog(payload.WorktreeName, fmt.Sprintf("%s open_or_update_pr task=%s", time.Now().Format(time.RFC3339), payload.TaskID)); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
	case TypeRebaseAndMerge:
		if err := w.git.MarkMerged(payload.RunID, payload.TaskID, payload.OriginBranch); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
	case TypeCleanup:
		if err := appendWorktreeLog(payload.WorktreeName, fmt.Sprintf("%s cleanup task=%s", time.Now().Format(time.RFC3339), payload.TaskID)); err != nil {
			w.metrics.MarkFailure()
			_ = w.writeCheckpoint(payload, current, "failed")
			return err
		}
	}

	if err := w.writeCheckpoint(payload, current, "completed"); err != nil {
		w.metrics.MarkFailure()
		return err
	}

	if w.audit != nil {
		if err := w.audit.Write(ctx, app.AuditEvent{Type: "task.phase.completed", TaskID: payload.TaskID, Message: current}); err != nil {
			w.metrics.MarkFailure()
			return fmt.Errorf("audit complete event: %w", err)
		}
	}

	if next != "" {
		nextTask, err := NewLifecycleTask(next, payload, w.queueName)
		if err != nil {
			w.metrics.MarkFailure()
			return err
		}
		if _, err := w.client.Enqueue(nextTask); err != nil {
			w.metrics.MarkFailure()
			return err
		}
	}

	w.metrics.MarkSuccess()
	return nil
}

func (w *Worker) writeCheckpoint(payload LifecyclePayload, phase string, status string) error {
	return w.checkpoints.Upsert(runstate.Checkpoint{
		RunID:        payload.RunID,
		TaskID:       payload.TaskID,
		Phase:        phase,
		Status:       status,
		OriginBranch: payload.OriginBranch,
	})
}

func appendWorktreeLog(worktreeName string, line string) error {
	if worktreeName == "" {
		worktreeName = "unknown-worktree"
	}
	if err := os.MkdirAll("logs/worktrees", 0o755); err != nil {
		return err
	}
	path := filepath.Join("logs/worktrees", worktreeName+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

type Client struct {
	client *asynq.Client
}

func NewClient(redisAddr string) *Client {
	return &Client{client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Enqueue(task *asynq.Task) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task)
}

package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/infrastructure/observability"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hibiken/asynq"
)

type Platform struct {
	config Config
	entry  *observability.Entry

	client *asynq.Client
	server *asynq.Server
	mux    *asynq.ServeMux

	deadLetterAudit taskengine.DeadLetterAudit
	started         bool
	mu              sync.Mutex
}

func NewPlatform(config Config, entry *observability.Entry) *Platform {
	normalizedConfig := config.normalized()
	redisOptions := asynq.RedisClientOpt{
		Addr:     normalizedConfig.RedisAddress,
		Password: normalizedConfig.RedisPassword,
		DB:       normalizedConfig.RedisDatabase,
	}
	return &Platform{
		config: normalizedConfig,
		entry:  entry,
		client: asynq.NewClient(redisOptions),
		server: asynq.NewServer(redisOptions, asynq.Config{Concurrency: normalizedConfig.Concurrency}),
		mux:    asynq.NewServeMux(),
	}
}

func (platform *Platform) SetDeadLetterAudit(audit taskengine.DeadLetterAudit) {
	if platform == nil {
		return
	}
	platform.deadLetterAudit = audit
}

func (platform *Platform) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	if platform == nil || platform.client == nil {
		return taskengine.EnqueueResult{}, fmt.Errorf("task engine platform is not initialized")
	}
	task := asynq.NewTask(string(request.Kind), request.Payload)
	options := make([]asynq.Option, 0, 5)
	if strings.TrimSpace(request.Queue) != "" {
		options = append(options, asynq.Queue(request.Queue))
	}
	if strings.TrimSpace(request.IdempotencyKey) != "" {
		options = append(options, asynq.TaskID(request.IdempotencyKey))
	}
	if request.UniqueFor > 0 {
		options = append(options, asynq.Unique(request.UniqueFor))
	}
	if request.Timeout > 0 {
		options = append(options, asynq.Timeout(request.Timeout))
	}
	options = append(options, asynq.MaxRetry(request.MaxRetry))

	taskInfo, enqueueErr := platform.client.EnqueueContext(ctx, task, options...)
	if enqueueErr != nil {
		if errors.Is(enqueueErr, asynq.ErrDuplicateTask) || errors.Is(enqueueErr, asynq.ErrTaskIDConflict) {
			platform.logWarn("duplicate enqueue suppressed", map[string]any{
				"job_kind":        request.Kind,
				"idempotency_key": request.IdempotencyKey,
				"queue":           request.Queue,
			})
			return taskengine.EnqueueResult{QueueTaskID: request.IdempotencyKey, Duplicate: true}, nil
		}
		return taskengine.EnqueueResult{}, fmt.Errorf("enqueue job: %w", enqueueErr)
	}
	if taskInfo == nil {
		return taskengine.EnqueueResult{}, fmt.Errorf("enqueue job: missing task info")
	}
	return taskengine.EnqueueResult{QueueTaskID: taskInfo.ID}, nil
}

func (platform *Platform) Register(kind taskengine.JobKind, handler taskengine.Handler) error {
	if platform == nil || platform.mux == nil {
		return fmt.Errorf("task engine platform is not initialized")
	}
	if strings.TrimSpace(string(kind)) == "" {
		return fmt.Errorf("job kind is required")
	}
	if handler == nil {
		return fmt.Errorf("handler is required")
	}

	platform.mux.HandleFunc(string(kind), func(ctx context.Context, task *asynq.Task) error {
		queueTaskID, _ := asynq.GetTaskID(ctx)
		return handler.Handle(ctx, taskengine.Job{
			Kind:        kind,
			QueueTaskID: strings.TrimSpace(queueTaskID),
			Payload:     task.Payload(),
		})
	})

	return nil
}

func (platform *Platform) Start() error {
	if platform == nil || platform.server == nil || platform.mux == nil {
		return fmt.Errorf("task engine platform is not initialized")
	}

	platform.mu.Lock()
	if platform.started {
		platform.mu.Unlock()
		return nil
	}
	startErr := platform.server.Start(platform.mux)
	if startErr == nil {
		platform.started = true
	}
	platform.mu.Unlock()
	if startErr != nil {
		return fmt.Errorf("start asynq worker: %w", startErr)
	}

	platform.logInfo("asynq worker started", map[string]any{
		"concurrency":   platform.config.Concurrency,
		"redis_address": platform.config.RedisAddress,
	})
	return nil
}

func (platform *Platform) Shutdown(ctx context.Context) error {
	if platform == nil {
		return nil
	}
	_ = ctx

	platform.mu.Lock()
	started := platform.started
	platform.started = false
	platform.mu.Unlock()

	var shutdownErr error
	if started && platform.server != nil {
		platform.server.Shutdown()
	}
	if platform.client != nil {
		if closeErr := platform.client.Close(); closeErr != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("close asynq client: %w", closeErr))
		}
	}
	return shutdownErr
}

func (platform *Platform) logInfo(message string, fields map[string]any) {
	if platform == nil || platform.entry == nil {
		return
	}
	platform.entry.WithFields(fields).Info(message)
}

func (platform *Platform) logWarn(message string, fields map[string]any) {
	if platform == nil || platform.entry == nil {
		return
	}
	platform.entry.WithFields(fields).Warn(message)
}

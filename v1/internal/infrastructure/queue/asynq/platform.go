package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	domainobservability "agentic-orchestrator/internal/domain/shared/observability"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hibiken/asynq"
)

type APIPlatform struct {
	entry    domainobservability.Entry
	redisURL string

	client *asynq.Client

	deadLetterAudit taskengine.DeadLetterAudit
}

type WorkerPlatform struct {
	entry       domainobservability.Entry
	redisURL    string
	concurrency int

	client *asynq.Client
	server *asynq.Server
	mux    *asynq.ServeMux

	deadLetterAudit taskengine.DeadLetterAudit
	started         bool
	mu              sync.Mutex
}

func NewAPIPlatform(config APIConfig, entry domainobservability.Entry) (*APIPlatform, error) {
	normalizedConfig := config.normalized()
	redisConnOpt, err := normalizedConfig.redisClientOpt()
	if err != nil {
		return nil, err
	}
	return &APIPlatform{
		entry:    entry,
		redisURL: normalizedConfig.RedisURL,
		client:   asynq.NewClient(redisConnOpt),
	}, nil
}

func NewWorkerPlatform(config WorkerConfig, entry domainobservability.Entry) (*WorkerPlatform, error) {
	normalizedConfig := config.normalized()
	redisConnOpt, err := normalizedConfig.redisClientOpt()
	if err != nil {
		return nil, err
	}
	return &WorkerPlatform{
		entry:       entry,
		redisURL:    normalizedConfig.RedisURL,
		concurrency: normalizedConfig.Concurrency,
		client:      asynq.NewClient(redisConnOpt),
		server:      asynq.NewServer(redisConnOpt, asynq.Config{Concurrency: normalizedConfig.Concurrency, Queues: normalizedConfig.Queues}),
		mux:         asynq.NewServeMux(),
	}, nil
}

func (platform *APIPlatform) SetDeadLetterAudit(audit taskengine.DeadLetterAudit) {
	if platform == nil {
		return
	}
	platform.deadLetterAudit = audit
}

func (platform *WorkerPlatform) SetDeadLetterAudit(audit taskengine.DeadLetterAudit) {
	if platform == nil {
		return
	}
	platform.deadLetterAudit = audit
}

func (platform *APIPlatform) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
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

func (platform *WorkerPlatform) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
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

func (platform *WorkerPlatform) Register(kind taskengine.JobKind, handler taskengine.Handler) error {
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

func (platform *WorkerPlatform) Start() error {
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
		"concurrency": platform.concurrency,
		"redis_url":   platform.redisURL,
	})
	return nil
}

func (platform *APIPlatform) Shutdown(ctx context.Context) error {
	if platform == nil {
		return nil
	}
	_ = ctx

	if platform.client != nil {
		if closeErr := platform.client.Close(); closeErr != nil {
			return fmt.Errorf("close asynq client: %w", closeErr)
		}
	}
	return nil
}

func (platform *WorkerPlatform) Shutdown(ctx context.Context) error {
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

func (platform *WorkerPlatform) logInfo(message string, fields map[string]any) {
	if platform == nil || platform.entry == nil {
		return
	}
	platform.entry.WithFields(fields).Info(message)
}

func (platform *APIPlatform) logWarn(message string, fields map[string]any) {
	if platform == nil || platform.entry == nil {
		return
	}
	platform.entry.WithFields(fields).Warn(message)
}

func (platform *WorkerPlatform) logWarn(message string, fields map[string]any) {
	if platform == nil || platform.entry == nil {
		return
	}
	platform.entry.WithFields(fields).Warn(message)
}

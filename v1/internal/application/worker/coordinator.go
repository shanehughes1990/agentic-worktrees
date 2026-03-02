package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"agentic-orchestrator/internal/application/taskengine"
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
)

type Coordinator struct {
	service *Service
	engine  taskengine.Engine
}

func NewCoordinator(service *Service, engine taskengine.Engine) (*Coordinator, error) {
	if service == nil {
		return nil, ErrRepositoryRequired
	}
	if engine == nil {
		return nil, fmt.Errorf("task engine is required")
	}
	return &Coordinator{service: service, engine: engine}, nil
}

func (coordinator *Coordinator) ProbeAndEscalate(ctx context.Context) error {
	if coordinator == nil || coordinator.service == nil || coordinator.engine == nil {
		return fmt.Errorf("worker coordinator is not initialized")
	}
	now := time.Now().UTC()
	settings, err := coordinator.service.GetSettings(ctx)
	if err != nil {
		return err
	}
	staleWorkers, err := coordinator.service.repository.ListStaleWorkers(ctx, now.Add(-settings.StaleAfter), 500)
	if err != nil {
		return err
	}
	for _, worker := range staleWorkers {
		reason := strings.TrimSpace(worker.RogueReason)
		if reason == "" {
			reason = "heartbeat timeout"
		}
		if _, updateErr := coordinator.service.RequestShutdown(ctx, worker.WorkerID, worker.Epoch, reason); updateErr != nil {
			continue
		}
		_ = coordinator.enqueueShutdownTasks(ctx, worker, reason)
	}
	return nil
}

func (coordinator *Coordinator) enqueueShutdownTasks(ctx context.Context, worker domainrealtime.Worker, reason string) error {
	if err := coordinator.enqueueShutdownTask(ctx, taskengine.JobKindWorkerShutdownAgent, worker, reason); err != nil {
		return err
	}
	if err := coordinator.enqueueShutdownTask(ctx, taskengine.JobKindWorkerShutdownRuntime, worker, reason); err != nil {
		return err
	}
	if err := coordinator.enqueueShutdownTask(ctx, taskengine.JobKindWorkerForceDeregister, worker, reason); err != nil {
		return err
	}
	return nil
}

func (coordinator *Coordinator) enqueueShutdownTask(ctx context.Context, kind taskengine.JobKind, worker domainrealtime.Worker, reason string) error {
	payload, err := json.Marshal(map[string]any{
		"worker_id": worker.WorkerID,
		"epoch":     worker.Epoch,
		"reason":    reason,
	})
	if err != nil {
		return fmt.Errorf("marshal shutdown payload: %w", err)
	}
	_, err = coordinator.engine.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           kind,
		Payload:        payload,
		IdempotencyKey: fmt.Sprintf("%s:%s:%d", kind, worker.WorkerID, worker.Epoch),
		CorrelationIDs: taskengine.CorrelationIDs{RunID: worker.WorkerID, TaskID: "worker-shutdown", JobID: fmt.Sprintf("%s-%d", worker.WorkerID, worker.Epoch)},
	})
	if err != nil {
		return fmt.Errorf("enqueue shutdown task %s: %w", kind, err)
	}
	return nil
}

package agent

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"fmt"
	"strings"
)

type LoopbackRemoteWorkerAdapter struct {
	workerID string
	handler  taskengine.Handler
}

func NewLoopbackRemoteWorkerAdapter(workerID string, handler taskengine.Handler) (*LoopbackRemoteWorkerAdapter, error) {
	if strings.TrimSpace(workerID) == "" {
		return nil, fmt.Errorf("worker_id is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler is required")
	}
	return &LoopbackRemoteWorkerAdapter{workerID: workerID, handler: handler}, nil
}

func (adapter *LoopbackRemoteWorkerAdapter) Execute(ctx context.Context, request taskengine.RemoteExecutionRequest) (taskengine.RemoteExecutionResult, error) {
	if adapter == nil || adapter.handler == nil {
		return taskengine.RemoteExecutionResult{}, fmt.Errorf("remote worker adapter is not initialized")
	}
	if err := request.Validate(); err != nil {
		return taskengine.RemoteExecutionResult{}, err
	}
	if err := adapter.handler.Handle(ctx, request.Job); err != nil {
		return taskengine.RemoteExecutionResult{}, err
	}
	result := taskengine.RemoteExecutionResult{WorkerID: adapter.workerID}
	if request.ResumeCheckpoint != nil {
		result.CompletedCheckpoint = &taskengine.RemoteCheckpoint{
			Step:  request.ResumeCheckpoint.Step,
			Token: request.ResumeCheckpoint.Token,
		}
	}
	if err := result.Validate(); err != nil {
		return taskengine.RemoteExecutionResult{}, err
	}
	return result, nil
}

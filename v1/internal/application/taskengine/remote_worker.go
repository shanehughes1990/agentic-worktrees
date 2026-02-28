package taskengine

import (
	"context"
	"fmt"
	"strings"
)

type RemoteCheckpoint struct {
	Step  string
	Token string
}

func (checkpoint RemoteCheckpoint) Validate() error {
	if strings.TrimSpace(checkpoint.Step) == "" {
		return fmt.Errorf("%w: checkpoint step is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(checkpoint.Token) == "" {
		return fmt.Errorf("%w: checkpoint token is required", ErrInvalidRemoteExecutionRequest)
	}
	return nil
}

type RemoteExecutionRequest struct {
	Job              Job
	CorrelationIDs   CorrelationIDs
	IdempotencyKey   string
	ResumeCheckpoint *RemoteCheckpoint
}

func (request RemoteExecutionRequest) Validate() error {
	if strings.TrimSpace(string(request.Job.Kind)) == "" {
		return fmt.Errorf("%w: job kind is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.Job.QueueTaskID) == "" {
		return fmt.Errorf("%w: queue_task_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if len(request.Job.Payload) == 0 {
		return fmt.Errorf("%w: payload is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.RunID) == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.TaskID) == "" {
		return fmt.Errorf("%w: task_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.CorrelationIDs.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidRemoteExecutionRequest)
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidRemoteExecutionRequest)
	}
	if request.ResumeCheckpoint != nil {
		if err := request.ResumeCheckpoint.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type RemoteExecutionResult struct {
	WorkerID            string
	CompletedCheckpoint *RemoteCheckpoint
}

type RemoteWorkerAdapter interface {
	Execute(ctx context.Context, request RemoteExecutionRequest) (RemoteExecutionResult, error)
}

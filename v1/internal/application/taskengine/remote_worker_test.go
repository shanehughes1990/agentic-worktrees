package taskengine

import (
	"errors"
	"testing"
)

func validRemoteExecutionRequest() RemoteExecutionRequest {
	return RemoteExecutionRequest{
		Job: Job{
			Kind:        JobKindSCMWorkflow,
			QueueTaskID: "queue-task-1",
			Payload:     []byte(`{"operation":"source_state"}`),
		},
		CorrelationIDs: CorrelationIDs{
			RunID:  "run-1",
			TaskID: "task-1",
			JobID:  "job-1",
		},
		IdempotencyKey: "idempotency-1",
	}
}

func TestRemoteExecutionRequestValidateAcceptsValidRequest(t *testing.T) {
	request := validRemoteExecutionRequest()
	if err := request.Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestRemoteExecutionRequestValidateRejectsMissingQueueTaskID(t *testing.T) {
	request := validRemoteExecutionRequest()
	request.Job.QueueTaskID = ""
	err := request.Validate()
	if !errors.Is(err, ErrInvalidRemoteExecutionRequest) {
		t.Fatalf("expected ErrInvalidRemoteExecutionRequest, got %v", err)
	}
}

func TestRemoteExecutionRequestValidateRejectsMissingIdempotencyKey(t *testing.T) {
	request := validRemoteExecutionRequest()
	request.IdempotencyKey = ""
	err := request.Validate()
	if !errors.Is(err, ErrInvalidRemoteExecutionRequest) {
		t.Fatalf("expected ErrInvalidRemoteExecutionRequest, got %v", err)
	}
}

func TestRemoteExecutionRequestValidateRejectsInvalidCheckpoint(t *testing.T) {
	request := validRemoteExecutionRequest()
	request.ResumeCheckpoint = &RemoteCheckpoint{Step: "source_state"}
	err := request.Validate()
	if !errors.Is(err, ErrInvalidRemoteExecutionRequest) {
		t.Fatalf("expected ErrInvalidRemoteExecutionRequest, got %v", err)
	}
}

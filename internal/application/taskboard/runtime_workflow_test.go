package taskboard

import (
	"context"
	"errors"
	"testing"
)

type runtimeWorkflowRepositoryStub struct {
	listWorkflowsResult []IngestionWorkflow
	getWorkflowResult   *IngestionWorkflow
	cancelResult        WorkflowCancelResult
	cancelErr           error
	lastCanceledRunID   string
}

func (stub *runtimeWorkflowRepositoryStub) ListRuntimeWorkflows(ctx context.Context) ([]IngestionWorkflow, error) {
	return stub.listWorkflowsResult, nil
}

func (stub *runtimeWorkflowRepositoryStub) GetRuntimeWorkflow(ctx context.Context, runID string) (*IngestionWorkflow, error) {
	return stub.getWorkflowResult, nil
}

func (stub *runtimeWorkflowRepositoryStub) CancelRuntimeWorkflow(ctx context.Context, runID string) (WorkflowCancelResult, error) {
	stub.lastCanceledRunID = runID
	if stub.cancelErr != nil {
		return WorkflowCancelResult{}, stub.cancelErr
	}
	return stub.cancelResult, nil
}

func TestRuntimeWorkflowServiceCancelWorkflow(t *testing.T) {
	repository := &runtimeWorkflowRepositoryStub{cancelResult: WorkflowCancelResult{RunID: "run-1", MatchedTasks: 2, CanceledTasks: 2}}
	service := NewRuntimeWorkflowService(repository)

	result, err := service.CancelWorkflow(context.Background(), " run-1 ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repository.lastCanceledRunID != "run-1" {
		t.Fatalf("expected trimmed run id to be passed, got %q", repository.lastCanceledRunID)
	}
	if result.RunID != "run-1" || result.CanceledTasks != 2 {
		t.Fatalf("unexpected cancel result: %+v", result)
	}
}

func TestRuntimeWorkflowServiceCancelWorkflowRequiresRunID(t *testing.T) {
	service := NewRuntimeWorkflowService(&runtimeWorkflowRepositoryStub{})

	_, err := service.CancelWorkflow(context.Background(), "   ")
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestRuntimeWorkflowServiceCancelWorkflowWrapsError(t *testing.T) {
	repository := &runtimeWorkflowRepositoryStub{cancelErr: errors.New("queue unavailable")}
	service := NewRuntimeWorkflowService(repository)

	_, err := service.CancelWorkflow(context.Background(), "run-2")
	if err == nil {
		t.Fatalf("expected wrapped error")
	}
}

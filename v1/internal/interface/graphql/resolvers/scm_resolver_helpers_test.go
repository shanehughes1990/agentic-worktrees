package resolvers

import (
	"agentic-orchestrator/internal/application/taskengine"
	"agentic-orchestrator/internal/interface/graphql/models"
	"context"
	"testing"
)

type fakeEngine struct {
	result taskengine.EnqueueResult
	err    error
}

func (engine *fakeEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	_ = request
	return engine.result, engine.err
}

func TestEnqueueSCMWorkflow(t *testing.T) {
	scheduler, err := taskengine.NewScheduler(&fakeEngine{result: taskengine.EnqueueResult{QueueTaskID: "q-1"}}, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	resolver := NewResolver(scheduler, nil, nil, nil, nil)

	result, enqueueErr := enqueueSCMWorkflow(context.Background(), resolver, models.EnqueueSCMWorkflowInput{
		Operation:      models.SCMOperationEnsureWorktree,
		Provider:       models.SCMProviderGithub,
		Owner:          "acme",
		Repository:     "repo",
		RunID:          "run-1",
		TaskID:         "task-1",
		JobID:          "job-1",
		IdempotencyKey: "id-1",
	})
	if enqueueErr != nil {
		t.Fatalf("enqueue scm workflow: %v", enqueueErr)
	}
	success, ok := result.(models.EnqueueSCMWorkflowSuccess)
	if !ok {
		t.Fatalf("expected EnqueueSCMWorkflowSuccess, got %T", result)
	}
	if success.QueueTaskID != "q-1" {
		t.Fatalf("expected queue task id q-1, got %q", success.QueueTaskID)
	}
}

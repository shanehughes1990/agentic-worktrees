package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"errors"
	"testing"
	"time"
)

type fakePromptRefinementRepository struct {
	requests      map[string]PromptRefinementRequest
	readyPrompt   string
	getCallCount  int
	markReadyCall int
}

func (repository *fakePromptRefinementRepository) CreatePromptRefinementRequest(ctx context.Context, request PromptRefinementRequest) (*PromptRefinementRequest, error) {
	_ = ctx
	if repository.requests == nil {
		repository.requests = map[string]PromptRefinementRequest{}
	}
	repository.requests[request.RequestID] = request
	copy := request
	return &copy, nil
}

func (repository *fakePromptRefinementRepository) GetPromptRefinementRequest(ctx context.Context, requestID string) (*PromptRefinementRequest, error) {
	_ = ctx
	repository.getCallCount++
	request, exists := repository.requests[requestID]
	if !exists {
		return nil, nil
	}
	if repository.readyPrompt != "" && repository.getCallCount >= 2 {
		request.Status = PromptRefinementStatusReady
		request.RefinedPrompt = repository.readyPrompt
		repository.requests[requestID] = request
	}
	copy := request
	return &copy, nil
}

func (repository *fakePromptRefinementRepository) MarkPromptRefinementReady(ctx context.Context, requestID string, refinedPrompt string) (*PromptRefinementRequest, error) {
	_ = ctx
	repository.markReadyCall++
	request, exists := repository.requests[requestID]
	if !exists {
		return nil, errors.New("request not found")
	}
	request.Status = PromptRefinementStatusReady
	request.RefinedPrompt = refinedPrompt
	repository.requests[requestID] = request
	copy := request
	return &copy, nil
}

func (repository *fakePromptRefinementRepository) MarkPromptRefinementFailed(ctx context.Context, requestID string, errorMessage string) error {
	_ = ctx
	request, exists := repository.requests[requestID]
	if !exists {
		return errors.New("request not found")
	}
	request.Status = PromptRefinementStatusFailed
	request.ErrorMessage = errorMessage
	repository.requests[requestID] = request
	return nil
}

type fakePromptRefiner struct {
	response string
	err      error
}

func (refiner *fakePromptRefiner) RefinePrompt(ctx context.Context, input PromptRefinerInput) (string, error) {
	_ = ctx
	if refiner.err != nil {
		return "", refiner.err
	}
	if refiner.response != "" {
		return refiner.response, nil
	}
	return "refined: " + input.TaskboardName, nil
}

func TestRefineIngestionPromptEnqueuesAndWaitsForReadyResult(t *testing.T) {
	engine := &deleteTestQueueEngine{}
	scheduler, err := taskengine.NewScheduler(engine, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	repository := &fakePromptRefinementRepository{readyPrompt: "Clear prompt"}
	service := &Service{
		scheduler:                  scheduler,
		promptRefinementRepository: repository,
		promptRefinementWait:       2 * time.Second,
	}

	result, err := service.RefineIngestionPrompt(context.Background(), RefineIngestionPromptInput{
		ProjectID:     "project-1",
		TaskboardName: "Board",
		UserPrompt:    "messy idea",
	})
	if err != nil {
		t.Fatalf("RefineIngestionPrompt() error = %v", err)
	}
	if result == nil || result.Prompt != "Clear prompt" {
		t.Fatalf("expected refined prompt, got %+v", result)
	}
	if len(engine.requests) != 1 {
		t.Fatalf("expected one queued prompt refinement job, got %d", len(engine.requests))
	}
	if engine.requests[0].Kind != taskengine.JobKindPromptRefinementAgent {
		t.Fatalf("expected job kind %q, got %q", taskengine.JobKindPromptRefinementAgent, engine.requests[0].Kind)
	}
}

func TestExecutePromptRefinementUsesRefinerAndMarksReady(t *testing.T) {
	repository := &fakePromptRefinementRepository{
		requests: map[string]PromptRefinementRequest{
			"req-1": {
				RequestID:     "req-1",
				ProjectID:     "project-1",
				TaskboardName: "Board",
				UserPrompt:    "messy",
				Status:        PromptRefinementStatusQueued,
			},
		},
	}
	service := &Service{
		promptRefinementRepository: repository,
		promptRefiner:              &fakePromptRefiner{response: "Tight prompt"},
	}

	if err := service.ExecutePromptRefinement(context.Background(), "req-1"); err != nil {
		t.Fatalf("ExecutePromptRefinement() error = %v", err)
	}
	request, _ := repository.GetPromptRefinementRequest(context.Background(), "req-1")
	if request == nil || request.Status != PromptRefinementStatusReady || request.RefinedPrompt != "Tight prompt" {
		t.Fatalf("expected ready prompt refinement, got %+v", request)
	}
}

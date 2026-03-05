package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	PromptRefinementStatusQueued = "queued"
	PromptRefinementStatusReady  = "ready"
	PromptRefinementStatusFailed = "failed"
)

type RefineIngestionPromptInput struct {
	ProjectID     string
	TaskboardName string
	UserPrompt    string
}

func (input RefineIngestionPromptInput) Validate() error {
	if strings.TrimSpace(input.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(input.TaskboardName) == "" {
		return fmt.Errorf("taskboard_name is required")
	}
	return nil
}

type RefineIngestionPromptResult struct {
	Prompt string
}

type PromptRefinementRequest struct {
	RequestID     string
	ProjectID     string
	TaskboardName string
	UserPrompt    string
	RefinedPrompt string
	Status        string
	ErrorMessage  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PromptRefinementRepository interface {
	CreatePromptRefinementRequest(ctx context.Context, request PromptRefinementRequest) (*PromptRefinementRequest, error)
	GetPromptRefinementRequest(ctx context.Context, requestID string) (*PromptRefinementRequest, error)
	MarkPromptRefinementReady(ctx context.Context, requestID string, refinedPrompt string) (*PromptRefinementRequest, error)
	MarkPromptRefinementFailed(ctx context.Context, requestID string, errorMessage string) error
}

type PromptRefinerInput struct {
	TaskboardName string
	UserPrompt    string
}

type PromptRefiner interface {
	RefinePrompt(ctx context.Context, input PromptRefinerInput) (string, error)
}

type PromptRefinementPayload struct {
	RequestID string `json:"request_id"`
}

func (service *Service) SetPromptRefinementRepository(repository PromptRefinementRepository) {
	if service == nil {
		return
	}
	service.promptRefinementRepository = repository
}

func (service *Service) SetPromptRefiner(refiner PromptRefiner) {
	if service == nil {
		return
	}
	service.promptRefiner = refiner
}

func (service *Service) SetPromptRefinementWait(timeout time.Duration) {
	if service == nil {
		return
	}
	service.promptRefinementWait = timeout
}

func (service *Service) RefineIngestionPrompt(ctx context.Context, input RefineIngestionPromptInput) (*RefineIngestionPromptResult, error) {
	if service == nil || service.scheduler == nil {
		return nil, fmt.Errorf("task scheduler is not configured")
	}
	if service.promptRefinementRepository == nil {
		return nil, fmt.Errorf("prompt refinement repository is not configured")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.TaskboardName = strings.TrimSpace(input.TaskboardName)
	input.UserPrompt = strings.TrimSpace(input.UserPrompt)

	if service.projectRepository != nil {
		setup, err := service.projectRepository.GetProjectSetup(ctx, input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("load project setup: %w", err)
		}
		if setup == nil {
			return nil, fmt.Errorf("project setup not found")
		}
	}

	requestID := fmt.Sprintf("prompt-refine-%d", time.Now().UTC().UnixNano())
	requestRecord, err := service.promptRefinementRepository.CreatePromptRefinementRequest(ctx, PromptRefinementRequest{
		RequestID:     requestID,
		ProjectID:     input.ProjectID,
		TaskboardName: input.TaskboardName,
		UserPrompt:    input.UserPrompt,
		Status:        PromptRefinementStatusQueued,
	})
	if err != nil {
		return nil, fmt.Errorf("create prompt refinement request: %w", err)
	}

	payloadBytes, err := json.Marshal(PromptRefinementPayload{RequestID: requestRecord.RequestID})
	if err != nil {
		return nil, fmt.Errorf("marshal prompt refinement payload: %w", err)
	}
	idempotencyKey := promptRefinementIdempotencyKey(input.ProjectID, input.TaskboardName, input.UserPrompt)
	if _, err := service.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindPromptRefinementAgent,
		Payload:        payloadBytes,
		Queue:          "agent",
		IdempotencyKey: idempotencyKey,
		UniqueFor:      2 * time.Minute,
		Timeout:        3 * time.Minute,
		MaxRetry:       2,
		CorrelationIDs: taskengine.CorrelationIDs{
			RunID:     fmt.Sprintf("prompt-refinement-%s", sanitizeProjectPathSegment(input.ProjectID)),
			TaskID:    "prompt-refinement",
			JobID:     requestRecord.RequestID,
			ProjectID: input.ProjectID,
		},
	}); err != nil {
		return nil, fmt.Errorf("enqueue prompt refinement job: %w", err)
	}

	waitFor := service.promptRefinementWait
	if waitFor <= 0 {
		waitFor = 60 * time.Second
	}
	deadline := time.Now().Add(waitFor)
	for {
		current, loadErr := service.promptRefinementRepository.GetPromptRefinementRequest(ctx, requestRecord.RequestID)
		if loadErr != nil {
			return nil, fmt.Errorf("load prompt refinement request: %w", loadErr)
		}
		if current == nil {
			return nil, fmt.Errorf("prompt refinement request not found")
		}
		switch strings.TrimSpace(current.Status) {
		case PromptRefinementStatusReady:
			return &RefineIngestionPromptResult{Prompt: strings.TrimSpace(current.RefinedPrompt)}, nil
		case PromptRefinementStatusFailed:
			if strings.TrimSpace(current.ErrorMessage) == "" {
				return nil, fmt.Errorf("prompt refinement failed")
			}
			return nil, fmt.Errorf("prompt refinement failed: %s", strings.TrimSpace(current.ErrorMessage))
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("prompt refinement timed out")
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		time.Sleep(120 * time.Millisecond)
	}
}

func (service *Service) ExecutePromptRefinement(ctx context.Context, requestID string) error {
	if service == nil || service.promptRefinementRepository == nil {
		return fmt.Errorf("prompt refinement repository is not configured")
	}
	if service.promptRefiner == nil {
		return fmt.Errorf("prompt refiner is not configured")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Errorf("request_id is required")
	}
	request, err := service.promptRefinementRepository.GetPromptRefinementRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("load prompt refinement request: %w", err)
	}
	if request == nil {
		return fmt.Errorf("prompt refinement request not found")
	}
	refinedPrompt, err := service.promptRefiner.RefinePrompt(ctx, PromptRefinerInput{
		TaskboardName: strings.TrimSpace(request.TaskboardName),
		UserPrompt:    strings.TrimSpace(request.UserPrompt),
	})
	if err != nil {
		_ = service.promptRefinementRepository.MarkPromptRefinementFailed(ctx, requestID, err.Error())
		return fmt.Errorf("refine prompt: %w", err)
	}
	if _, err := service.promptRefinementRepository.MarkPromptRefinementReady(ctx, requestID, strings.TrimSpace(refinedPrompt)); err != nil {
		return fmt.Errorf("mark prompt refinement ready: %w", err)
	}
	return nil
}

func promptRefinementIdempotencyKey(projectID string, taskboardName string, userPrompt string) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(strings.TrimSpace(projectID)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(taskboardName)))
	_, _ = hasher.Write([]byte("|" + strings.TrimSpace(userPrompt)))
	return "prompt-refinement:" + hex.EncodeToString(hasher.Sum(nil))
}

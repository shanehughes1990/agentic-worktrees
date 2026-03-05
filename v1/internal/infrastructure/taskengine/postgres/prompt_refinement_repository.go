package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type promptRefinementRequestRecord struct {
	gorm.Model
	RequestID     string `gorm:"column:request_id;size:255;not null;uniqueIndex"`
	ProjectID     string `gorm:"column:project_id;size:255;not null;index"`
	TaskboardName string `gorm:"column:taskboard_name;size:255;not null"`
	UserPrompt    string `gorm:"column:user_prompt;type:text"`
	RefinedPrompt string `gorm:"column:refined_prompt;type:text"`
	Status        string `gorm:"column:status;size:64;not null;index"`
	ErrorMessage  string `gorm:"column:error_message;type:text"`
}

func (promptRefinementRequestRecord) TableName() string {
	return "prompt_refinement_requests"
}

type PromptRefinementRepository struct {
	db *gorm.DB
}

func NewPromptRefinementRepository(db *gorm.DB) (*PromptRefinementRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("prompt refinement repository db is required")
	}
	if err := db.AutoMigrate(&promptRefinementRequestRecord{}); err != nil {
		return nil, fmt.Errorf("prompt refinement repository migrate: %w", err)
	}
	return &PromptRefinementRepository{db: db}, nil
}

func (repository *PromptRefinementRepository) CreatePromptRefinementRequest(ctx context.Context, request applicationcontrolplane.PromptRefinementRequest) (*applicationcontrolplane.PromptRefinementRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("prompt refinement repository is not initialized")
	}
	record := promptRefinementRequestRecord{
		RequestID:     strings.TrimSpace(request.RequestID),
		ProjectID:     strings.TrimSpace(request.ProjectID),
		TaskboardName: strings.TrimSpace(request.TaskboardName),
		UserPrompt:    strings.TrimSpace(request.UserPrompt),
		Status:        strings.TrimSpace(request.Status),
	}
	if err := repository.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create prompt refinement request: %w", err)
	}
	mapped := mapPromptRefinementRequestRecord(record)
	return &mapped, nil
}

func (repository *PromptRefinementRepository) GetPromptRefinementRequest(ctx context.Context, requestID string) (*applicationcontrolplane.PromptRefinementRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("prompt refinement repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	var record promptRefinementRequestRecord
	err := repository.db.WithContext(ctx).Model(&promptRefinementRequestRecord{}).Where("request_id = ?", requestID).Take(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get prompt refinement request: %w", err)
	}
	mapped := mapPromptRefinementRequestRecord(record)
	return &mapped, nil
}

func (repository *PromptRefinementRepository) MarkPromptRefinementReady(ctx context.Context, requestID string, refinedPrompt string) (*applicationcontrolplane.PromptRefinementRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("prompt refinement repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	result := repository.db.WithContext(ctx).Model(&promptRefinementRequestRecord{}).
		Where("request_id = ?", requestID).
		Updates(map[string]any{
			"status":         applicationcontrolplane.PromptRefinementStatusReady,
			"refined_prompt": strings.TrimSpace(refinedPrompt),
			"error_message":  "",
			"updated_at":     time.Now().UTC(),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("mark prompt refinement ready: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("prompt refinement request not found")
	}
	return repository.GetPromptRefinementRequest(ctx, requestID)
}

func (repository *PromptRefinementRepository) MarkPromptRefinementFailed(ctx context.Context, requestID string, errorMessage string) error {
	if repository == nil || repository.db == nil {
		return fmt.Errorf("prompt refinement repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Errorf("request_id is required")
	}
	result := repository.db.WithContext(ctx).Model(&promptRefinementRequestRecord{}).
		Where("request_id = ?", requestID).
		Updates(map[string]any{
			"status":        applicationcontrolplane.PromptRefinementStatusFailed,
			"error_message": strings.TrimSpace(errorMessage),
			"updated_at":    time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("mark prompt refinement failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("prompt refinement request not found")
	}
	return nil
}

func mapPromptRefinementRequestRecord(record promptRefinementRequestRecord) applicationcontrolplane.PromptRefinementRequest {
	return applicationcontrolplane.PromptRefinementRequest{
		RequestID:     strings.TrimSpace(record.RequestID),
		ProjectID:     strings.TrimSpace(record.ProjectID),
		TaskboardName: strings.TrimSpace(record.TaskboardName),
		UserPrompt:    strings.TrimSpace(record.UserPrompt),
		RefinedPrompt: strings.TrimSpace(record.RefinedPrompt),
		Status:        strings.TrimSpace(record.Status),
		ErrorMessage:  strings.TrimSpace(record.ErrorMessage),
		CreatedAt:     record.CreatedAt.UTC(),
		UpdatedAt:     record.UpdatedAt.UTC(),
	}
}

var _ applicationcontrolplane.PromptRefinementRepository = (*PromptRefinementRepository)(nil)

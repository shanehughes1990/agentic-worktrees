package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectDocumentRecord struct {
	gorm.Model
	ProjectID   string `gorm:"column:project_id;size:255;not null;uniqueIndex:idx_project_document_unique,priority:1;index"`
	DocumentID  string `gorm:"column:document_id;size:255;not null;uniqueIndex:idx_project_document_unique,priority:2;index"`
	FileName    string `gorm:"column:file_name;size:255;not null"`
	ContentType string `gorm:"column:content_type;size:255;not null"`
	ObjectPath  string `gorm:"column:object_path;size:1024;not null"`
	CDNURL      string `gorm:"column:cdn_url;size:2048;not null"`
	Status      string `gorm:"column:status;size:64;not null;index"`
}

func (projectDocumentRecord) TableName() string {
	return "project_documents"
}

type projectDocumentUploadRequestRecord struct {
	gorm.Model
	RequestID    string    `gorm:"column:request_id;size:255;not null;uniqueIndex"`
	ProjectID    string    `gorm:"column:project_id;size:255;not null;index"`
	DocumentID   string    `gorm:"column:document_id;size:255;not null;index"`
	FileName     string    `gorm:"column:file_name;size:255;not null"`
	ContentType  string    `gorm:"column:content_type;size:255;not null"`
	ObjectPath   string    `gorm:"column:object_path;size:1024;not null"`
	Status       string    `gorm:"column:status;size:64;not null;index"`
	UploadURL    string    `gorm:"column:upload_url;size:4096"`
	CDNURL       string    `gorm:"column:cdn_url;size:2048"`
	ErrorMessage string    `gorm:"column:error_message;size:2048"`
	ExpiresAt    time.Time `gorm:"column:expires_at"`
}

func (projectDocumentUploadRequestRecord) TableName() string {
	return "project_document_upload_requests"
}

type ProjectDocumentRepository struct {
	db *gorm.DB
}

func NewProjectDocumentRepository(db *gorm.DB) (*ProjectDocumentRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("project document repository db is required")
	}
	if err := db.AutoMigrate(&projectDocumentRecord{}, &projectDocumentUploadRequestRecord{}); err != nil {
		return nil, fmt.Errorf("project document repository migrate: %w", err)
	}
	return &ProjectDocumentRepository{db: db}, nil
}

func (repository *ProjectDocumentRepository) CreateUploadRequest(ctx context.Context, request applicationcontrolplane.ProjectDocumentUploadRequest) (*applicationcontrolplane.ProjectDocumentUploadRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project document repository is not initialized")
	}
	record := projectDocumentUploadRequestRecord{
		RequestID:   strings.TrimSpace(request.RequestID),
		ProjectID:   strings.TrimSpace(request.ProjectID),
		DocumentID:  strings.TrimSpace(request.DocumentID),
		FileName:    strings.TrimSpace(request.FileName),
		ContentType: strings.TrimSpace(request.ContentType),
		ObjectPath:  strings.TrimSpace(request.ObjectPath),
		Status:      strings.TrimSpace(request.Status),
	}
	if err := repository.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create project document upload request: %w", err)
	}
	mapped := mapUploadRequestRecord(record)
	return &mapped, nil
}

func (repository *ProjectDocumentRepository) GetUploadRequest(ctx context.Context, requestID string) (*applicationcontrolplane.ProjectDocumentUploadRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project document repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	var record projectDocumentUploadRequestRecord
	err := repository.db.WithContext(ctx).Model(&projectDocumentUploadRequestRecord{}).Where("request_id = ?", requestID).Take(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get project document upload request: %w", err)
	}
	mapped := mapUploadRequestRecord(record)
	return &mapped, nil
}

func (repository *ProjectDocumentRepository) MarkUploadRequestReady(ctx context.Context, requestID string, uploadURL string, cdnURL string, expiresAt time.Time) (*applicationcontrolplane.ProjectDocumentUploadRequest, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project document repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	uploadURL = strings.TrimSpace(uploadURL)
	cdnURL = strings.TrimSpace(cdnURL)
	if uploadURL == "" {
		return nil, fmt.Errorf("upload_url is required")
	}
	if cdnURL == "" {
		return nil, fmt.Errorf("cdn_url is required")
	}
	var requestRecord projectDocumentUploadRequestRecord
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&projectDocumentUploadRequestRecord{}).
			Where("request_id = ?", requestID).
			Take(&requestRecord).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("project document upload request not found")
			}
			return fmt.Errorf("load project document upload request: %w", err)
		}
		if err := tx.Model(&projectDocumentUploadRequestRecord{}).
			Where("id = ?", requestRecord.ID).
			Updates(map[string]any{
				"status":     applicationcontrolplane.ProjectDocumentUploadStatusReady,
				"upload_url": uploadURL,
				"cdn_url":    cdnURL,
				"expires_at": expiresAt.UTC(),
				"updated_at": time.Now().UTC(),
			}).Error; err != nil {
			return fmt.Errorf("update upload request ready state: %w", err)
		}
		documentRecord := projectDocumentRecord{
			ProjectID:   strings.TrimSpace(requestRecord.ProjectID),
			DocumentID:  strings.TrimSpace(requestRecord.DocumentID),
			FileName:    strings.TrimSpace(requestRecord.FileName),
			ContentType: strings.TrimSpace(requestRecord.ContentType),
			ObjectPath:  strings.TrimSpace(requestRecord.ObjectPath),
			CDNURL:      cdnURL,
			Status:      applicationcontrolplane.ProjectDocumentStatusPendingUpload,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "project_id"}, {Name: "document_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"file_name", "content_type", "object_path", "cdn_url", "status", "updated_at"}),
		}).Create(&documentRecord).Error; err != nil {
			return fmt.Errorf("upsert project document record: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return repository.GetUploadRequest(ctx, requestID)
}

func (repository *ProjectDocumentRepository) MarkUploadRequestFailed(ctx context.Context, requestID string, errorMessage string) error {
	if repository == nil || repository.db == nil {
		return fmt.Errorf("project document repository is not initialized")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Errorf("request_id is required")
	}
	result := repository.db.WithContext(ctx).Model(&projectDocumentUploadRequestRecord{}).
		Where("request_id = ?", requestID).
		Updates(map[string]any{
			"status":        applicationcontrolplane.ProjectDocumentUploadStatusFailed,
			"error_message": strings.TrimSpace(errorMessage),
			"updated_at":    time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("mark project document upload request failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project document upload request not found")
	}
	return nil
}

func (repository *ProjectDocumentRepository) ListProjectDocuments(ctx context.Context, projectID string, limit int) ([]applicationcontrolplane.ProjectDocument, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project document repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	records := make([]projectDocumentRecord, 0)
	if err := repository.db.WithContext(ctx).
		Model(&projectDocumentRecord{}).
		Where("project_id = ?", projectID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list project documents: %w", err)
	}
	result := make([]applicationcontrolplane.ProjectDocument, 0, len(records))
	for _, record := range records {
		result = append(result, mapProjectDocumentRecord(record))
	}
	return result, nil
}

func (repository *ProjectDocumentRepository) GetProjectDocument(ctx context.Context, projectID string, documentID string) (*applicationcontrolplane.ProjectDocument, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("project document repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if documentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}
	var record projectDocumentRecord
	err := repository.db.WithContext(ctx).
		Model(&projectDocumentRecord{}).
		Where("project_id = ? AND document_id = ?", projectID, documentID).
		Take(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get project document: %w", err)
	}
	mapped := mapProjectDocumentRecord(record)
	return &mapped, nil
}

func (repository *ProjectDocumentRepository) DeleteProjectDocument(ctx context.Context, projectID string, documentID string) error {
	if repository == nil || repository.db == nil {
		return fmt.Errorf("project document repository is not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if documentID == "" {
		return fmt.Errorf("document_id is required")
	}
	return repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		deleteResult := tx.Where("project_id = ? AND document_id = ?", projectID, documentID).Delete(&projectDocumentRecord{})
		if deleteResult.Error != nil {
			return fmt.Errorf("delete project document: %w", deleteResult.Error)
		}
		if deleteResult.RowsAffected == 0 {
			return fmt.Errorf("project document not found")
		}
		if err := tx.Where("project_id = ? AND document_id = ?", projectID, documentID).Delete(&projectDocumentUploadRequestRecord{}).Error; err != nil {
			return fmt.Errorf("delete project document upload requests: %w", err)
		}
		return nil
	})
}

func mapUploadRequestRecord(record projectDocumentUploadRequestRecord) applicationcontrolplane.ProjectDocumentUploadRequest {
	return applicationcontrolplane.ProjectDocumentUploadRequest{
		RequestID:    strings.TrimSpace(record.RequestID),
		ProjectID:    strings.TrimSpace(record.ProjectID),
		DocumentID:   strings.TrimSpace(record.DocumentID),
		FileName:     strings.TrimSpace(record.FileName),
		ContentType:  strings.TrimSpace(record.ContentType),
		ObjectPath:   strings.TrimSpace(record.ObjectPath),
		Status:       strings.TrimSpace(record.Status),
		UploadURL:    strings.TrimSpace(record.UploadURL),
		CDNURL:       strings.TrimSpace(record.CDNURL),
		ErrorMessage: strings.TrimSpace(record.ErrorMessage),
		ExpiresAt:    record.ExpiresAt.UTC(),
		CreatedAt:    record.CreatedAt.UTC(),
		UpdatedAt:    record.UpdatedAt.UTC(),
	}
}

func mapProjectDocumentRecord(record projectDocumentRecord) applicationcontrolplane.ProjectDocument {
	return applicationcontrolplane.ProjectDocument{
		ProjectID:   strings.TrimSpace(record.ProjectID),
		DocumentID:  strings.TrimSpace(record.DocumentID),
		FileName:    strings.TrimSpace(record.FileName),
		ContentType: strings.TrimSpace(record.ContentType),
		ObjectPath:  strings.TrimSpace(record.ObjectPath),
		CDNURL:      strings.TrimSpace(record.CDNURL),
		Status:      strings.TrimSpace(record.Status),
		CreatedAt:   record.CreatedAt.UTC(),
		UpdatedAt:   record.UpdatedAt.UTC(),
	}
}

var _ applicationcontrolplane.ProjectDocumentRepository = (*ProjectDocumentRepository)(nil)

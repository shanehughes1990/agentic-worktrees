package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	ProjectDocumentStatusPendingUpload = "pending_upload"
	ProjectDocumentStatusAvailable     = "available"

	ProjectDocumentUploadStatusQueued = "queued"
	ProjectDocumentUploadStatusReady  = "ready"
	ProjectDocumentUploadStatusFailed = "failed"
)

type ProjectDocument struct {
	ProjectID   string
	DocumentID  string
	FileName    string
	ContentType string
	ObjectPath  string
	CDNURL      string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectDocumentUploadRequest struct {
	RequestID    string
	ProjectID    string
	DocumentID   string
	FileName     string
	ContentType  string
	ObjectPath   string
	Status       string
	UploadURL    string
	CDNURL       string
	ErrorMessage string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RequestProjectDocumentUploadInput struct {
	ProjectID   string
	FileName    string
	ContentType string
}

func (input RequestProjectDocumentUploadInput) Validate() error {
	if strings.TrimSpace(input.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(input.FileName) == "" {
		return fmt.Errorf("file_name is required")
	}
	if strings.Contains(strings.TrimSpace(input.FileName), "/") {
		return fmt.Errorf("file_name must not contain path separators")
	}
	if strings.TrimSpace(input.ContentType) == "" {
		return fmt.Errorf("content_type is required")
	}
	return nil
}

type RequestProjectDocumentUploadResult struct {
	RequestID   string
	ProjectID   string
	DocumentID  string
	FileName    string
	ContentType string
	ObjectPath  string
	UploadURL   string
	CDNURL      string
	ExpiresAt   time.Time
	Status      string
}

type ProjectDocumentPreviewResult struct {
	ProjectID   string
	DocumentID  string
	FileName    string
	ContentType string
	ObjectPath  string
	CDNURL      string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectDocumentRepository interface {
	CreateUploadRequest(ctx context.Context, request ProjectDocumentUploadRequest) (*ProjectDocumentUploadRequest, error)
	GetUploadRequest(ctx context.Context, requestID string) (*ProjectDocumentUploadRequest, error)
	MarkUploadRequestReady(ctx context.Context, requestID string, uploadURL string, cdnURL string, expiresAt time.Time) (*ProjectDocumentUploadRequest, error)
	MarkUploadRequestFailed(ctx context.Context, requestID string, errorMessage string) error
	ListProjectDocuments(ctx context.Context, projectID string, limit int) ([]ProjectDocument, error)
	GetProjectDocument(ctx context.Context, projectID string, documentID string) (*ProjectDocument, error)
	DeleteProjectDocument(ctx context.Context, projectID string, documentID string) error
}

type ProjectFileStore interface {
	CreateSignedUploadURL(ctx context.Context, objectPath string, contentType string, expiresAt time.Time) (string, error)
	DeleteObject(ctx context.Context, objectPath string) error
}

type ProjectCDNSigner interface {
	SignedObjectURL(objectPath string, expiresAt time.Time) (string, error)
}

type PrepareProjectDocumentUploadPayload struct {
	RequestID string `json:"request_id"`
}

type DeleteProjectDocumentPayload struct {
	ProjectID  string `json:"project_id"`
	DocumentID string `json:"document_id"`
}

var (
	documentNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	projectIDSanitizer    = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
)

func (service *Service) SetProjectDocumentRepository(repository ProjectDocumentRepository) {
	if service == nil {
		return
	}
	service.projectDocumentRepository = repository
}

func (service *Service) SetProjectFileStore(store ProjectFileStore) {
	if service == nil {
		return
	}
	service.projectFileStore = store
}

func (service *Service) SetProjectCDNSigner(signer ProjectCDNSigner) {
	if service == nil {
		return
	}
	service.projectCDNSigner = signer
}

func (service *Service) SetProjectDocumentRootPrefix(prefix string) {
	if service == nil {
		return
	}
	service.projectDocumentRootPrefix = strings.TrimSpace(prefix)
}

func (service *Service) SetProjectDocumentRemoteStorageType(storageType string) {
	if service == nil {
		return
	}
	service.projectDocumentRemoteStorageType = strings.ToLower(strings.TrimSpace(storageType))
}

func (service *Service) SetProjectDocumentGoogleApplicationCredentialsPath(applicationCredentialsFilePath string) {
	if service == nil {
		return
	}
	service.projectDocumentGoogleApplicationCredentialsPath = strings.TrimSpace(applicationCredentialsFilePath)
}

func (service *Service) SetProjectDocumentUploadWait(timeout time.Duration) {
	if service == nil {
		return
	}
	service.projectDocumentUploadWait = timeout
}

func (service *Service) RequestProjectDocumentUpload(ctx context.Context, input RequestProjectDocumentUploadInput) (*RequestProjectDocumentUploadResult, error) {
	if service == nil || service.projectDocumentRepository == nil {
		return nil, fmt.Errorf("project document repository is not configured")
	}
	if service.scheduler == nil {
		return nil, fmt.Errorf("task scheduler is not configured")
	}
	switch strings.ToLower(strings.TrimSpace(service.projectDocumentRemoteStorageType)) {
	case "", "gcs":
		// currently supported and expected for desktop upload requests
	default:
		return nil, fmt.Errorf("unsupported remote storage type %q for project document uploads", service.projectDocumentRemoteStorageType)
	}
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.FileName = strings.TrimSpace(path.Base(input.FileName))
	input.ContentType = strings.TrimSpace(input.ContentType)
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if service.projectRepository != nil {
		setup, err := service.projectRepository.GetProjectSetup(ctx, input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("load project setup: %w", err)
		}
		if setup == nil {
			return nil, fmt.Errorf("project setup not found")
		}
	}

	documentID := fmt.Sprintf("doc-%d", time.Now().UTC().UnixNano())
	requestID := fmt.Sprintf("upload-%d", time.Now().UTC().UnixNano())
	objectPath := projectDocumentObjectPath(service.projectDocumentRootPrefix, input.ProjectID, documentID, input.FileName)

	uploadRequest, err := service.projectDocumentRepository.CreateUploadRequest(ctx, ProjectDocumentUploadRequest{
		RequestID:   requestID,
		ProjectID:   input.ProjectID,
		DocumentID:  documentID,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		ObjectPath:  objectPath,
		Status:      ProjectDocumentUploadStatusQueued,
	})
	if err != nil {
		return nil, fmt.Errorf("create project document upload request: %w", err)
	}

	payloadBytes, err := json.Marshal(PrepareProjectDocumentUploadPayload{RequestID: uploadRequest.RequestID})
	if err != nil {
		return nil, fmt.Errorf("marshal project document upload task payload: %w", err)
	}
	_, err = service.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindProjectDocumentUploadPrepare,
		Payload:        payloadBytes,
		Queue:          "files",
		IdempotencyKey: uploadRequest.RequestID,
		UniqueFor:      2 * time.Minute,
		Timeout:        2 * time.Minute,
		MaxRetry:       2,
		CorrelationIDs: taskengine.CorrelationIDs{ProjectID: input.ProjectID},
	})
	if err != nil {
		return nil, fmt.Errorf("enqueue project document upload task: %w", err)
	}

	waitFor := service.projectDocumentUploadWait
	if waitFor <= 0 {
		waitFor = 5 * time.Second
	}
	deadline := time.Now().Add(waitFor)
	for {
		current, loadErr := service.projectDocumentRepository.GetUploadRequest(ctx, uploadRequest.RequestID)
		if loadErr != nil {
			return nil, fmt.Errorf("load project document upload request: %w", loadErr)
		}
		if current == nil {
			return nil, fmt.Errorf("project document upload request not found")
		}
		switch strings.TrimSpace(current.Status) {
		case ProjectDocumentUploadStatusReady:
			return &RequestProjectDocumentUploadResult{
				RequestID:   current.RequestID,
				ProjectID:   current.ProjectID,
				DocumentID:  current.DocumentID,
				FileName:    current.FileName,
				ContentType: current.ContentType,
				ObjectPath:  current.ObjectPath,
				UploadURL:   current.UploadURL,
				CDNURL:      current.CDNURL,
				ExpiresAt:   current.ExpiresAt,
				Status:      current.Status,
			}, nil
		case ProjectDocumentUploadStatusFailed:
			if strings.TrimSpace(current.ErrorMessage) != "" {
				return nil, fmt.Errorf("prepare project document upload: %s", strings.TrimSpace(current.ErrorMessage))
			}
			return nil, fmt.Errorf("prepare project document upload failed")
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("prepare project document upload timed out")
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (service *Service) ProjectDocuments(ctx context.Context, projectID string, limit int) ([]ProjectDocument, error) {
	if service == nil || service.projectDocumentRepository == nil {
		return nil, fmt.Errorf("project document repository is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	return service.projectDocumentRepository.ListProjectDocuments(ctx, projectID, limit)
}

func (service *Service) ProjectDocumentPreview(ctx context.Context, projectID string, documentID string) (*ProjectDocumentPreviewResult, error) {
	if service == nil || service.projectDocumentRepository == nil {
		return nil, fmt.Errorf("project document repository is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if documentID == "" {
		return nil, fmt.Errorf("document_id is required")
	}
	document, err := service.projectDocumentRepository.GetProjectDocument(ctx, projectID, documentID)
	if err != nil {
		return nil, fmt.Errorf("load project document: %w", err)
	}
	if document == nil {
		return nil, nil
	}
	return &ProjectDocumentPreviewResult{
		ProjectID:   document.ProjectID,
		DocumentID:  document.DocumentID,
		FileName:    document.FileName,
		ContentType: document.ContentType,
		ObjectPath:  document.ObjectPath,
		CDNURL:      document.CDNURL,
		Status:      document.Status,
		CreatedAt:   document.CreatedAt,
		UpdatedAt:   document.UpdatedAt,
	}, nil
}

func (service *Service) DeleteProjectDocument(ctx context.Context, projectID string, documentID string) error {
	if service == nil || service.scheduler == nil {
		return fmt.Errorf("task scheduler is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if documentID == "" {
		return fmt.Errorf("document_id is required")
	}
	payloadBytes, err := json.Marshal(DeleteProjectDocumentPayload{ProjectID: projectID, DocumentID: documentID})
	if err != nil {
		return fmt.Errorf("marshal project document delete task payload: %w", err)
	}
	_, err = service.scheduler.Enqueue(ctx, taskengine.EnqueueRequest{
		Kind:           taskengine.JobKindProjectDocumentDelete,
		Payload:        payloadBytes,
		Queue:          "files",
		IdempotencyKey: fmt.Sprintf("delete-%s-%s", projectID, documentID),
		UniqueFor:      2 * time.Minute,
		Timeout:        2 * time.Minute,
		MaxRetry:       2,
		CorrelationIDs: taskengine.CorrelationIDs{ProjectID: projectID},
	})
	if err != nil {
		return fmt.Errorf("enqueue project document delete task: %w", err)
	}
	return nil
}

func (service *Service) PrepareProjectDocumentUpload(ctx context.Context, requestID string) error {
	if service == nil || service.projectDocumentRepository == nil {
		return fmt.Errorf("project document repository is not configured")
	}
	if service.projectFileStore == nil {
		return fmt.Errorf("project file store is not configured")
	}
	if service.projectCDNSigner == nil {
		return fmt.Errorf("project cdn signer is not configured")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Errorf("request_id is required")
	}
	uploadRequest, err := service.projectDocumentRepository.GetUploadRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("load project document upload request: %w", err)
	}
	if uploadRequest == nil {
		return fmt.Errorf("project document upload request not found")
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	uploadURL, err := service.projectFileStore.CreateSignedUploadURL(ctx, uploadRequest.ObjectPath, uploadRequest.ContentType, expiresAt)
	if err != nil {
		_ = service.projectDocumentRepository.MarkUploadRequestFailed(ctx, requestID, err.Error())
		return fmt.Errorf("create signed upload url: %w", err)
	}
	publicURL, err := service.projectCDNSigner.SignedObjectURL(uploadRequest.ObjectPath, expiresAt)
	if err != nil {
		_ = service.projectDocumentRepository.MarkUploadRequestFailed(ctx, requestID, err.Error())
		return fmt.Errorf("sign project document cdn url: %w", err)
	}
	if _, err := service.projectDocumentRepository.MarkUploadRequestReady(ctx, requestID, uploadURL, publicURL, expiresAt); err != nil {
		return fmt.Errorf("mark upload request ready: %w", err)
	}
	return nil
}

func (service *Service) ExecuteProjectDocumentDelete(ctx context.Context, projectID string, documentID string) error {
	if service == nil || service.projectDocumentRepository == nil {
		return fmt.Errorf("project document repository is not configured")
	}
	if service.projectFileStore == nil {
		return fmt.Errorf("project file store is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if documentID == "" {
		return fmt.Errorf("document_id is required")
	}
	document, err := service.projectDocumentRepository.GetProjectDocument(ctx, projectID, documentID)
	if err != nil {
		return fmt.Errorf("load project document: %w", err)
	}
	if document == nil {
		return nil
	}
	if err := service.projectFileStore.DeleteObject(ctx, strings.TrimSpace(document.ObjectPath)); err != nil {
		return fmt.Errorf("delete project document object: %w", err)
	}
	if err := service.projectDocumentRepository.DeleteProjectDocument(ctx, projectID, documentID); err != nil {
		return fmt.Errorf("delete project document record: %w", err)
	}
	return nil
}

func projectDocumentObjectPath(rootPrefix string, projectID string, documentID string, fileName string) string {
	cleanRoot := strings.Trim(path.Clean(strings.TrimSpace(rootPrefix)), "/")
	if cleanRoot == "" || cleanRoot == "." {
		cleanRoot = "projects"
	}
	cleanProjectID := sanitizeProjectPathSegment(projectID)
	cleanDocumentID := sanitizeProjectPathSegment(documentID)
	cleanFileName := sanitizeDocumentFileName(fileName)
	return path.Join(cleanRoot, cleanProjectID, "documents", cleanDocumentID, cleanFileName)
}

func sanitizeProjectPathSegment(value string) string {
	cleaned := strings.TrimSpace(value)
	if cleaned == "" {
		return "unscoped"
	}
	cleaned = projectIDSanitizer.ReplaceAllString(cleaned, "_")
	cleaned = strings.Trim(cleaned, "_-")
	if cleaned == "" {
		return "unscoped"
	}
	return strings.ToLower(cleaned)
}

func sanitizeDocumentFileName(fileName string) string {
	baseName := strings.TrimSpace(path.Base(fileName))
	if baseName == "" || baseName == "." || baseName == ".." {
		return "file"
	}
	cleaned := documentNameSanitizer.ReplaceAllString(baseName, "_")
	cleaned = strings.Trim(cleaned, "_")
	if cleaned == "" {
		return "file"
	}
	return cleaned
}

package controlplane

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type deleteTestProjectDocumentRepository struct {
	documents         map[string]ProjectDocument
	deleteCalls       int
	deletedProjectID  string
	deletedDocumentID string
}

func (repository *deleteTestProjectDocumentRepository) key(projectID string, documentID string) string {
	return fmt.Sprintf("%s/%s", projectID, documentID)
}

func (repository *deleteTestProjectDocumentRepository) CreateUploadRequest(ctx context.Context, request ProjectDocumentUploadRequest) (*ProjectDocumentUploadRequest, error) {
	_ = ctx
	_ = request
	return nil, errors.New("not implemented in test")
}

func (repository *deleteTestProjectDocumentRepository) GetUploadRequest(ctx context.Context, requestID string) (*ProjectDocumentUploadRequest, error) {
	_ = ctx
	_ = requestID
	return nil, errors.New("not implemented in test")
}

func (repository *deleteTestProjectDocumentRepository) MarkUploadRequestReady(ctx context.Context, requestID string, uploadURL string, cdnURL string, expiresAt time.Time) (*ProjectDocumentUploadRequest, error) {
	_ = ctx
	_ = requestID
	_ = uploadURL
	_ = cdnURL
	_ = expiresAt
	return nil, errors.New("not implemented in test")
}

func (repository *deleteTestProjectDocumentRepository) MarkUploadRequestFailed(ctx context.Context, requestID string, errorMessage string) error {
	_ = ctx
	_ = requestID
	_ = errorMessage
	return errors.New("not implemented in test")
}

func (repository *deleteTestProjectDocumentRepository) ListProjectDocuments(ctx context.Context, projectID string, limit int) ([]ProjectDocument, error) {
	_ = ctx
	_ = projectID
	_ = limit
	return nil, errors.New("not implemented in test")
}

func (repository *deleteTestProjectDocumentRepository) GetProjectDocument(ctx context.Context, projectID string, documentID string) (*ProjectDocument, error) {
	_ = ctx
	if repository.documents == nil {
		return nil, nil
	}
	document, exists := repository.documents[repository.key(projectID, documentID)]
	if !exists {
		return nil, nil
	}
	copy := document
	return &copy, nil
}

func (repository *deleteTestProjectDocumentRepository) DeleteProjectDocument(ctx context.Context, projectID string, documentID string) error {
	_ = ctx
	repository.deleteCalls++
	repository.deletedProjectID = projectID
	repository.deletedDocumentID = documentID
	if repository.documents != nil {
		delete(repository.documents, repository.key(projectID, documentID))
	}
	return nil
}

type deleteTestProjectFileStore struct {
	deletedObjectPaths []string
	deleteErr          error
}

func (store *deleteTestProjectFileStore) CreateSignedUploadURL(ctx context.Context, objectPath string, contentType string, expiresAt time.Time) (string, error) {
	_ = ctx
	_ = objectPath
	_ = contentType
	_ = expiresAt
	return "", errors.New("not implemented in test")
}

func (store *deleteTestProjectFileStore) DeleteObject(ctx context.Context, objectPath string) error {
	_ = ctx
	store.deletedObjectPaths = append(store.deletedObjectPaths, objectPath)
	return store.deleteErr
}

type deleteTestQueueEngine struct {
	requests []taskengine.EnqueueRequest
}

func (engine *deleteTestQueueEngine) Enqueue(ctx context.Context, request taskengine.EnqueueRequest) (taskengine.EnqueueResult, error) {
	_ = ctx
	engine.requests = append(engine.requests, request)
	return taskengine.EnqueueResult{QueueTaskID: request.IdempotencyKey}, nil
}

func TestExecuteProjectDocumentDeleteRemovesRemoteObjectAndRecord(t *testing.T) {
	repository := &deleteTestProjectDocumentRepository{
		documents: map[string]ProjectDocument{
			"project-1/doc-1": {
				ProjectID:  "project-1",
				DocumentID: "doc-1",
				ObjectPath: "projects/project-1/documents/doc-1/file.md",
			},
		},
	}
	fileStore := &deleteTestProjectFileStore{}
	service := &Service{}
	service.SetProjectDocumentRepository(repository)
	service.SetProjectFileStore(fileStore)

	if err := service.ExecuteProjectDocumentDelete(context.Background(), "project-1", "doc-1"); err != nil {
		t.Fatalf("ExecuteProjectDocumentDelete() error = %v", err)
	}
	if len(fileStore.deletedObjectPaths) != 1 || fileStore.deletedObjectPaths[0] != "projects/project-1/documents/doc-1/file.md" {
		t.Fatalf("expected remote object delete call, got %+v", fileStore.deletedObjectPaths)
	}
	if repository.deleteCalls != 1 || repository.deletedProjectID != "project-1" || repository.deletedDocumentID != "doc-1" {
		t.Fatalf("expected repository delete for project-1/doc-1, got calls=%d project=%q document=%q", repository.deleteCalls, repository.deletedProjectID, repository.deletedDocumentID)
	}
}

func TestExecuteProjectDocumentDeleteStopsWhenRemoteObjectDeleteFails(t *testing.T) {
	repository := &deleteTestProjectDocumentRepository{
		documents: map[string]ProjectDocument{
			"project-1/doc-1": {
				ProjectID:  "project-1",
				DocumentID: "doc-1",
				ObjectPath: "projects/project-1/documents/doc-1/file.md",
			},
		},
	}
	fileStore := &deleteTestProjectFileStore{deleteErr: errors.New("bucket unavailable")}
	service := &Service{}
	service.SetProjectDocumentRepository(repository)
	service.SetProjectFileStore(fileStore)

	err := service.ExecuteProjectDocumentDelete(context.Background(), "project-1", "doc-1")
	if err == nil {
		t.Fatalf("expected delete error when remote object delete fails")
	}
	if repository.deleteCalls != 0 {
		t.Fatalf("expected repository record delete to be skipped when object delete fails, got calls=%d", repository.deleteCalls)
	}
}

func TestDeleteProjectDocumentEnqueuesDeleteJob(t *testing.T) {
	engine := &deleteTestQueueEngine{}
	scheduler, err := taskengine.NewScheduler(engine, taskengine.DefaultPolicies())
	if err != nil {
		t.Fatalf("new scheduler: %v", err)
	}
	service := &Service{scheduler: scheduler}

	if err := service.DeleteProjectDocument(context.Background(), "project-1", "doc-1"); err != nil {
		t.Fatalf("DeleteProjectDocument() error = %v", err)
	}
	if len(engine.requests) != 1 {
		t.Fatalf("expected one enqueue request, got %d", len(engine.requests))
	}
	request := engine.requests[0]
	if request.Kind != taskengine.JobKindProjectDocumentDelete {
		t.Fatalf("expected kind %q, got %q", taskengine.JobKindProjectDocumentDelete, request.Kind)
	}
	if request.Queue != "files" {
		t.Fatalf("expected files queue, got %q", request.Queue)
	}
}

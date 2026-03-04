package postgres

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"testing"
	"time"
)

func TestProjectDocumentRepositoryUploadLifecycleAndDeleteAssociations(t *testing.T) {
	repository, err := NewProjectDocumentRepository(newTestDB(t))
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	created, err := repository.CreateUploadRequest(context.Background(), applicationcontrolplane.ProjectDocumentUploadRequest{
		RequestID:   "req-1",
		ProjectID:   "project-1",
		DocumentID:  "doc-1",
		FileName:    "spec.pdf",
		ContentType: "application/pdf",
		ObjectPath:  "projects/project-1/documents/doc-1/spec.pdf",
		Status:      applicationcontrolplane.ProjectDocumentUploadStatusQueued,
	})
	if err != nil {
		t.Fatalf("create upload request: %v", err)
	}
	if created.Status != applicationcontrolplane.ProjectDocumentUploadStatusQueued {
		t.Fatalf("expected queued upload status, got %q", created.Status)
	}

	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	ready, err := repository.MarkUploadRequestReady(
		context.Background(),
		"req-1",
		"https://uploads.example.com/req-1",
		"https://cdn.example.com/doc-1",
		expiresAt,
	)
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	if ready.Status != applicationcontrolplane.ProjectDocumentUploadStatusReady {
		t.Fatalf("expected ready upload status, got %q", ready.Status)
	}

	document, err := repository.GetProjectDocument(context.Background(), "project-1", "doc-1")
	if err != nil {
		t.Fatalf("get document: %v", err)
	}
	if document == nil {
		t.Fatal("expected project document to be created on ready")
	}
	if document.Status != applicationcontrolplane.ProjectDocumentStatusPendingUpload {
		t.Fatalf("expected pending_upload document status, got %q", document.Status)
	}

	if err := repository.MarkUploadRequestFailed(context.Background(), "req-1", "checksum mismatch"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	failed, err := repository.GetUploadRequest(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("load failed upload request: %v", err)
	}
	if failed == nil {
		t.Fatal("expected upload request after failure update")
	}
	if failed.Status != applicationcontrolplane.ProjectDocumentUploadStatusFailed {
		t.Fatalf("expected failed upload status, got %q", failed.Status)
	}
	if failed.ErrorMessage != "checksum mismatch" {
		t.Fatalf("expected persisted error message, got %q", failed.ErrorMessage)
	}

	if err := repository.DeleteProjectDocument(context.Background(), "project-1", "doc-1"); err != nil {
		t.Fatalf("delete document: %v", err)
	}

	deletedDocument, err := repository.GetProjectDocument(context.Background(), "project-1", "doc-1")
	if err != nil {
		t.Fatalf("load deleted document: %v", err)
	}
	if deletedDocument != nil {
		t.Fatalf("expected deleted document to be absent, got %+v", deletedDocument)
	}

	deletedRequest, err := repository.GetUploadRequest(context.Background(), "req-1")
	if err != nil {
		t.Fatalf("load deleted upload request: %v", err)
	}
	if deletedRequest != nil {
		t.Fatalf("expected upload requests to be deleted with document, got %+v", deletedRequest)
	}
}

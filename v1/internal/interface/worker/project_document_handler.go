package worker

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ProjectDocumentPrepareUploadHandler struct {
	service *applicationcontrolplane.Service
}

func NewProjectDocumentPrepareUploadHandler(service *applicationcontrolplane.Service) (*ProjectDocumentPrepareUploadHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("control-plane service is required")
	}
	return &ProjectDocumentPrepareUploadHandler{service: service}, nil
}

func (handler *ProjectDocumentPrepareUploadHandler) Handle(ctx context.Context, job taskengine.Job) error {
	if handler == nil || handler.service == nil {
		return fmt.Errorf("project document prepare-upload handler is not initialized")
	}
	var payload applicationcontrolplane.PrepareProjectDocumentUploadPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode project document prepare-upload payload: %w", err)
	}
	if strings.TrimSpace(payload.RequestID) == "" {
		return fmt.Errorf("request_id is required")
	}
	if err := handler.service.PrepareProjectDocumentUpload(ctx, payload.RequestID); err != nil {
		return err
	}
	return nil
}

type ProjectDocumentDeleteHandler struct {
	service *applicationcontrolplane.Service
}

func NewProjectDocumentDeleteHandler(service *applicationcontrolplane.Service) (*ProjectDocumentDeleteHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("control-plane service is required")
	}
	return &ProjectDocumentDeleteHandler{service: service}, nil
}

func (handler *ProjectDocumentDeleteHandler) Handle(ctx context.Context, job taskengine.Job) error {
	if handler == nil || handler.service == nil {
		return fmt.Errorf("project document delete handler is not initialized")
	}
	var payload applicationcontrolplane.DeleteProjectDocumentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode project document delete payload: %w", err)
	}
	if strings.TrimSpace(payload.ProjectID) == "" {
		return fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(payload.DocumentID) == "" {
		return fmt.Errorf("document_id is required")
	}
	if err := handler.service.ExecuteProjectDocumentDelete(ctx, payload.ProjectID, payload.DocumentID); err != nil {
		return err
	}
	return nil
}

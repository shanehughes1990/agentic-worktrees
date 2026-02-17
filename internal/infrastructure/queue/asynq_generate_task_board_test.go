package queue

import (
	"context"
	"testing"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
)

func TestNewAsynqClientValidation(t *testing.T) {
	if _, err := NewAsynqClient(""); err == nil {
		t.Fatalf("expected redis address validation error")
	}
}

func TestNewAsynqGenerateTaskBoardClientValidation(t *testing.T) {
	if _, err := NewAsynqGenerateTaskBoardClient(nil, "default"); err == nil {
		t.Fatalf("expected asynq client validation error")
	}
	baseClient := &AsynqClient{}
	if _, err := NewAsynqGenerateTaskBoardClient(baseClient, ""); err == nil {
		t.Fatalf("expected queue validation error")
	}
}

func TestEnqueueGenerateTaskBoardValidation(t *testing.T) {
	client := &AsynqGenerateTaskBoardClient{}
	if _, err := client.EnqueueGenerateTaskBoard(context.Background(), application.GenerateTaskBoardPayload{}); err == nil {
		t.Fatalf("expected payload validation error")
	}
	if _, err := client.EnqueueGenerateTaskBoard(context.Background(), application.GenerateTaskBoardPayload{
		Metadata:      domainservices.AgentRequestMetadata{JobID: "job-1"},
		Prompt:        "build",
		RootDirectory: "docs",
		MaxDepth:      1,
		Documents:     []domainservices.DocumentationSourceFile{{Path: "docs/a.md", Content: "x"}},
	}); err == nil {
		t.Fatalf("expected enqueue error due to nil client")
	}
}

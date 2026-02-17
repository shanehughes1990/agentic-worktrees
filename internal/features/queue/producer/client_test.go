package producer

import (
	"testing"

	queuedomain "github.com/shanehughes1990/agentic-worktrees/internal/features/queue/domain"
)

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient("", "default"); err == nil {
		t.Fatalf("expected error for empty redis address")
	}
}

func TestNewClientTrimsQueueName(t *testing.T) {
	client, err := NewClient("127.0.0.1:6379", "  default  ")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if client.queueName != "default" {
		t.Fatalf("expected trimmed queue name, got %q", client.queueName)
	}
}

func TestEnqueuePlanBoardValidatesPayload(t *testing.T) {
	client, err := NewClient("127.0.0.1:6379", "default")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	_, err = client.EnqueuePlanBoard(queuedomain.PlanBoardPayload{})
	if err == nil {
		t.Fatalf("expected payload validation error")
	}
}

func TestCloseNilClient(t *testing.T) {
	var client *Client
	if err := client.Close(); err != nil {
		t.Fatalf("expected nil close to succeed, got %v", err)
	}
}

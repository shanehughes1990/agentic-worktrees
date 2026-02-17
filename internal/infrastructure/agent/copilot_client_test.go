package agent

import (
	"context"
	"testing"

	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
)

func TestNewCopilotClientRejectsNilLogger(t *testing.T) {
	client, err := NewCopilotClient(nil, "", "", "")
	if err == nil {
		t.Fatalf("expected error for nil logger")
	}
	if client != nil {
		t.Fatalf("expected nil client when constructor fails")
	}
}

func TestNewCopilotClientCreatesClient(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewCopilotClient(appLogger, "", "", "")
	if err != nil {
		t.Fatalf("new copilot client: %v", err)
	}
	if client == nil {
		t.Fatalf("expected non-nil client")
	}
	if client.Client() == nil {
		t.Fatalf("expected non-nil underlying sdk client")
	}
}

func TestCopilotClientClientNilReceiver(t *testing.T) {
	var client *CopilotClient
	if client.Client() != nil {
		t.Fatalf("expected nil sdk client for nil receiver")
	}
}

func TestCopilotClientStartRejectsUninitializedClient(t *testing.T) {
	client := &CopilotClient{}
	if err := client.Start(context.Background()); err == nil {
		t.Fatalf("expected error for uninitialized client")
	}
}

func TestCopilotClientStopAllowsNilReceiver(t *testing.T) {
	var client *CopilotClient
	if err := client.Stop(); err != nil {
		t.Fatalf("expected nil error for nil receiver: %v", err)
	}
}

func TestCopilotClientRunPromptRejectsEmptyPrompt(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	client, err := NewCopilotClient(appLogger, "", "", "")
	if err != nil {
		t.Fatalf("new copilot client: %v", err)
	}

	if _, err := client.RunPrompt(context.Background(), "", ""); err == nil {
		t.Fatalf("expected error for empty prompt")
	}
}

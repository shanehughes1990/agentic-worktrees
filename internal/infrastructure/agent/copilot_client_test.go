package agent

import (
	"context"
	"strings"
	"testing"

	domainservices "github.com/shanehughes1990/agentic-worktrees/internal/domain/services"
	infralogger "github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/logger"
)

func TestNewCopilotClientRejectsNilLogger(t *testing.T) {
	client, err := NewCopilotClient(nil, "", "", "")
	if err == nil {
		t.Fatalf("expected nil logger error")
	}
	if client != nil {
		t.Fatalf("expected nil client")
	}
}

func TestCopilotClientStartRejectsUninitializedClient(t *testing.T) {
	client := &CopilotClient{}
	if err := client.Start(context.Background()); err == nil {
		t.Fatalf("expected uninitialized error")
	}
}

func TestCopilotClientRunPromptRejectsEmptyPrompt(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	client, err := NewCopilotClient(appLogger, "", "", "")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.runPrompt(context.Background(), "", ""); err == nil {
		t.Fatalf("expected empty prompt error")
	}
}

func TestResolveCopilotModelDefaultsToGPT53Codex(t *testing.T) {
	if got := resolveCopilotModel(""); got != DefaultCopilotModel {
		t.Fatalf("expected default model, got %q", got)
	}
	if got := resolveCopilotModel("   "); got != DefaultCopilotModel {
		t.Fatalf("expected default model for whitespace, got %q", got)
	}
}

func TestResolveCopilotModelAllowsOverride(t *testing.T) {
	if got := resolveCopilotModel("gpt-5.4"); got != "gpt-5.4" {
		t.Fatalf("expected explicit override model, got %q", got)
	}
}

func TestBuildGenerateTaskBoardPromptIncludesDocuments(t *testing.T) {
	prompt := buildGenerateTaskBoardPrompt("build board", []domainservices.DocumentationSourceFile{{Path: "docs/a.md", Content: "A"}})
	if !strings.Contains(prompt, "build board") {
		t.Fatalf("expected base prompt")
	}
	if !strings.Contains(prompt, "DOCUMENT START: docs/a.md") {
		t.Fatalf("expected document marker")
	}
}

func TestGenerateTaskBoardValidation(t *testing.T) {
	appLogger, err := infralogger.New("trace", "text")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	client, err := NewCopilotClient(appLogger, "", "", "")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.GenerateTaskBoard(context.Background(), domainservices.GenerateTaskBoardRequest{}); err == nil {
		t.Fatalf("expected request validation error")
	}
}

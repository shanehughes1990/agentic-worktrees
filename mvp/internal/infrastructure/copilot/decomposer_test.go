package copilot

import (
	"context"
	"testing"

	appcopilot "github.com/shanehughes1990/agentic-worktrees/internal/application/copilot"
)

func TestDecomposeRequiresPrompt(t *testing.T) {
	decomposer := NewDecomposer(ClientConfig{}, nil)
	_, err := decomposer.Decompose(context.Background(), appcopilot.DecomposeRequest{RunID: "run-1", Prompt: "   "})
	if err == nil {
		t.Fatalf("expected prompt validation error")
	}
}

func TestClientConfigDefaultModel(t *testing.T) {
	cfg := (ClientConfig{}).Normalized()
	if cfg.DefaultModel != "gpt-5.3-codex" {
		t.Fatalf("expected default model gpt-5.3-codex, got %s", cfg.DefaultModel)
	}
}

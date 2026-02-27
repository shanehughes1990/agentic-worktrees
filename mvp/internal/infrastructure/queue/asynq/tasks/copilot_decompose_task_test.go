package tasks

import (
	"encoding/json"
	"testing"
)

func TestNewCopilotDecomposeTaskValidatesInput(t *testing.T) {
	if _, _, err := NewCopilotDecomposeTask(CopilotDecomposePayload{Prompt: "x"}); err == nil {
		t.Fatalf("expected run_id validation error")
	}
	if _, _, err := NewCopilotDecomposeTask(CopilotDecomposePayload{RunID: "run"}); err == nil {
		t.Fatalf("expected prompt validation error")
	}
}

func TestNewCopilotDecomposeTaskBuildsTask(t *testing.T) {
	task, options, err := NewCopilotDecomposeTask(CopilotDecomposePayload{RunID: "run", Prompt: "prompt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type() != TaskTypeCopilotDecompose {
		t.Fatalf("unexpected task type: %s", task.Type())
	}
	if len(options) == 0 {
		t.Fatalf("expected default queue option")
	}
	payload := CopilotDecomposePayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unexpected payload decode error: %v", err)
	}
	if payload.IdempotencyKey == "" {
		t.Fatalf("expected idempotency key fallback to be populated")
	}
}

func TestNewCopilotDecomposeTaskDoesNotRequireAdditionalIngestionInputs(t *testing.T) {
	task, _, err := NewCopilotDecomposeTask(CopilotDecomposePayload{
		RunID:  "run-legacy",
		Prompt: "prompt",
	})
	if err != nil {
		t.Fatalf("expected legacy ingestion payload to remain valid: %v", err)
	}

	payload := CopilotDecomposePayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		t.Fatalf("unexpected payload decode error: %v", err)
	}
	if payload.Model != "" || payload.WorkingDirectory != "" || len(payload.SkillDirectories) != 0 || payload.GithubToken != "" || payload.CLIPath != "" || payload.CLIURL != "" {
		t.Fatalf("expected optional ingestion inputs to remain optional, got %#v", payload)
	}
}

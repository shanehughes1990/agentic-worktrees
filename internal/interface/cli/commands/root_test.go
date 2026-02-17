package commands

import "testing"

func TestResolveBoardPromptDefaults(t *testing.T) {
	if got := resolveBoardPrompt(""); got != DefaultBoardPrompt {
		t.Fatalf("expected default prompt, got %q", got)
	}
	if got := resolveBoardPrompt("   "); got != DefaultBoardPrompt {
		t.Fatalf("expected default prompt for whitespace, got %q", got)
	}
}

func TestResolveBoardPromptOverride(t *testing.T) {
	override := "Use these docs to create a kanban backlog."
	if got := resolveBoardPrompt(override); got != override {
		t.Fatalf("expected override prompt, got %q", got)
	}
}

package taskboard

import (
	"strings"
	"testing"
)

func TestBuildTaskboardPromptIncludesNormalizedDocumentsInStableOrder(t *testing.T) {
	prompt := BuildTaskboardPrompt("repo", NormalizedDocument{
		RelativePath: "z.md",
		Content:      "z-content",
	}, NormalizedDocument{
		RelativePath: "a.md",
		Content:      "a-content",
	})

	first := strings.Index(prompt, "path: a.md")
	second := strings.Index(prompt, "path: z.md")
	if first < 0 || second < 0 || first > second {
		t.Fatalf("expected prompt documents in sorted order, got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Prioritize facts from the normalized UTF-8 documents listed below.") {
		t.Fatalf("expected normalized document instruction in prompt")
	}
}

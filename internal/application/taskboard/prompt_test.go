package taskboard

import (
	"strings"
	"testing"
	"time"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
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
	if !strings.Contains(prompt, "You are a senior project manager for an autonomous engineering team.") {
		t.Fatalf("expected project manager persona in prompt")
	}
	if !strings.Contains(prompt, "Do not create duplicate or near-duplicate tasks") {
		t.Fatalf("expected anti-duplication instruction in prompt")
	}
	if !strings.Contains(prompt, "Never output parallel tasks that could implement the same thing.") {
		t.Fatalf("expected anti-overlap parallelism guard in prompt")
	}
	if !strings.Contains(prompt, "Assign one owning layer per behavior change") {
		t.Fatalf("expected single-owner-per-behavior instruction in prompt")
	}
	if !strings.Contains(prompt, "Separate implementation work from verification work") {
		t.Fatalf("expected implementation-vs-verification separation instruction in prompt")
	}
	if !strings.Contains(prompt, "one-check-per-task") {
		t.Fatalf("expected one-check-per-task instruction in prompt")
	}
}

func TestBuildBoardSupervisorPromptIncludesBoardAndConstraints(t *testing.T) {
	board := &domaintaskboard.Board{
		BoardID:   "board-1",
		RunID:     "run-1",
		Status:    domaintaskboard.StatusNotStarted,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Epics: []domaintaskboard.Epic{
			{
				WorkItem: domaintaskboard.WorkItem{ID: "e1", BoardID: "board-1", Title: "Epic", Status: domaintaskboard.StatusNotStarted},
				Tasks: []domaintaskboard.Task{
					{WorkItem: domaintaskboard.WorkItem{ID: "t1", BoardID: "board-1", Title: "Task", Status: domaintaskboard.StatusNotStarted}},
				},
			},
		},
	}

	prompt := BuildBoardSupervisorPrompt("original prompt", board)
	if !strings.Contains(prompt, "You are a taskboard quality supervisor.") {
		t.Fatalf("expected supervisor persona in prompt")
	}
	if !strings.Contains(prompt, "Remove duplicate or near-duplicate tasks") {
		t.Fatalf("expected anti-duplicate instruction in supervisor prompt")
	}
	if !strings.Contains(prompt, "Candidate board JSON") {
		t.Fatalf("expected candidate board section in supervisor prompt")
	}
	if !strings.Contains(prompt, "Current quality report") {
		t.Fatalf("expected quality report section in supervisor prompt")
	}
	if !strings.Contains(prompt, "split bundled checks") {
		t.Fatalf("expected bundled-check splitting instruction in supervisor prompt")
	}
	if !strings.Contains(prompt, "\"board_id\": \"board-1\"") {
		t.Fatalf("expected serialized board payload in supervisor prompt")
	}
}

package taskboard

import (
	"fmt"
	"sort"
	"strings"
)

func BuildTaskboardPrompt(directory string, documents ...NormalizedDocument) string {
	sortedDocuments := append([]NormalizedDocument(nil), documents...)
	sort.SliceStable(sortedDocuments, func(left, right int) bool {
		return sortedDocuments[left].RelativePath < sortedDocuments[right].RelativePath
	})

  prompt := fmt.Sprintf(`You are a senior project manager for an autonomous engineering team.

Analyze the code and documentation in directory %q.

Return ONLY valid JSON (no markdown, no prose) in this schema:
{
  "board_id": "<string>",
  "run_id": "<string>",
  "title": "<short board title>",
  "goal": "<overall goal>",
  "status": "not-started|in-progress|completed|blocked",
  "epics": [
    {
      "id": "<epic-id>",
      "board_id": "<string>",
      "title": "<epic title>",
      "status": "not-started|in-progress|completed|blocked",
      "depends_on": ["<epic-id>", "..."],
      "tasks": [
        {
          "id": "<task-id>",
          "board_id": "<string>",
          "title": "<small executable task>",
          "status": "not-started|in-progress|completed|blocked",
          "depends_on": ["<task-id>", "..."],
          "outcome": {
            "status": "merged|no_changes|failed|precompleted",
            "reason": "<short explanation>"
          }
        }
      ]
    }
  ],
  "created_at": "RFC3339 datetime",
  "updated_at": "RFC3339 datetime"
}

Execution planning requirements:
- Decompose work into the smallest practical execution chunks.
- Each task must produce one clear, non-overlapping outcome.
- Tasks must be small enough for a single agent to execute independently.
- Model task dependencies explicitly in task.depends_on.
- Keep dependencies acyclic.
- Encode concurrency by leaving independent tasks without unnecessary dependencies.
- Put first-executable tasks early in each epic.tasks array.
- Ensure IDs are unique across epics and tasks.
- Do not create duplicate or near-duplicate tasks (including renamed variants with the same intent).
- If two candidate tasks touch the same artifact, behavior, or acceptance outcome, merge them into one task OR connect them with an explicit dependency.
- Never output parallel tasks that could implement the same thing.
- Prefer one canonical task per artifact/outcome, then fan out only truly distinct follow-up work.
- Each task title must be specific enough to distinguish it from every other task title in the board.
- Prioritize facts from the normalized UTF-8 documents listed below.
`, directory)
	if len(sortedDocuments) == 0 {
		return prompt + "\nNo normalized documents were extracted; inspect the directory directly.\n"
	}

	var builder strings.Builder
	builder.WriteString(prompt)
	builder.WriteString("\nNormalized UTF-8 documents (stable path order):\n")
	for _, document := range sortedDocuments {
		relativePath := strings.TrimSpace(document.RelativePath)
		if relativePath == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("\n---\npath: %s\ncontent:\n%s\n", relativePath, document.Content))
	}
	return builder.String()
}

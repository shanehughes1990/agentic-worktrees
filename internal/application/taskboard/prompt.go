package taskboard

import "fmt"

func BuildTaskboardPrompt(directory string) string {
	return fmt.Sprintf(`Analyze the code and documentation in directory %q.

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
          "depends_on": ["<task-id>", "..."]
        }
      ]
    }
  ],
  "created_at": "RFC3339 datetime",
  "updated_at": "RFC3339 datetime"
}

Execution planning requirements:
- Tasks must be small enough for a single agent to execute independently.
- Model task dependencies explicitly in task.depends_on.
- Keep dependencies acyclic.
- Encode concurrency by leaving independent tasks without unnecessary dependencies.
- Put first-executable tasks early in each epic.tasks array.
- Ensure IDs are unique across epics and tasks.
`, directory)
}

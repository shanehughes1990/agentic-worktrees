package taskboard

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domaintaskboard "github.com/shanehughes1990/agentic-worktrees/internal/domain/taskboard"
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
- Separate implementation work from verification work: implementation tasks change behavior; test tasks verify it.
- For verification tasks, enforce one-check-per-task; do not bundle multiple independent checks in a single test task.
- Example: split "fake-provider behavior + lifecycle parity" into two test tasks.
- Example: split "filesystem parity tests + error-mapping tests" into two test tasks.
- Model task dependencies explicitly in task.depends_on.
- Keep dependencies acyclic.
- Encode concurrency by leaving independent tasks without unnecessary dependencies.
- Put first-executable tasks early in each epic.tasks array.
- Ensure IDs are unique across epics and tasks.
- Do not create duplicate or near-duplicate tasks (including renamed variants with the same intent).
- If two candidate tasks touch the same artifact, behavior, or acceptance outcome, merge them into one task OR connect them with an explicit dependency.
- Never output parallel tasks that could implement the same thing.
- Prefer one canonical task per artifact/outcome, then fan out only truly distinct follow-up work.
- Assign one owning layer per behavior change (domain OR application OR infrastructure), not multiple parallel owner tasks for the same behavior.
- Do not create separate "enforce/preserve/wire default" tasks for the same outcome across multiple layers; pick one implementation owner and express the rest as dependency-aware verification.
- Do not split the same file/package touch into sibling tasks unless one task explicitly depends on the other and has a distinct acceptance outcome.
- Each task title must be specific enough to distinguish it from every other task title in the board.
- Use task titles that include concrete scope (artifact/package + intent), so similarly named tasks are not produced.
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

func BuildBoardSupervisorPrompt(sourcePrompt string, board *domaintaskboard.Board) string {
  return BuildBoardSupervisorPromptWithReport(sourcePrompt, board, BoardQualityReport{})
}

func BuildBoardSupervisorPromptWithReport(sourcePrompt string, board *domaintaskboard.Board, report BoardQualityReport) string {
  serializedBoard := "{}"
  if board != nil {
    if payload, err := json.MarshalIndent(board, "", "  "); err == nil {
      serializedBoard = string(payload)
    }
  }

  serializedReport := "{}"
  if payload, err := json.MarshalIndent(report, "", "  "); err == nil {
    serializedReport = string(payload)
  }

  return fmt.Sprintf(`You are a taskboard quality supervisor.

Review the candidate board JSON and return an improved board JSON.

Return ONLY valid JSON matching the same board schema.

Supervisor requirements:
- Preserve the same scope and intent from the original request.
- Remove duplicate or near-duplicate tasks.
- Remove overlap/conflict-prone sibling tasks that touch the same artifact/outcome in parallel.
- Prefer one implementation owner task per behavior change; represent additional checks as verification tasks.
- For verification tasks, enforce one-check-per-task and split bundled checks into separate tasks with explicit dependencies.
- Keep dependencies acyclic and explicit.
- Keep tasks execution-ready and specific (artifact/package + action).
- Do not add out-of-scope work.
- Ensure IDs remain unique.
- If a quality report is provided, you MUST resolve every critical failure and improve category scores above threshold.

Original decomposition prompt context:
%s

Current quality report:
%s

Candidate board JSON:
%s
`, strings.TrimSpace(sourcePrompt), serializedReport, serializedBoard)
}

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	urcli "github.com/urfave/cli/v3"

	"github.com/shanehughes1990/agentic-worktrees/internal/application"
)

const DefaultBoardPrompt = `You are generating a delivery-ready task board from authoritative project documentation.

Scope invariance:
- The provided documentation set may describe an entire project OR a very small single task.
- Scope size does not change the requirement: fully deconstruct the provided files into a vividly detailed task board.
- Use only the provided files as source of truth.

Output format contract (MUST follow exactly):
- Return a single JSON object that matches this board schema shape:
  {
    "id": "string",
    "title": "string",
    "epics": [
      {
        "id": "string",
        "title": "string",
        "description": "string",
        "dependencies": ["epic_id", "..."],
        "tasks": [
          {
            "id": "string",
            "title": "string",
            "description": "string",
            "status": "pending|in_progress|completed|failed",
            "dependencies": ["task_id", "..."]
          }
        ]
      }
    ],
    "created_at": "RFC3339 timestamp",
    "updated_at": "RFC3339 timestamp"
  }
- Do not add any extra top-level or nested fields.
- All generated tasks should initialize with status "pending".

Goal:
- Build an expressive epic -> task -> micro-task plan that is execution-ready.
- Prefer many small, leaf-level tasks over broad, vague tasks.
- Keep each leaf task independently executable and verifiable.

Required planning method:
1) Determine the documentation scope represented by the provided files.
2) Identify epics from that scope.
3) Decompose each epic into tasks.
4) Decompose each task into micro leaf tasks that can be executed directly.
5) Continue decomposition until each leaf task has clear completion criteria.

Concurrency and dependency modeling:
- Model strict ordering using task "dependencies" arrays (task IDs).
- Model parallelism by leaving independent tasks without blocking dependencies.
- Use epic "dependencies" arrays when an entire epic is blocked by another epic.
- Never leave dependency intent ambiguous.
- Maximize safe parallelism while preserving correctness.

Leaf task quality requirements:
- Each leaf task must include:
  - concise action title
  - vivid implementation detail in description
  - expected artifact/output and validation criteria in description
  - dependency references via "dependencies"
- Leaf tasks should be small enough to complete quickly.
- Avoid umbrella tasks that hide multiple implementation steps.

Execution-tree requirement:
- Construct a tree where internal nodes are epics/tasks and leaves are executable micro tasks.
- The board is not complete until all work is represented by leaf tasks.
- Ensure parent-child relationships are explicit and complete.

Output intent:
- Produce a board that an automated runner can execute iteratively until completion.
- The resulting plan must make ordering, concurrency, and completion state unambiguous.
- The decomposition must remain vividly detailed regardless of whether scope is tiny or large.`

type GenerateTaskBoardRunFunc func(ctx context.Context, input application.GenerateTaskBoardInput) (string, error)

type GenerateTaskBoardCommandOptions struct {
	RootDirectory  *string
	TraversalDepth *int
	BoardPrompt    *string
	BoardModel     *string
	LoggerProvider func() *logrus.Logger
}

func NewGenerateTaskBoardCommand(runGenerateTaskBoard GenerateTaskBoardRunFunc, options GenerateTaskBoardCommandOptions) *urcli.Command {
	return &urcli.Command{
		Name:  "generate-task-board",
		Usage: "enqueue task-board generation from local docs payload",
		Flags: []urcli.Flag{
			&urcli.StringFlag{Name: "ROOT_DIRECTORY", Value: "docs", Destination: options.RootDirectory, Sources: urcli.EnvVars("ROOT_DIRECTORY")},
			&urcli.IntFlag{Name: "MAX_DEPTH", Value: 2, Destination: options.TraversalDepth, Sources: urcli.EnvVars("MAX_DEPTH")},
			&urcli.StringFlag{Name: "PROMPT", Value: DefaultBoardPrompt, Destination: options.BoardPrompt, Sources: urcli.EnvVars("PROMPT")},
			&urcli.StringFlag{Name: "MODEL", Value: "gpt-5.3-codex", Destination: options.BoardModel, Sources: urcli.EnvVars("MODEL")},
		},
		Action: func(actionCtx context.Context, _ *urcli.Command) error {
			if runGenerateTaskBoard == nil {
				return fmt.Errorf("generate task board runner cannot be nil")
			}
			if options.RootDirectory == nil || options.TraversalDepth == nil || options.BoardPrompt == nil || options.BoardModel == nil {
				return fmt.Errorf("generate-task-board command options are not initialized")
			}

			taskID, err := runGenerateTaskBoard(actionCtx, application.GenerateTaskBoardInput{
				RootDirectory: *options.RootDirectory,
				MaxDepth:      *options.TraversalDepth,
				Prompt:        resolveBoardPrompt(*options.BoardPrompt),
				Model:         *options.BoardModel,
			})
			if err != nil {
				return err
			}
			if options.LoggerProvider != nil {
				if logger := options.LoggerProvider(); logger != nil {
					logger.WithField("task_id", taskID).Info("enqueued generate-task-board")
				}
			}
			return nil
		},
	}
}

func resolveBoardPrompt(prompt string) string {
	resolvedPrompt := strings.TrimSpace(prompt)
	if resolvedPrompt == "" {
		return DefaultBoardPrompt
	}
	return resolvedPrompt
}

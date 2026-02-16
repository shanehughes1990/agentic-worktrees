package mcp

import (
	"context"
	"fmt"

	"github.com/shanehughes1990/agentic-worktrees/internal/board"
	"github.com/shanehughes1990/agentic-worktrees/internal/queue"
	"github.com/shanehughes1990/agentic-worktrees/internal/runstate"
)

type Adapter struct {
	QueueClient *queue.Client
	QueueName   string
	Metrics     *queue.Metrics
	Checkpoints *runstate.Store
}

func (a *Adapter) Call(ctx context.Context, tool string, input map[string]any) (map[string]any, error) {
	switch tool {
	case "ingest_scope":
		scopePath, _ := input["scope_path"].(string)
		outputPath, _ := input["output_path"].(string)
		if scopePath == "" {
			return nil, fmt.Errorf("scope_path is required")
		}
		if outputPath == "" {
			outputPath = "state/board.json"
		}
		built, err := board.BuildBoardFromFile(scopePath)
		if err != nil {
			return nil, err
		}
		repo := board.NewRepository(outputPath)
		if err := repo.Write(built); err != nil {
			return nil, err
		}
		return map[string]any{"ok": true, "board_path": outputPath, "tasks": len(built.Tasks)}, nil
	case "enqueue_task":
		runID, _ := input["run_id"].(string)
		taskID, _ := input["task_id"].(string)
		worktree, _ := input["worktree"].(string)
		prompt, _ := input["prompt"].(string)
		originBranch, _ := input["origin_branch"].(string)

		task, err := queue.NewLifecycleTask(queue.TypePrepareWorktree, queue.LifecyclePayload{
			RunID:        runID,
			TaskID:       taskID,
			WorktreeName: worktree,
			Prompt:       prompt,
			OriginBranch: originBranch,
		}, a.QueueName)
		if err != nil {
			return nil, err
		}
		if a.QueueClient == nil {
			return nil, fmt.Errorf("queue client not configured")
		}
		info, err := a.QueueClient.Enqueue(task)
		if err != nil {
			return nil, err
		}
		return map[string]any{"ok": true, "job_id": info.ID, "type": info.Type, "queue": info.Queue}, nil
	case "status":
		if a.Metrics == nil || a.Checkpoints == nil {
			return nil, fmt.Errorf("status dependencies are not configured")
		}
		summary, err := a.Checkpoints.Summary()
		if err != nil {
			return nil, err
		}
		rows, err := a.Checkpoints.All()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"ok":                   true,
			"metrics":              a.Metrics.Snapshot(),
			"checkpoint_by_status": summary,
			"checkpoint_rows":      len(rows),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported tool %q", tool)
	}
}

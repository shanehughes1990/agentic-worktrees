package queue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
)

const (
	TypePrepareWorktree = "task:prepare_worktree"
	TypeExecuteAgent    = "task:execute_agent"
	TypeValidate        = "task:validate"
	TypeOpenOrUpdatePR  = "task:open_or_update_pr"
	TypeRebaseAndMerge  = "task:rebase_and_merge"
	TypeCleanup         = "task:cleanup"
)

type LifecyclePayload struct {
	RunID        string `json:"run_id"`
	TaskID       string `json:"task_id"`
	WorktreeName string `json:"worktree_name"`
	Prompt       string `json:"prompt"`
	OriginBranch string `json:"origin_branch"`
}

func NewLifecycleTask(taskType string, payload LifecyclePayload, queue string) (*asynq.Task, error) {
	payload.RunID = strings.TrimSpace(payload.RunID)
	payload.TaskID = strings.TrimSpace(payload.TaskID)
	payload.WorktreeName = strings.TrimSpace(payload.WorktreeName)
	payload.Prompt = strings.TrimSpace(payload.Prompt)
	payload.OriginBranch = strings.TrimSpace(payload.OriginBranch)

	if !isSupportedType(taskType) {
		return nil, fmt.Errorf("unsupported task type %q", taskType)
	}
	if payload.RunID == "" {
		return nil, fmt.Errorf("run id cannot be empty")
	}
	if payload.TaskID == "" {
		return nil, fmt.Errorf("task id cannot be empty")
	}
	if payload.WorktreeName == "" {
		payload.WorktreeName = payload.TaskID
	}
	if payload.OriginBranch == "" {
		payload.OriginBranch = "main"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if strings.TrimSpace(queue) == "" {
		return asynq.NewTask(taskType, body), nil
	}
	return asynq.NewTask(taskType, body, asynq.Queue(queue)), nil
}

func ParseLifecycleTaskPayload(task *asynq.Task) (LifecyclePayload, error) {
	var payload LifecyclePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return LifecyclePayload{}, fmt.Errorf("invalid payload: %w", err)
	}
	payload.RunID = strings.TrimSpace(payload.RunID)
	payload.TaskID = strings.TrimSpace(payload.TaskID)
	payload.WorktreeName = strings.TrimSpace(payload.WorktreeName)
	payload.Prompt = strings.TrimSpace(payload.Prompt)
	payload.OriginBranch = strings.TrimSpace(payload.OriginBranch)

	if payload.RunID == "" {
		return LifecyclePayload{}, fmt.Errorf("payload missing run_id")
	}
	if payload.TaskID == "" {
		return LifecyclePayload{}, fmt.Errorf("payload missing task_id")
	}
	if payload.WorktreeName == "" {
		payload.WorktreeName = payload.TaskID
	}
	if payload.OriginBranch == "" {
		payload.OriginBranch = "main"
	}
	return payload, nil
}

func isSupportedType(taskType string) bool {
	switch taskType {
	case TypePrepareWorktree, TypeExecuteAgent, TypeValidate, TypeOpenOrUpdatePR, TypeRebaseAndMerge, TypeCleanup:
		return true
	default:
		return false
	}
}

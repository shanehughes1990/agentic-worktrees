# 04 - Workflow and State Machine

## End-to-End Automation Loop

For each eligible task:

1. Capture run-start context (origin target branch, repo revision, policy set)
2. Run preflight admission checks (`git`/`gh` version + `gh` auth status)
3. Validate persisted schema versions and backward-compatibility support for active artifacts
4. If preflight fails, block intake and enqueueing until healthy
5. Evaluate dependency graph and mark runnable tasks as `ready`
6. Enqueue `task.prepare_worktree` in Asynq with deterministic payload
7. Worker claims job and writes `worktree_preparing` checkpoint
8. On success, enqueue `task.execute_agent`
9. Worker runs agent execution and emits realtime progress events
10. On success, enqueue `task.validate`
11. On success, enqueue `task.open_or_update_pr`
12. On success, enqueue `task.rebase_and_merge` (targeting captured run-start origin branch)
13. On success, enqueue `task.cleanup`
14. Mark task `done`; otherwise retry or archive/dead-letter by policy
15. Repeat continuously until board completion or terminal escalation

On failure during any stage:

- preserve current worktree state and task branch by default
- resume from last successful checkpoint in the same task context
- only discard/recreate context when policy classifies failure as extreme/terminal

## Deterministic State Machine

### Task States

- `backlog` → not yet schedulable
- `ready` → dependencies satisfied
- `in_progress` → assigned to worker/run
- `blocked` → waiting on dependency or external event
- `review` → PR/checks in progress
- `done` → successfully merged
- `failed` → exhausted retries or terminal failure

### Run States

- `queued`
- `dequeued`
- `worktree_preparing`
- `agent_running`
- `validating`
- `pr_open`
- `rebasing`
- `conflict_resolution`
- `merging`
- `cleanup`
- `retry_scheduled`
- `resume_pending`
- `resuming`
- `dead_lettered`
- `completed`
- `errored`
- `cancelled`

Each transition emits an event and checkpoint write.

## Idempotency Rules

- Re-running `prepare_worktree` must not duplicate worktrees
- Re-running `open_or_update_pr` must update existing PR when present
- Re-running `rebase_and_merge` must detect already-merged state
- Cleanup operations are best-effort and retryable
- Every handler must tolerate at-least-once delivery from the queue

## Checkpoint Boundaries

Persist state after:

- enqueue accepted (job id + payload hash + queue)
- dequeue claimed (worker identity)
- worktree created/recovered
- agent execution started/stopped
- commit pushed
- PR opened/updated
- rebase finished
- merge attempt result
- retry scheduled/archived reason

These boundaries allow crash-safe resume.

Resume behavior must be checkpoint-directed and work-preserving; the system should continue from the furthest valid checkpoint rather than restarting entire task lifecycle unless required.

## Control Plane Operations

- `pause queue`
- `resume queue`
- `cancel run`
- `retry task`
- `drain workers`
- `archive/dead-letter inspect`

All control actions are audit logged.

In autonomous walk-away mode, these operations are expected to be used only for exceptional escalations, not routine execution.

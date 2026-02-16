# 04 - Workflow and State Machine

## End-to-End Automation Loop

For each eligible task:

1. Run preflight admission checks (`git`/`gh` version + `gh` auth status)
2. Validate persisted schema versions and backward-compatibility support for active artifacts
3. If preflight fails, block intake and enqueueing until healthy
4. Evaluate dependency graph and mark runnable tasks as `ready`
5. Enqueue `task.prepare_worktree` in Asynq with deterministic payload
6. Worker claims job and writes `worktree_preparing` checkpoint
7. On success, enqueue `task.execute_agent`
8. Worker runs agent execution and emits realtime progress events
9. On success, enqueue `task.validate`
10. On success, enqueue `task.open_or_update_pr`
11. On success, enqueue `task.rebase_and_merge`
12. On success, enqueue `task.cleanup`
13. Mark task `done`; otherwise retry or archive/dead-letter by policy
14. Repeat continuously with latest origin freshness guarantees

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

## Control Plane Operations

- `pause queue`
- `resume queue`
- `cancel run`
- `retry task`
- `drain workers`
- `archive/dead-letter inspect`

All control actions are audit logged.

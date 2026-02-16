# 05 - Git Strategy and Conflict Handling

## Branching Model

- Base branch: configurable (default `main`)
- Task branch naming: `task/<task-id>-<slug>`
- One active task branch per task execution context

See [15 - Git Worktree Task Lifecycle](./15-git-worktree-task-lifecycle.md) for the end-to-end parallel session flow and lock/rebase/PR sequence.

## Mandatory Freshness Rule

Every iteration must begin with latest origin:

1. `git fetch --prune`
2. refresh local tracking branch from `origin/<base>`
3. rebase task branch onto latest `origin/<base>` before agent run
4. rebase again before merge attempt

This reduces long-lived drift and surprise conflicts.

## Conflict Handling Policy Matrix

### Policy A: Auto-Safe

Auto-resolve only if all conditions hold:

- conflict limited to allowlisted files/patterns
- resolution strategy deterministic
- tests/policy checks pass after resolution

Else escalate.

### Policy B: Agent-Assisted

Agent proposes resolution patch with explicit reasoning artifact. System verifies:

- `git diff` scope against conflict set
- formatting/lint/test gates
- policy compliance

If validation fails, escalate.

### Policy C: Manual Escalation

Pause run, mark task `blocked`, attach conflict bundle:

- conflicted file list
- base/head patch context
- previous auto-resolution attempts

## Merge Safety Gates

Merge allowed only when:

- PR status checks are green
- branch is rebased and up to date
- no unresolved conflict markers
- policy checks pass
- mergeability is confirmed by remote provider

## Recovery Scenarios

- Rebase interrupted: abort and restore checkpointed branch state
- Non-fast-forward push: fetch/rebase/retry with bounded attempts
- Force-push policy violations: fail safe and escalate

## Queue-Coupled Git Execution Rules

- Git operations run inside Asynq lifecycle handlers with idempotency keys
- Retryable Git failures must return transient classification to queue runtime
- Non-retryable Git failures must return terminal classification and include remediation metadata
- Merge-critical handlers must checkpoint before and after every side effect

## Audit Requirements

For every conflict event, store:

- conflict classification
- attempted strategy
- operator/agent identity
- final outcome

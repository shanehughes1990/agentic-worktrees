# Git Worktree Task Execution + Merge-Back Architecture

## 1) Purpose

This document defines the architecture for executing each task in an isolated Git worktree, then merging only that task’s changes back into the branch that initiated the run.

Primary outcome:

- Start from a caller-provided source branch.
- Create a dedicated worktree per task.
- Execute implementation in isolation.
- Merge back into the same source branch after validation.
- Resolve merge conflicts with Copilot SDK guidance while preserving branch safety.

## 2) Scope and Non-Goals

### In Scope

- Source-branch capture at run start.
- Worktree creation, branch naming, and workspace isolation.
- Per-task execution workflow inside worktree.
- Merge-back rules to the originating branch only.
- Guardrails to prevent unrelated task changes from entering merge.
- Conflict detection and Copilot SDK-assisted conflict resolution.
- Audit trail for task branch, worktree path, merge result, and conflict artifacts.

### Out of Scope

- Multi-repo orchestration.
- Cross-task squashing/cherry-picking strategies beyond task-local branch merge.
- Replacing Git merge semantics with custom patch engines.

## 3) Domain Terminology

- **SourceBranch**: the branch active when a task run starts (merge target).
- **TaskBranch**: branch dedicated to exactly one task.
- **TaskWorktree**: filesystem checkout for `TaskBranch`.
- **TaskMergeCandidate**: validated commit range produced by a task.
- **ConflictSet**: files/regions reported by Git as merge conflicts.
- **ConflictResolutionRun**: Copilot SDK-guided resolution attempt over a `ConflictSet`.

## 4) Invariants and Safety Rules

1. One task => one task branch => one worktree.
2. Merge target is always the captured `SourceBranch` for that run.
3. Task branch must contain only task-scoped commits.
4. Never merge or pull changes from unrelated task branches.
5. Conflict resolution must preserve code correctness and pass required tests before merge.
6. Worktree cleanup is mandatory after terminal success/failure.

## 5) Layered Architecture (DDD aligned)

### Interface Layer

Responsibilities:

- Accept task execution request (task id, source branch, repository root).
- Show progress and merge status.
- Surface conflict files and resolution outcomes.

### Application Layer

Use-cases:

- `StartTaskInWorktree`:
  - capture source branch
  - create task branch + worktree
  - invoke execution pipeline
- `MergeTaskBack`:
  - validate candidate
  - merge into source branch
  - trigger conflict flow if needed
- `ResolveMergeConflictsWithCopilot`:
  - collect conflict context
  - invoke Copilot SDK resolution prompt
  - apply patches and re-validate

### Domain Layer

Core models and rules:

- `TaskExecutionSession`
- `MergePolicy`
- `ConflictResolutionPolicy`
- Validation rules for allowed merge scope and commit ancestry.

### Infrastructure Layer

Adapters:

- Git adapter (`worktree add/remove`, `checkout`, `merge`, conflict introspection).
- Copilot SDK adapter for resolution suggestions.
- Persistence/audit adapter for task run metadata.

## 6) End-to-End Flow

1. Capture `SourceBranch` (e.g., `revamp`).
2. Create `TaskBranch` from `SourceBranch`.
3. Create `TaskWorktree` at deterministic path.
4. Execute task implementation and tests inside worktree.
5. Validate only task-scoped changes are present.
6. Checkout `SourceBranch` in primary repo context.
7. Merge `TaskBranch` into `SourceBranch`.
8. If conflict-free:
   - run post-merge validation
   - finalize and clean worktree.
9. If conflicts:
   - build `ConflictSet`
   - run Copilot SDK conflict-resolution workflow
   - apply and validate
   - complete merge and cleanup.

## 7) Merge Safety Model

### Branch Provenance

- `TaskBranch` must be created from captured `SourceBranch` head (or pinned commit SHA).
- Merge is rejected if branch ancestry shows unrelated merge parents from other task branches.

### Change Isolation

- Pre-merge check compares diff scope against task ownership metadata.
- Out-of-scope files fail merge gate.

### Allowed Merge Direction

- Allowed: `TaskBranch -> SourceBranch`.
- Forbidden: `TaskBranch -> other task branch` and `other task branch -> TaskBranch`.

## 8) Conflict Resolution with Copilot SDK

### Inputs to Copilot SDK

- Base/ours/theirs content for each conflict region.
- File path and language context.
- Task objective and acceptance criteria.
- Nearby symbols/tests for semantic correctness.

### Resolution Loop

1. Request minimal conflict patch from Copilot SDK.
2. Apply patch to worktree file.
3. Re-run compile/tests for impacted scope.
4. If failing, retry with focused diagnostics.
5. Mark resolved only after all conflict markers removed and validations pass.

### Guardrails

- No broad refactors during conflict resolution.
- No unrelated file edits.
- No acceptance of unresolved markers (`<<<<<<<`, `=======`, `>>>>>>>`).

## 9) Data and Audit Model

Recommended persisted record per task run:

- `task_id`
- `source_branch`
- `task_branch`
- `worktree_path`
- `start_commit`
- `end_commit`
- `merge_status` (`pending`, `merged`, `conflicted`, `failed`)
- `conflict_files[]`
- `resolution_attempts`
- `validation_summary`
- timestamps and operator/agent metadata.

## 10) Failure Handling

- Worktree creation failure: abort run, no branch mutation.
- Merge conflict unresolved after retry budget: mark failed, keep artifacts for manual takeover.
- Validation failure post-resolution: rollback merge attempt and return to conflict state.
- Always ensure deterministic cleanup or explicit retained-debug state.

## 11) Operational Checklist

Before execution:

- Confirm clean repository state.
- Capture source branch and head SHA.
- Provision task worktree.

Before merge:

- Ensure task tests pass.
- Ensure task-only changes.
- Ensure ancestry policy passes.

After merge:

- Run merge-target validation.
- Persist audit record.
- Remove worktree and optionally prune task branch.

## 12) Acceptance Criteria

- Every task executes in its own worktree.
- Merges always target the branch captured at task start.
- No unrelated task changes enter the merge.
- Conflicts are resolved through Copilot SDK-assisted flow with validation gates.
- Audit trail is complete for every run.

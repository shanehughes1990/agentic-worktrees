# Tracker Domain V1

## Goal

Define a brand new Project Board domain model for V1 with three Postgres tables:

- project_boards
- project_board_epics
- project_board_tasks

This model treats the board as the aggregate root and requires task-level model audit + outcome persistence.

## Domain Model

### Aggregate: ProjectBoard

The board is the root entity for planning and execution tracking per project.

Core properties:

- board_id
- project_id
- name
- board_state (pending, active, completed, failed)
- created_at
- updated_at

Rules:

- A project can have one active board at a time.
- Board deletion cascades to epics and tasks.

### Entity: ProjectBoardEpic

An epic is a board-scoped grouping of related tasks.

Core properties:

- epic_id
- board_id
- title
- objective
- epic_state (planned, in_progress, completed, blocked, failed)
- rank
- depends_on_epic_ids
- created_at
- updated_at

Rules:

- Every epic belongs to exactly one board.
- Epic rank is unique within a board for deterministic ordering.
- Epic dependencies can only reference epics in the same board.

### Entity: ProjectBoardTask

A task is an executable work unit under an epic and stores AI/model audit + outcome.

Core properties:

- task_id
- board_id
- epic_id
- title
- task_type
- task_state (planned, in_progress, completed, failed, no_work_needed)
- rank
- depends_on_task_ids
- created_at
- updated_at

Rules:

- Every task belongs to exactly one epic and one board.
- Task board_id must equal the board_id of its epic.
- Task dependencies can only reference tasks in the same board.

## Postgres Schema Contract

### project_boards

Columns:

- id uuid primary key
- project_id uuid not null
- name text not null
- state text not null
- created_at timestamptz not null default now()
- updated_at timestamptz not null default now()

Constraints / indexes:

- unique (project_id, state) where state = 'active'
- index idx_project_boards_project_id (project_id)
- index idx_project_boards_updated_at (updated_at desc)

### project_board_epics

Columns:

- id uuid primary key
- board_id uuid not null references project_boards(id) on delete cascade
- title text not null
- objective text
- state text not null
- rank int not null
- depends_on_epic_ids uuid[] not null default '{}'
- created_at timestamptz not null default now()
- updated_at timestamptz not null default now()

Constraints / indexes:

- unique (board_id, rank)
- index idx_project_board_epics_board_id (board_id)
- index idx_project_board_epics_state (state)
- index idx_project_board_epics_depends_on_epic_ids (depends_on_epic_ids) using gin

### project_board_tasks

Columns:

- id uuid primary key
- board_id uuid not null references project_boards(id) on delete cascade
- epic_id uuid not null references project_board_epics(id) on delete cascade
- title text not null
- task_type text not null
- state text not null
- rank int not null
- depends_on_task_ids uuid[] not null default '{}'
- claimed_by_agent_id text
- claimed_at timestamptz
- claim_expires_at timestamptz
- claim_token uuid
- attempt_count int not null default 0

Model audit columns:

- model_provider text not null
- model_name text not null
- model_version text
- model_run_id uuid
- agent_session_id text
- agent_stream_id text
- prompt_fingerprint text
- input_tokens int
- output_tokens int
- started_at timestamptz
- completed_at timestamptz

Outcome columns:

- outcome_status text not null
- outcome_summary text not null
- outcome_error_code text
- outcome_error_message text

Lifecycle columns:

- created_at timestamptz not null default now()
- updated_at timestamptz not null default now()

Constraints / indexes:

- unique (epic_id, rank)
- index idx_project_board_tasks_board_id (board_id)
- index idx_project_board_tasks_epic_id (epic_id)
- index idx_project_board_tasks_state (state)
- index idx_project_board_tasks_outcome_status (outcome_status)
- index idx_project_board_tasks_model_run_id (model_run_id)
- index idx_project_board_tasks_depends_on_task_ids (depends_on_task_ids) using gin
- index idx_project_board_tasks_ready_scan (board_id, state, epic_id, rank)
- index idx_project_board_tasks_claim_expiry (claim_expires_at)

## Mandatory Invariants

- Task must reference a valid board and epic.
- Task.board_id must match Epic.board_id.
- model_provider + model_name are required for any model-produced task result.
- outcome_status is always required.
- If outcome_status = 'failed', outcome_error_code is required.
- A task can be claimed by at most one active worker lease at a time.
- Dependency-gated tasks are not claimable until all prerequisite tasks are in terminal-unblocking state.

## Atomic NextTask Scheduling (DAG + Fan-Out/Fan-In)

### Why

- Many concurrent agents must request work without claiming the same task.
- Task dispatch must respect task and epic dependency DAG constraints.
- Selection order must be deterministic: epic rank first, then task rank.

### NextTask contract

- Input: `board_id`, `agent_id`, `lease_ttl`
- Output: one claimed task or no task available
- Atomicity: select + claim happen in one transaction

### Eligibility rules

A task is eligible for claim only if all are true:

- `task.state = 'planned'`
- task is not currently claimed, or claim lease expired (`claim_expires_at <= now()`)
- all `depends_on_task_ids` tasks are terminal-unblocking (`completed` or `no_work_needed`)
- task's epic has no unresolved epic dependency lock

Epic dependency lock rule:

- An epic is runnable only when all epics in `depends_on_epic_ids` are `completed`.

### Deterministic ordering

- Order by `project_board_epics.rank ASC`, then `project_board_tasks.rank ASC`

### Atomic claim pattern (Postgres)

Use a single transaction with row locking:

1. Find first eligible task by rank order.
2. Lock candidate row with `FOR UPDATE SKIP LOCKED`.
3. Update claim fields (`claimed_by_agent_id`, `claimed_at`, `claim_expires_at`, `claim_token`, `state='in_progress'`, `attempt_count=attempt_count+1`).
4. Commit and return claimed task.

This guarantees agent1 gets task1, agent2 gets task2, etc. without duplicate claims.

### Fan-out behavior

- Every agent calls `NextTask` asynchronously.
- If tasks are available, calls distribute work across many agents concurrently.

### Fan-in behavior

- If no task is currently eligible, return `no_task_available` (not error).
- Agents backoff/jitter and retry, or wake via event/notification when dependency state changes.
- Completion of prerequisite tasks naturally unlocks dependent tasks for future `NextTask` calls.

### Lease and recovery rules

- If an agent crashes, expired lease makes task reclaimable.
- Optional heartbeat may extend `claim_expires_at` for long-running tasks.
- A completion/failure write must validate `claim_token` so only the current holder can finalize.

### Ingestion retry continuity (same asynq task invocation)

- Taskboard generation/validation retries occur inside one asynq task invocation (no handoff to a new worker task for each validation retry).
- Retry attempts must reuse and carry forward available run continuity metadata (`agent_session_id`, `agent_stream_id`).
- Persist the latest known continuity metadata on the task model audit so resume/follow-up calls can continue the same agent stream/session when supported.

## State Transition Guidance

- board_state: pending -> active -> completed | failed
- epic_state: planned -> in_progress -> completed | blocked | failed
- task_state: planned -> in_progress -> completed | failed | no_work_needed
- outcome_status: success | partial | failed

## Implementation Notes

- Keep orchestration in application layer.
- Keep invariants and transitions in domain layer.
- Keep Postgres-specific details in infrastructure repositories.
- Keep API/worker correlation identifiers persisted via model_run_id and upstream run/task/job IDs.

## Scope Guardrails (Explicit)

- Ignore redesign of already-implemented agent task execution internals.
- Existing agent task execution code should only be touched enough to compile against this new contract.
- Primary focus is wiring, the new tracker domain contract, and the atomic fan-in/fan-out `NextTask` scheduler.
- Do not introduce additional Postgres tables for this tracker rewrite unless absolutely necessary.
- Prefer extending the three existing tracker tables and their columns/indexes over adding new tables.

## Go Domain Struct Model

```go
package domain

import (
 "time"

 "github.com/google/uuid"
)

type BoardState string

const (
 BoardStatePending   BoardState = "pending"
 BoardStateActive    BoardState = "active"
 BoardStateCompleted BoardState = "completed"
 BoardStateFailed    BoardState = "failed"
)

type EpicState string

const (
 EpicStatePlanned    EpicState = "planned"
 EpicStateInProgress EpicState = "in_progress"
 EpicStateCompleted  EpicState = "completed"
 EpicStateBlocked    EpicState = "blocked"
 EpicStateFailed     EpicState = "failed"
)

type TaskState string

const (
 TaskStatePlanned    TaskState = "planned"
 TaskStateInProgress TaskState = "in_progress"
 TaskStateCompleted  TaskState = "completed"
 TaskStateFailed     TaskState = "failed"
 TaskStateNoWorkNeeded TaskState = "no_work_needed"
)

type OutcomeStatus string

const (
 OutcomeStatusSuccess OutcomeStatus = "success"
 OutcomeStatusPartial OutcomeStatus = "partial"
 OutcomeStatusFailed  OutcomeStatus = "failed"
)

type ProjectBoard struct {
 ID        uuid.UUID
 ProjectID uuid.UUID
 Name      string
 State     BoardState
 CreatedAt time.Time
 UpdatedAt time.Time
}

type ProjectBoardEpic struct {
 ID        uuid.UUID
 BoardID   uuid.UUID
 Title     string
 Objective *string
 State     EpicState
 Rank      int
 DependsOnEpicIDs []uuid.UUID
 CreatedAt time.Time
 UpdatedAt time.Time
}

type TaskModelAudit struct {
 ModelProvider     string
 ModelName         string
 ModelVersion      *string
 ModelRunID        *uuid.UUID
	AgentSessionID    *string
	AgentStreamID     *string
 PromptFingerprint *string
 InputTokens       *int
 OutputTokens      *int
 StartedAt         *time.Time
 CompletedAt       *time.Time
}

type TaskOutcome struct {
 Status       OutcomeStatus
 Summary      string
 ErrorCode    *string
 ErrorMessage *string
}

type ProjectBoardTask struct {
 ID        uuid.UUID
 BoardID   uuid.UUID
 EpicID    uuid.UUID
 Title     string
 TaskType  string
 State     TaskState
 Rank      int
 DependsOnTaskIDs []uuid.UUID
 Audit     TaskModelAudit
 Outcome   TaskOutcome
 CreatedAt time.Time
 UpdatedAt time.Time
}
```

### Validation expectations for the Go model

- `ProjectBoardTask.BoardID` must match the owning epic `BoardID`.
- `ProjectBoardEpic.DependsOnEpicIDs` can only contain epics from the same board and cannot include itself.
- `ProjectBoardTask.DependsOnTaskIDs` can only contain tasks from the same board and cannot include itself.
- `TaskOutcome.Status` is always required.
- `TaskOutcome.Summary` is the only outcome content field (no large payload/blob field).
- `TaskModelAudit.ModelProvider` and `TaskModelAudit.ModelName` are required for model-produced results.
- `TaskModelAudit.AgentSessionID` and `TaskModelAudit.AgentStreamID` are optional but should be set when the agent runtime provides resumable session/stream identifiers.
- If `TaskOutcome.Status == OutcomeStatusFailed`, `TaskOutcome.ErrorCode` must be present.

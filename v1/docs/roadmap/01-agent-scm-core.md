# Slice 01 — SCM Core (GitHub First, Agents Second)

## Status

- Completed: **Completed**
- Reviewed At: **2026-03-01T01:25:40Z**

## Objective

Deliver SCM and agent orchestration with strict DDD boundaries, then harden reliability with durable Postgres-backed execution state.

## Delivery Order

1. **Part 01A — SCM for GitHub (Full Vertical Slice)**
2. **Part 01B — SCM for Agents**
3. **Reliability Hardening (Postgres-backed persistence)**

## Part 01A — SCM for GitHub (Full Vertical Slice)

### Scope

Implement a complete GitHub-first SCM vertical slice across API admission, taskengine dispatch, worker execution, and infrastructure adapters:

- Source-state operations
- Worktree lifecycle (`create`, `sync/update`, `cleanup`)
- Branch lifecycle
- Pull request lifecycle
- Review lifecycle
- Merge-intent/merge-readiness operations

### Task Checklist

- [x] Define `internal/domain/scm` contracts and invariants for source/worktree/branch/PR/review/merge-intent.
- [x] Implement `internal/application/scm` use-cases for full SCM management flows.
- [x] Add SCM API admission contract (GraphQL/control-plane).
- [x] Add SCM job kind/policy wiring in taskengine for queue, idempotency, timeout, and retry defaults.
- [x] Implement worker SCM handlers that execute SCM flows through application services.
- [x] Implement worker-only SCM authentication contract and runtime execution path.
- [x] Add first concrete `internal/infrastructure/scm` GitHub adapter.
- [x] Add typed failure classification (`transient` vs `terminal`) and retry-safe mapping.
- [x] Add checkpoint/resume boundaries for long-running SCM worker operations.
- [x] Add integration tests for API admission -> queue -> worker -> SCM adapter path.

## Part 01B — SCM for Agents

### Scope

Implement agent-facing SCM orchestration that consumes Part 01A SCM contracts through application ports only.

### Task Checklist

- [x] Define `internal/domain/agent` contracts referencing SCM capabilities via ports.
- [x] Implement `internal/application/agent` orchestration for SCM-driven flows.
- [x] Align correlation and idempotency contracts across agent + SCM operations.
- [x] Add checkpoint/resume boundaries for long-running agent+SCM orchestration.
- [x] Add worker handler integration points for agent tasks invoking SCM application services.
- [x] Add integration tests for agent orchestration through SCM ports.

## Reliability Hardening — Persistence Alignment

### Scope

Make execution reliability durable by persisting state transitions and resumability signals in Postgres.

### Task Checklist

- [x] Persist retry checkpoints in Postgres (`job_checkpoints`) and load by `idempotency_key` at job start.
- [x] Persist execution state in Postgres (`job_execution_events`) for deterministic replay/audit.
- [x] Persist workflow admission writes (`queued`) for `workflow_runs/workflow_tasks/workflow_jobs`.
- [x] Persist dead-letter triage requeue audit events (`dead_letter_events`).
- [x] Replace in-memory SCM lease manager with Postgres lease manager (`scm_repo_leases`).

## AO-Informed Runtime Patterns (Adopted)

The `.docs/agent-orchestrator` review reinforced the following patterns, which are now part of Slice 01 operating contracts:

- [x] **Correlation-first tracing**: propagate `run_id/task_id/job_id` across admission, worker execution, and persistence surfaces.
- [x] **Checkpoint-guarded idempotent handlers**: load checkpoint, short-circuit completed steps, then persist next checkpoint boundary.
- [x] **Typed failure mapping**: classify failures as `transient` vs `terminal` for deterministic retry/escalation behavior.
- [x] **Append-only operational audit trail**: persist admission and dead-letter events; persist execution lifecycle records for replay/debug.
- [x] **Worker-owned provider execution**: keep SCM auth/provider calls in worker runtime only; interface/application remain orchestration-only.

## Deliverables

- GitHub-first SCM contracts and adapters for source/worktree/branch/PR/review/merge-intent.
- Agent orchestration contracts that compose SCM through ports.
- Durable reliability foundation in Postgres for retries, execution status, admission, dead-letter auditing, and SCM lease coordination.
- AO-aligned runtime execution patterns codified as Slice 01 baseline behavior.

## In Scope

- Part 01A full GitHub-first SCM vertical slice.
- Part 01B agent orchestration consuming SCM application services.
- Postgres-backed reliability persistence needed by worker execution and operations.
- AO-derived reliability/runtime patterns listed above.

## Out of Scope

- Multi-provider parity beyond GitHub baseline.
- UI/client product flows beyond required control-plane admission.
- Non-worker authentication execution paths.

## Acceptance Criteria

- Worker is the only runtime executing SCM auth/provider operations.
- API layer validates/maps/admits and does not execute provider logic.
- Agent workflows consume SCM through application/domain contracts only.
- Retry behavior resumes from persisted Postgres checkpoint state.
- Execution/admission/triage/lease state survives worker restarts and Redis flushes.
- Correlation IDs and typed failures are preserved end-to-end for deterministic replay and auditability.

## Dependencies

- Slice 00 complete.
- Existing V1 observability and bootstrap scaffolding available.

## Exit Check

Slice 01 is complete when SCM + agent flows are contract-driven, reliability-critical operational state is durably persisted in Postgres, and AO-informed runtime patterns are enforced by default.

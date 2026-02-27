# Slice 01 — Agent and SCM Core (Two-Part Delivery)

## Objective

Deliver Slice 01 in two explicit parts:

1. **Part 01A — Core Contracts and Foundations**
2. **Part 01B — GitHub-First Vertical Slice (API + Worker Execution + Worker-Only SCM Auth)**

This keeps DDD boundaries strict while enabling immediate end-to-end SCM execution.

## Part 01A — Core Contracts and Foundations

### Scope

Define native V1 `agent` and `scm` contracts and orchestration boundaries, plus typed failures and correlation/idempotency primitives.

### Task Checklist

- [ ] Define `internal/domain/agent` contracts and invariants.
- [ ] Define `internal/domain/scm` contracts and invariants.
- [ ] Implement `internal/application/agent` orchestration ports/use-cases.
- [ ] Implement `internal/application/scm` orchestration ports/use-cases.
- [ ] Add typed failure taxonomy (`transient` vs `terminal`) for agent/scm operations.
- [ ] Define shared correlation/idempotency contract used by agent + scm workflows.
- [ ] Add integration tests through application ports (no direct SDK/provider calls in application).

### Deliverables

- `agent` contract for execution requests, status, and session introspection.
- `scm` contract for source/branch/PR/review operations.
- Correlation and idempotency contract shared across long-running flows.
- Initial adapter seams ready for provider wiring.

### Acceptance Criteria

- Application flows call only slot contracts, not concrete providers.
- Failure typing is explicit and consumable by retry policy.
- Contracts are sufficient to support worker execution-plane integration.
- DDD dependency direction is preserved (`interface -> application -> domain`).

## Part 01B — GitHub-First Vertical Slice

### Scope

Implement a full end-to-end GitHub-first SCM slice:

- **Configured from API/control-plane entry points**
- **Executed only on worker handlers**
- **SCM authentication performed only on worker**

This includes PR management operations required by initial workflow logic.

### Task Checklist

- [ ] Add API-facing SCM control surface to enqueue SCM workflows (GraphQL/control-plane admission).
- [ ] Add SCM job kind/policy wiring for queue dispatch and idempotent scheduling.
- [ ] Implement worker SCM handler(s) that execute SCM workflows through application services.
- [ ] Implement GitHub-first SCM infrastructure adapter for source/branch/PR/review operations.
- [ ] Implement worker-side SCM auth flow (credential validation/bootstrap/usage) with no API-side auth execution.
- [ ] Propagate and emit `run_id`, `task_id`, `job_id` for all SCM jobs.
- [ ] Add checkpoint/resume boundaries for long-running SCM execution.
- [ ] Add integration tests covering API enqueue -> worker execution -> SCM adapter path.

### Deliverables

- API-admitted SCM workflow requests mapped to application use-cases.
- Worker-registered SCM handlers executing deterministic SCM tasks.
- GitHub-first adapter for branch/PR/review primitives.
- Worker-owned SCM auth contract + implementation.
- End-to-end observability and retry-safe execution semantics.

### Acceptance Criteria

- API layer validates/maps/enqueues only; it does not execute SCM auth or provider SDK logic.
- Worker handlers own SCM execution and authentication lifecycle.
- SCM workflows execute through ports/adapters under DDD boundaries.
- Retry behavior is safe via idempotency, checkpoints, and typed failures.
- Vertical slice supports practical PR-management workflow for GitHub baseline.

## In Scope (Slice 01)

- Part 01A and Part 01B as defined above.
- GitHub-first provider baseline only.
- API configuration/admission path + worker execution path for SCM workflows.

## Out of Scope (Slice 01)

- Multi-provider parity beyond GitHub baseline.
- UI/client experience beyond required API control-plane admission.
- Non-worker SCM authentication execution paths.

## Dependencies

- Slice 00 complete.
- Existing V1 observability and bootstrap scaffolding available.

## Exit Check

Slice 01 is complete when:

1. Core `agent` + `scm` contracts are implemented and application-orchestrated.
2. A GitHub-first SCM workflow can be admitted via API, executed/authenticated on worker, and run end-to-end under typed failures, idempotency, and checkpoint semantics.

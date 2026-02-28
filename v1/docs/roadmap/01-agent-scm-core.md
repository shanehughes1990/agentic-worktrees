# Slice 01 — SCM Core (GitHub First, Agents Second)

## Objective

Deliver Slice 01 in two SCM-specific parts, in this order:

1. **Part 01A — SCM for GitHub (Full Vertical Slice)**
2. **Part 01B — SCM for Agents**

This ordering ensures complete SCM management capability exists first (admission, execution, auth, provider integration), then agent execution composes it through strict DDD ports.

## Part 01A — SCM for GitHub (Full Vertical Slice)

### Part 01A Scope

Implement a complete GitHub-first SCM vertical slice across API, queue, worker, and infrastructure:

- Source-state operations
- Worktree lifecycle operations (`create`, `sync/update`, `cleanup`)
- Branch lifecycle operations
- Pull request lifecycle operations
- Review operations
- Merge-intent/merge-readiness operations

This includes API control-plane admission, taskengine dispatch policy, worker execution, worker-only SCM auth execution, and GitHub adapter behavior under typed failure semantics.

### Part 01A Task Checklist

- [x] Define `internal/domain/scm` contracts and invariants for source/worktree/branch/PR/review/merge-intent.
- [x] Implement `internal/application/scm` use-cases for full SCM management flows.
- [x] Add SCM API admission contract (GraphQL/control-plane) for required SCM management operations.
- [x] Add SCM job kind/policy wiring in taskengine for queue, idempotency, timeout, and retry defaults.
- [x] Implement worker SCM handler(s) that execute SCM flows through application services.
- [x] Implement worker-only SCM authentication contract and runtime execution path.
- [x] Add first concrete `internal/infrastructure/scm` GitHub adapter for full Part 01A scope.
- [x] Add typed failure classification (`transient` vs `terminal`) and retry-safe mapping.
- [ ] Add checkpoint/resume boundaries for long-running SCM worker operations.
- [ ] Add integration tests for API admission -> queue -> worker -> SCM adapter path.

### Part 01A Deliverables

- GitHub-first SCM contracts covering `source`, `worktree`, `branch`, `pull request`, `review`, and merge-intent checks.
- SCM application orchestration services using ports only.
- SCM API admission + worker execution path wired end-to-end.
- Worker-owned SCM authentication implementation for provider access.
- Initial GitHub infrastructure adapter implementation.
- Retry-safe typed failure semantics and correlation/idempotency propagation.
- Integration coverage proving full vertical slice behavior.

### Part 01A Acceptance Criteria

- SCM use-cases call only SCM ports; no provider calls from application layer.
- API layer validates/maps/admits; it does not execute SCM auth/provider logic.
- Worker is the only runtime that executes SCM auth logic and provider operations.
- GitHub baseline source/worktree/branch/PR/review/merge-intent primitives are functional through contracts.
- End-to-end SCM flow executes through API -> taskengine -> worker -> SCM adapter with correlation IDs and idempotency.
- DDD boundaries remain strict (`interface -> application -> domain`; infrastructure implements ports).

## Part 01B — SCM for Agents

### Part 01B Scope

Implement agent-facing SCM orchestration that consumes Part 01A SCM contracts for execution workflows.

### Part 01B Task Checklist

- [x] Define `internal/domain/agent` contracts that reference SCM capabilities through ports.
- [x] Implement `internal/application/agent` orchestration for SCM-driven execution/session flows.
- [ ] Add correlation and idempotency contract alignment between agent workflows and SCM operations.
- [ ] Add checkpoint/resume boundaries for long-running agent+SCM orchestration.
- [ ] Add worker handler integration points for agent tasks that invoke SCM application services.
- [ ] Add integration tests for agent orchestration through SCM ports (without bypassing application layer).

### Part 01B Deliverables

- Agent contract for SCM-aware execution and session introspection.
- Agent application orchestration that composes SCM services via ports.
- Shared execution semantics (correlation IDs, idempotency, typed failures) across agent+SCM.
- Test coverage for agent-to-SCM orchestration path.

### Part 01B Acceptance Criteria

- Agent layer never calls SCM provider adapters directly.
- SCM interactions in agent workflows occur only through application/domain contracts.
- Retry, checkpoint, and correlation behavior is deterministic and test-covered.
- No Copilot/SCM SDK execution occurs outside worker handler paths.

## In Scope (Slice 01)

- Part 01A full GitHub-first SCM vertical slice.
- Part 01B agent orchestration that consumes Part 01A SCM capabilities.
- API admission/configuration and worker execution paths needed to support both parts.

## Out of Scope (Slice 01)

- Multi-provider parity beyond GitHub baseline.
- UI/client product flows beyond required control-plane admission.
- Non-worker authentication execution paths.

## Dependencies

- Slice 00 complete.
- Existing V1 observability and bootstrap scaffolding available.

## Exit Check

Slice 01 is complete when:

1. GitHub-first SCM primitives (`source`, `worktree`, `branch`, `PR`, `review`, merge-intent) are contract-defined, application-orchestrated, API-admitted, worker-executed, and infrastructure-backed.
2. Agent execution flows consume SCM through ports with typed failures, correlation/idempotency, and checkpoint/resume semantics under strict DDD boundaries.

# Slice 01 — SCM Core (GitHub First, Agents Second)

## Objective

Deliver Slice 01 in two SCM-specific parts, in this order:

1. **Part 01A — SCM for GitHub**
2. **Part 01B — SCM for Agents**

This ordering ensures provider-safe SCM primitives exist first, then agent execution can consume them through strict DDD ports.

## Part 01A — SCM for GitHub

### Scope

Implement the GitHub-first SCM foundation and vertical path for source, branch, PR, and review operations.

### Task Checklist

- [ ] Define `internal/domain/scm` GitHub-first contracts and invariants.
- [ ] Implement `internal/application/scm` use-cases for source-state, branch, PR, and review flows.
- [ ] Add first concrete `internal/infrastructure/scm` GitHub adapter seam.
- [ ] Add SCM job policy/wiring needed for API admission and worker execution readiness.
- [ ] Enforce worker-only SCM authentication boundaries (no API-side SCM auth execution).
- [ ] Add typed failure classification (`transient` vs `terminal`) for SCM operations.
- [ ] Add integration tests through SCM application ports/adapters.

### Deliverables

- GitHub-first SCM contracts (`source`, `branch`, `pull request`, `review`).
- SCM orchestration service in application layer using ports only.
- Initial GitHub adapter implementation/seam in infrastructure layer.
- Retry-safe failure semantics and idempotency/correlation alignment.

### Acceptance Criteria

- SCM use-cases call only SCM ports; no provider calls from application layer.
- Worker is the only runtime that executes SCM auth logic.
- GitHub baseline branch/PR/review primitives are functional through contracts.
- DDD boundaries remain strict (`interface -> application -> domain`; infrastructure implements ports).

## Part 01B — SCM for Agents

### Scope

Implement agent-facing SCM orchestration that consumes Part 01A SCM contracts for execution workflows.

### Task Checklist

- [ ] Define `internal/domain/agent` contracts that reference SCM capabilities through ports.
- [ ] Implement `internal/application/agent` orchestration for SCM-driven execution/session flows.
- [ ] Add correlation and idempotency contract alignment between agent workflows and SCM operations.
- [ ] Add checkpoint/resume boundaries for long-running agent+SCM orchestration.
- [ ] Add worker handler integration points for agent tasks that invoke SCM application services.
- [ ] Add integration tests for agent orchestration through SCM ports (without bypassing application layer).

### Deliverables

- Agent contract for SCM-aware execution and session introspection.
- Agent application orchestration that composes SCM services via ports.
- Shared execution semantics (correlation IDs, idempotency, typed failures) across agent+SCM.
- Test coverage for agent-to-SCM orchestration path.

### Acceptance Criteria

- Agent layer never calls SCM provider adapters directly.
- SCM interactions in agent workflows occur only through application/domain contracts.
- Retry, checkpoint, and correlation behavior is deterministic and test-covered.
- No Copilot/SCM SDK execution occurs outside worker handler paths.

## In Scope (Slice 01)

- Part 01A GitHub-first SCM foundations and provider baseline.
- Part 01B agent orchestration that consumes SCM contracts.
- API admission/configuration and worker execution readiness needed to support these two parts.

## Out of Scope (Slice 01)

- Multi-provider parity beyond GitHub baseline.
- UI/client product flows beyond required control-plane admission.
- Non-worker authentication execution paths.

## Dependencies

- Slice 00 complete.
- Existing V1 observability and bootstrap scaffolding available.

## Exit Check

Slice 01 is complete when:

1. GitHub-first SCM primitives (source/branch/PR/review) are contract-defined, application-orchestrated, and infrastructure-backed.
2. Agent execution flows consume SCM through ports with typed failures, correlation/idempotency, and checkpoint/resume semantics under strict DDD boundaries.

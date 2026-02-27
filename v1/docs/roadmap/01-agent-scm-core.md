# Slice 01 — Agent and SCM Core Contracts

## Objective

Define and implement `agent` and `scm` slot contracts with initial provider adapters under strict DDD boundaries.

## Task Checklist

- [ ] Define `internal/domain/agent` core contracts and invariants.
- [ ] Define `internal/domain/scm` core contracts and invariants.
- [ ] Implement `internal/application/agent` orchestration ports/use-cases.
- [ ] Implement `internal/application/scm` orchestration ports/use-cases.
- [ ] Create first concrete `infrastructure/agent` adapter.
- [ ] Create first concrete `infrastructure/scm` adapter.
- [ ] Add typed error taxonomy (`transient` vs `terminal`) for agent/scm operations.
- [ ] Add integration tests that execute through ports (not concrete SDK calls from application).

## Deliverables

- `agent` contract for execution requests, status, and session introspection.
- `scm` contract for branch/worktree/PR/review operations.
- Shared correlation and idempotency contract across agent + scm workflows.
- Initial provider adapters wired through application ports.

## In Scope

- Slot interfaces and application orchestration boundaries.
- SCM primitives needed by local and remote worker bootstrap.
- Error typing (`transient` vs `terminal`) for retry policies.

## Out of Scope

- Multi-provider breadth beyond initial adapter baseline.
- UI concerns.

## Acceptance Criteria

- Application flows call only slot contracts, not concrete providers.
- SCM operations support source-state bootstrap requirements.
- Agent contract supports resumable session-level orchestration.
- Contracts are sufficient for execution-plane integration tests.

## Dependencies

- Slice 00 complete.

## Exit Check

This slice is complete when core execution can run through `agent` and `scm` ports without direct provider coupling.

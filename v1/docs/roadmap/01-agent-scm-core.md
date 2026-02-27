# Slice 01 — Agent and SCM Core Contracts

## Objective

Define and implement `agent` and `scm` slot contracts with initial provider adapters under strict DDD boundaries.

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

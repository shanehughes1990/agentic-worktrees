# Slice 06 — GraphQL Control Plane

## Objective

Deliver a GraphQL-first control plane (`gqlgen`) as the primary contract for orchestrator operations.

## Task Checklist

- [x] Initialize `gqlgen` project structure and schema/resolver scaffolding.
- [x] Place GraphQL files under DDD-aligned interface paths.
- [ ] Implement session/workflow/worker queries against application services.
- [ ] Implement core control mutations (spawn/send/cancel/restore/assign/merge).
- [ ] Implement subscriptions for session + orchestrator events.
- [ ] Add contract/integration tests for GraphQL operations.

## Deliverables

- Schema for sessions, workers, workflows, and events.
- Mutations for core operator actions (spawn/send/cancel/restore/assign/merge).
- Queries for system state and fleet visibility.
- Subscriptions for realtime session and orchestrator updates.

## In Scope

- Contract-first schema evolution with resolver boundaries.
- Mapping between application use cases and GraphQL operations.
- Error and status models suitable for operator clients.

## Out of Scope

- Non-GraphQL control-plane alternatives.
- Client-only UX decisions.

## Acceptance Criteria

- Core control-plane actions are available via GraphQL.
- Read models reflect orchestrator, worker, and tracker state coherently.
- Subscriptions deliver production-usable update streams.

## Dependencies

- Slices 01, 03, 04, and 05.

## Exit Check

This slice is complete when operator workflows can be executed entirely through GraphQL contracts.

# Slice 06 — GraphQL Control Plane

## Objective

Deliver a GraphQL-first control plane (`gqlgen`) as the primary contract for orchestrator operations, backed by Postgres read models for durable queryability.

## Control-Plane Mandate

- API runtime REST surface is restricted to:
  - GraphQL playground endpoint.
  - GraphQL handler endpoint.
  - Health endpoints (liveness/readiness).
- All client-facing communication must be GraphQL, with type-safe contracts mapped to internal typed application/domain contracts.
- GraphQL contracts must remove guesswork by using explicit required/optional fields and typed payloads for both inputs and outputs.
- Forward-facing errors must be typed and returned as explicit union outputs (or equivalent typed schema patterns) so clients can deterministically handle failures.
- Additional REST endpoints are prohibited for control-plane features.
- Exception: a REST endpoint may be added only when strictly required for third-party integration ingress/configuration (for example webhook ingestion), and must be justified by that requirement.

## Task Checklist

- [x] Initialize `gqlgen` project structure and schema/resolver scaffolding.
- [x] Place GraphQL files under DDD-aligned interface paths.
- [ ] Implement session/workflow/worker queries against application query services.
- [ ] Implement core control mutations (spawn/send/cancel/restore/assign/merge).
- [ ] Implement subscriptions for session + orchestrator events.
- [ ] Add Postgres-backed query repositories for run/task/job/worker and dead-letter history views.
- [ ] Add contract/integration tests for GraphQL operations.

## Deliverables

- Schema for sessions, workers, workflows, execution history, and events.
- Mutations for core operator actions.
- Queries backed by Postgres read models (not transient runtime memory).
- Subscriptions for realtime updates with replay alignment from persisted stream data.

## In Scope

- Contract-first schema evolution with resolver boundaries.
- Mapping between application use cases and GraphQL operations.
- Postgres-backed query repositories and pagination/filtering contracts.

## Out of Scope

- Non-GraphQL control-plane alternatives.
- Client-only UX decisions.

## Acceptance Criteria

- Core control-plane actions are available via GraphQL.
- Read models reflect orchestrator, worker, and tracker state coherently from Postgres.
- Subscriptions deliver production-usable update streams aligned with persisted event replay.

## Dependencies

- Slices 01, 03, 04, and 05.

## Exit Check

This slice is complete when operator workflows can be executed entirely through GraphQL contracts with durable persisted state behind queries/subscriptions.

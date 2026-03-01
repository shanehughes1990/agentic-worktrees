# Slice 05 — Realtime Streams

## Objective

Provide reliable real-time event streams for session activity, agent output, and orchestrator decisions, with Postgres-backed persistence for replay and diagnostics.

## Task Checklist

- [ ] Define stream event schemas for session, agent output, workflow execution, and supervisor decisions.
- [ ] Implement stream publication pipeline with correlation IDs.
- [ ] Persist stream events in Postgres (`stream_events`) with replay cursor semantics.
- [ ] Implement reconnect replay protocol from persisted offsets.
- [ ] Implement ordering guarantees and backpressure handling strategy.
- [ ] Add integration tests for publish/subscribe + replay behavior.

## Deliverables

- Session activity stream contract.
- Agent output/log stream contract.
- Orchestrator/supervisor decision stream contract.
- Postgres `stream_events` persistence and replay model.

## In Scope

- Event schemas with timestamps and correlation IDs.
- Ordering, durability, and replay semantics for operational visibility.
- Integration points for GraphQL subscriptions.

## Out of Scope

- Final client UI presentation details.
- Non-essential stream channels not tied to V1 operations.

## Acceptance Criteria

- Streams are consumable in near-real-time.
- Operators can trace end-to-end flow from correlation IDs.
- Reconnect/replay paths are served from persisted Postgres stream data.

## Dependencies

- Slices 01 and 04.
- Slice 03 supervisor event taxonomy.
- Slice 06 subscription layer.

## Exit Check

This slice is complete when operational decisions and agent activity are inspectable live and replayable from Postgres.

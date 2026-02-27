# Slice 05 — Realtime Streams

## Objective

Provide reliable real-time event streams for session activity, agent output, and orchestrator decisions.

## Deliverables

- Session activity stream contract.
- Agent output/log stream contract.
- Orchestrator decision stream contract.
- Stream persistence/replay model for diagnostics and client reconnects.

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
- Reconnect/replay paths work for active sessions.

## Dependencies

- Slices 01 and 04.
- Slice 06 subscription layer.

## Exit Check

This slice is complete when operational decisions and agent activity are inspectable live and after-the-fact.

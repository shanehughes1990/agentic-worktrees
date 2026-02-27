# Slice 06 — GraphQL Control Plane

## Objective

Implement a GraphQL-first control plane using `gqlgen` as the primary contract.

## Scope

- Initial schema for sessions, workers, workflows, and events.
- Core mutations: spawn/send/cancel/restore/assign/merge.
- Queries for operational state and worker fleet visibility.
- Subscriptions for session and orchestrator events.

## Acceptance Criteria

- All control actions are available through GraphQL contracts.
- Read models expose orchestrator, worker, and tracker state coherently.
- Subscription pathways deliver realtime updates to clients.

## Dependencies

- Orchestrator supervisor.
- Worker execution plane.
- Realtime event streams.

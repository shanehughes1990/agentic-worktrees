# Slice 01 — Orchestrator Supervisor

## Objective

Implement the central orchestration brain that owns policy decisions, global session monitoring, and escalation signaling.

## Scope

- Supervisor state model for session lifecycle and attention zones.
- Decision policy engine for merge/respond/review/working/done transitions.
- Event taxonomy for orchestrator decisions and state transitions.
- Correlation IDs through all orchestrator actions.

## Acceptance Criteria

- Orchestrator emits deterministic state transitions for tracked sessions.
- Decision events are streamable and queryable by correlation ID.
- Policy decisions are decoupled from concrete provider implementations.

## Dependencies

- `00-project-layout.md` completed.
- Worker execution contracts.
- Realtime stream transport.
- GraphQL subscriptions and query models.

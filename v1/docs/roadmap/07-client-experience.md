# Slice 07 — Cross-Platform Client Experience

## Objective

Ship a cross-platform client that operates the system through GraphQL and presents durable operational state sourced from Postgres-backed control-plane models.

## Task Checklist

- [ ] Implement runtime-configurable GraphQL endpoint setup flow.
- [ ] Implement multi-session status board and attention zones.
- [ ] Implement session detail stream view with replay-from-offset support.
- [ ] Implement control actions wired to GraphQL mutations.
- [ ] Add reconnect/resilience behavior for stream disruptions.
- [ ] Surface historical execution/dead-letter/supervisor history from persisted data.
- [ ] Add cross-platform packaging + smoke test coverage.

## Deliverables

- Runtime-configurable backend endpoint management.
- Multi-session status board with attention zones.
- Session detail view with live stream output + persisted history replay.
- Operator-oriented interaction model for escalation and intervention.

## In Scope

- Client flows required for day-1 V1 operation.
- Integration with GraphQL queries/mutations/subscriptions.
- Resilient reconnect behavior using persisted replay data.

## Out of Scope

- Advanced visual polish not needed for operational readiness.
- Terminal-based fallback UX.

## Acceptance Criteria

- Client runs across supported OS targets.
- Operators can monitor and control workflows from one interface.
- Session activity and decision streams are visible in near-real-time and can recover from persisted replay history.

## Dependencies

- Slice 06 control plane.
- Slice 05 stream layer.

## Exit Check

This slice is complete when core orchestration operations can be run from the client alone with resilient replayable operational visibility.

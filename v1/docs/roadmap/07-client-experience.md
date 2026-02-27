# Slice 07 — Cross-Platform Client Experience

## Objective

Ship a cross-platform client that operates the system through GraphQL without terminal-first UX.

## Task Checklist

- [ ] Implement runtime-configurable GraphQL endpoint setup flow.
- [ ] Implement multi-session status board and attention zones.
- [ ] Implement session detail stream view.
- [ ] Implement control actions wired to GraphQL mutations.
- [ ] Add reconnect/resilience behavior for stream disruptions.
- [ ] Add cross-platform packaging + smoke test coverage.

## Deliverables

- Runtime-configurable backend endpoint management.
- Multi-session status board with attention zones.
- Session detail view with live stream output and control actions.
- Operator-oriented interaction model for escalation and intervention.

## In Scope

- Client flows required for day-1 V1 operation.
- Integration with GraphQL queries/mutations/subscriptions.
- Basic resiliency for reconnect and stream continuity.

## Out of Scope

- Advanced visual polish not needed for operational readiness.
- Terminal-based fallback UX.

## Acceptance Criteria

- Client runs across supported OS targets.
- Operators can monitor and control workflows from one interface.
- Session activity and decision streams are visible in near-real-time.

## Dependencies

- Slice 06 control plane.
- Slice 05 stream layer.

## Exit Check

This slice is complete when core orchestration operations can be run from the client alone.

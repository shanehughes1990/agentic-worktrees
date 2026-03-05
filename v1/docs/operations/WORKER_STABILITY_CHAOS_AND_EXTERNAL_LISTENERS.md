# Worker Stability Chaos and External Listener Validation

## Chaos Scenarios

Run these scenarios against the canonical lifecycle path:

1. Worker crash/restart during active run.
2. Listener outage with repeated transient failures until circuit-open.
3. Delayed event arrival causing sequence gap and reconciliation.

Expected outcomes:

- Canonical history remains append-safe and ordered.
- Snapshot converges to recovered state after restart.
- Listener outages do not block canonical event persistence.
- Gap events are emitted and visible to board/subscription consumers.

## External Listener Adapters

Phase-2 listener types are supported by the same lifecycle listener contract:

- `webhook`
- `slack`
- `bus`

All listener outcomes remain queryable via `project_session_feedback_deliveries` with the same audit model as internal listeners.

## Validation Gates

- External listener rows are persisted for lifecycle events.
- External failures update delivery status without interrupting canonical writes.
- Replay from `project_session_history` remains deterministic regardless of listener status.

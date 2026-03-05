# Worker Stability SLO Dashboard

## Purpose

Define the measurable SLO dashboard and alert contract for worker stability and lifecycle fan-out reliability.

## SLO Targets

- Event persistence latency p95 <= 1s.
- API desktop delivery latency p95 <= 2s.
- Fan-out delivery success >= 99.9% (excluding terminal listener failures).

## Metrics Sources

Use GraphQL + persisted lifecycle tables as the source of truth:

- `interventionMetrics(projectID, limit)` for intervention count and MTTR.
- `projectEvents(projectID, fromOffset, limit)` for event volume and gap visibility.
- `lifecycleSessionSnapshots(projectID, pipelineType, limit)` for active state distribution.
- `project_session_feedback_deliveries` for listener delivery status and retry lag.

## Required Dashboard Panels

1. Transition volume by state and pipeline.
2. Gap detected vs reconciled trend.
3. Listener delivery success/retry/terminal failure trend.
4. Intervention MTTR and outcome ratio.
5. Unexpected exits by project and run.

## Alert Rules

- `p95_persist_latency_seconds > 1` for 10m.
- `p95_delivery_latency_seconds > 2` for 10m.
- `fanout_success_ratio < 0.999` for 15m.
- `gap_reconciliation_ratio < 0.95` for 15m.
- `unexpected_exit_rate > baseline * 2` for 10m.

## Verification Checklist

- Alerts can be triggered in staging.
- Dashboard values reconcile with lifecycle table queries.
- Runbook links are attached to each alert condition.

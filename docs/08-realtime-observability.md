# 08 - Realtime Observability

## Goal

Provide full realtime insight into Asynq queues, running workers, agent progression, and git/PR lifecycle so operators can trust and diagnose the system quickly.

## Observability Pillars

### Logs

- Structured JSON logs only
- Mandatory fields: `timestamp`, `level`, `run_id`, `task_id`, `component`, `event_type`
- Correlation IDs across planner/enqueuer, worker, git, and agent runtime

### Metrics

Core metrics:

- queue depth by queue name
- enqueue-to-start latency
- in-progress worker count by task type
- retry rate by task type
- archive/dead-letter count and age
- queue depth by status
- run duration by phase
- conflict rate and resolution outcome rate
- retry count and dead-letter count
- merge success/failure rates
- agent heartbeat and timeout counts

### Events / Timeline

- Append-only event stream per run
- Each state transition emits event with payload
- Tail-able stream for CLI dashboards and future interface adapters
- Backend file/store changes are normalized into shared task events
- Event fan-out supports all active task threads concurrently
- Per-subscriber cursor/offset tracking supports catch-up and replay
- Asynq job lifecycle events are included (`enqueued`, `started`, `retry`, `archived`, `succeeded`)

### Tracing (Recommended)

- Span per lifecycle phase
- Child spans for git operations and runtime calls
- Error tags for transient vs terminal failures

## Realtime Views (CLI MVP, extensible later)

Minimum views:

- queue health summary (depth, retries, archived)
- active runs with phase and elapsed time
- blocked tasks with reason
- conflict queue with action required
- recent failures and retry ETA
- live board diff feed (task moves/dependency/status changes)

## Alerting (Baseline)

- no progress heartbeat for active run
- queue depth/latency breach per SLA threshold
- repeated rebase failure above threshold
- dead-letter queue growth
- merge gate failure spike

## Data Retention

- hot logs/events for recent runs
- archived artifacts for audit window
- configurable retention with cost caps

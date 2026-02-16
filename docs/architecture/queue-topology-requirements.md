# Queue Topology Requirements for ADK Workloads

## Objective

Guarantee resilient ADK execution through durable queue semantics.

## Required ADK Task Classes

- `task.adk.ingest_scope`
- `task.adk.plan_board`
- `task.adk.execute_agent`
- `task.adk.summarize_result`

## Required Queue Controls

- Max retries per task type.
- Exponential backoff with jitter.
- Timeout per task type.
- Dead-letter/archive route for terminal failures.
- Concurrency caps by queue and task type.

## Required Payload Contracts

- Deterministic payload fields.
- Idempotency key.
- Correlation metadata (`run_id`, `task_id`, `job_id`).

## Required State Transitions

- `queued`
- `claimed`
- `executing_adk`
- `persisting_result`
- `succeeded` | `retry_scheduled` | `dead_lettered`

Each transition must emit telemetry and write checkpoints.

# ADK Task Resilience Policy (Asynq Required)

## Scope

This policy applies to every ADK interaction in this repository.

## Policy Statement

All ADK operations must run inside Asynq task handlers. Interface layers must only enqueue work and return control.

## Canonical Flow

1. Request accepted in CLI/MCP/API.
2. Payload normalized and idempotency key assigned.
3. Asynq task enqueued to durable queue.
4. Worker claims task and restores checkpoint context.
5. Worker executes ADK call within handler.
6. Worker persists result/checkpoint and emits telemetry.
7. Worker ack/retry/dead-letter according to typed failure class.

## Non-Compliant Examples

- Direct ADK call from command code.
- ADK call from repository/model utility.
- ADK call with no retry policy.
- ADK call with no checkpoint writes.

## Compliance Requirements

- Task type registered in queue topology.
- Retry/backoff policy explicitly configured.
- Dead-letter path documented and tested.
- Checkpoint transitions persisted.
- Correlation IDs present in logs and audit.

## Enforcement

This is a hard requirement. Any direct ADK invocation path outside Asynq handlers must be rejected.

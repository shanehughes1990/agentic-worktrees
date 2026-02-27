# Slice 05 — Realtime Streams

## Objective

Expose live operational telemetry for sessions, agent output, and orchestrator decisions.

## Scope

- Session activity stream.
- Agent output/log stream.
- Orchestrator decision stream.
- Persistence model for stream replay and inspection.

## Acceptance Criteria

- Streams are consumable in near-real-time by API and client layers.
- Events include correlation IDs and timestamped lifecycle context.
- Operators can trace decision rationale from stream data.

## Dependencies

- Orchestrator supervisor event taxonomy.
- GraphQL subscription layer.

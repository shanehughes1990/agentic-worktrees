# Slice 05 — Realtime Streams

## Objective

Provide reliable real-time event streams for session activity, agent output, and orchestrator decisions, with Postgres-backed persistence for replay and diagnostics.

Copilot CLI is the primary realtime source for V1.

## Copilot CLI-First Constraints (Research Summary)

### Supported, stable surfaces

- `copilot --acp` exposes a supported protocol stream for session lifecycle and message chunks.
- ACP integration examples include `sessionUpdate` events with `agent_message_chunk` payloads for realtime output.
- Copilot CLI supports steering while thinking by enqueuing additional messages in-session.

### Physical files on device (supported/observable)

- Config location defaults to `~/.copilot/config.json` (overridable via `XDG_CONFIG_HOME`).
- Log directory defaults to `~/.copilot/logs/` (or `--log-dir`).
- Session state defaults to `~/.copilot/session-state/{session-id}/`, including `events.jsonl`, `workspace.yaml`, `plan.md`, and checkpoints.

### Important limitation

- VS Code Chat extension APIs do not provide a global tap into all Copilot chat traffic.
- Extension chat history is scoped to the current participant context, not a universal Copilot transcript stream.
- Therefore, V1 should not depend on scraping VS Code chat internals.

## V1 Architecture Decision

- Primary ingestion path: ACP session stream (`copilot --acp --stdio` or TCP mode).
- Secondary recovery path: session-state file ingestion from `events.jsonl` for resume/replay.
- Healthcheck signal source: ACP heartbeats/session updates + process liveness + recent event lag.
- Steering/injection path: send new prompt turns to the same ACP `sessionId` instead of terminal keystroke injection.
- Persistence target: Postgres `stream_events` with monotonic offsets and correlation IDs.

## Event Contract (Copilot-focused)

Emit normalized events (envelope + payload) from both ACP and file recovery:

- `stream.session.started`
- `stream.session.updated`
- `stream.agent.chunk`
- `stream.agent.turn_completed`
- `stream.tool.started`
- `stream.tool.completed`
- `stream.permission.requested`
- `stream.permission.decided`
- `stream.session.checkpointed`
- `stream.session.ended`
- `stream.session.recovered`
- `stream.session.health`
- `stream.session.injected_prompt`

Required envelope fields:

- `event_id`
- `stream_offset`
- `occurred_at`
- `run_id`
- `session_id`
- `task_id` (nullable)
- `correlation_id`
- `source` (`acp` | `session_file` | `worker`)
- `event_type`
- `payload` (JSONB)

## Task Checklist

- [ ] Define stream event schemas for session, agent output, workflow execution, and supervisor decisions.
- [ ] Implement ACP stream publication pipeline with correlation IDs.
- [ ] Implement session-state recovery reader for `~/.copilot/session-state/*/events.jsonl`.
- [ ] Persist stream events in Postgres (`stream_events`) with replay cursor semantics.
- [ ] Implement reconnect replay protocol from persisted offsets.
- [ ] Implement ordering guarantees and backpressure handling strategy.
- [ ] Implement healthcheck evaluator for agent process + event freshness.
- [ ] Implement in-session prompt injection command path (ACP session prompt enqueue).
- [ ] Add integration tests for publish/subscribe + replay behavior.

## Deliverables

- Session activity stream contract.
- Agent output/log stream contract.
- Orchestrator/supervisor decision stream contract.
- Postgres `stream_events` persistence and replay model.
- ACP adapter for realtime ingest + prompt injection.
- Session-state recovery adapter for replay continuity.

## In Scope

- Event schemas with timestamps and correlation IDs.
- Ordering, durability, and replay semantics for operational visibility.
- Copilot CLI ACP ingestion and prompt injection support.
- Session-state file recovery from local CLI state.
- Integration points for GraphQL subscriptions.

## Out of Scope

- Reading undocumented VS Code Copilot internal transport.
- Non-essential stream channels not tied to V1 operations.
- Final client UI presentation details.

## Acceptance Criteria

- Streams are consumable in near-real-time from ACP-originated events.
- Operators can trace end-to-end flow from correlation IDs.
- Reconnect/replay paths are served from persisted Postgres stream data.
- If ACP disconnects, recovery can continue from session-state artifacts without offset corruption.
- Operators can inject a prompt into an active session and observe resulting stream events.

## Dependencies

- Slices 01 and 04.
- Slice 03 supervisor event taxonomy.
- Slice 06 subscription layer.

## Exit Check

This slice is complete when operational decisions and Copilot CLI agent activity are inspectable live, steerable via prompt injection, and replayable from Postgres.

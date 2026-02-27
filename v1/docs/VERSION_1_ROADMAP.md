# Version 1 Roadmap

This document defines the V1 target architecture for this repository.

Reference-project analysis is kept separate in `docs/AGENT_ORCHESTRATOR_ANALYSIS.md`.

## Context

- Current production-style baseline remains in `mvp/`.
- New implementation workspace is `v1/` (`module agentic-orchestrator`).

## Ground-Up Rewrite Policy

V1 is a full rewrite.

- `mvp/` is inspiration and behavioral reference only.
- No source import, package reuse, file copy, or code migration from `mvp/` into `v1/`.
- All V1 code, contracts, schemas, and adapters must be newly authored in `v1/`.
- Similar outcomes are allowed; implementation must be original to V1.

## V1 Non-Negotiable Requirements

### A) Agnostic slot model (V1 canonical)

V1 pluggability is defined by these five slots:

- `agent`
  - AI worker implementation and session introspection.
  - Examples: Claude Code, Codex, Aider-style adapters.

- `scm`
  - Source control and PR/CI/review provider integration.
  - Examples: GitHub now, GitLab/others later.

- `tracker`
  - Issue/work-item + board/task source and lifecycle integration.
  - Examples: GitHub Issues now; local JSON board model, Jira, Linear later.

- `notifier`
  - Human escalation and delivery channel.
  - Examples: desktop, Slack, webhook.

- `client` (cross-platform GraphQL client)
  - User-facing control surface connecting to the GraphQL API at runtime-configured server addresses.
  - Examples: desktop/web cross-platform client surfaces.

Note: These slots are implemented with strict DDD boundaries in this repo; slots are contracts, while implementations live in the correct layer.

### B) Real-time agent streams

V1 should support:

- Session activity stream (state transitions, liveness, attention level).
- Agent output/log stream (tail-style, structured where possible).
- Orchestrator decision stream (why a reaction/escalation happened).

Source candidates:

- Runtime output capture.
- Agent session artifacts/log files.
- Orchestrator lifecycle events.

### C) Remote + local workers

V1 execution model:

- Local worker mode: same-host execution (like MVP simplicity).
- Remote worker mode: distributed workers over network, not tied to same filesystem, and able to bootstrap directly from SCM origin.

Required properties:

- SCM-backed task context instead of local-path-only assumptions.
- Remote worker self-bootstrap flow:
  - authenticate with SCM credentials
  - pull/clone from `origin`
  - checkout/pull the source branch
  - start execution pipeline from that fetched source state
- No requirement that workers are launched from an already-present local source checkout.
- Worker registration/capability advertisement.
- Dispatch and lease model for jobs.
- Correlation IDs and resumable checkpoints across worker boundaries.

### D) API-driven backend with gqlgen

V1 control plane must be GraphQL-first:

- `gqlgen` schema as primary contract.
- Mutations for spawn/send/cancel/restore/assign/merge actions.
- Queries for sessions, tasks, workflows, worker fleet status.
- Subscriptions for real-time session + orchestrator events.

### E) Cross-platform UI

V1 UI requirements:

- Works across macOS/Linux/Windows.
- Runtime-configurable GraphQL endpoint (server address entered/updated at runtime).
- Live status board for all sessions + orchestrator attention zones.
- Session detail view with stream output and control actions.

### F) Tracker-agnostic taskboard model

V1 task planning/execution must be board-provider agnostic:

- Implement local JSON taskboard support as a newly authored V1 adapter (MVP behavior-compatible, not code-shared).
- Add a canonical internal taskboard domain model independent of vendor schemas.
- Support adapter-based ingestion/sync for external systems over time (e.g., Jira, Linear).
- Treat board providers as part of the `tracker` slot so orchestration logic stays unchanged.
- Allow project-level board source selection (local JSON vs external provider).

### G) Container-first deployment

V1 deployment and runtime packaging must be container-first:

- All core runtime components (API/control plane, worker runtime, and supporting services) must have first-class container images.
- Local development and integration execution should run from container orchestration (`docker compose`) before host-native shortcuts.
- Environment/configuration contracts must be externalized (env/secrets) so the same images run in local and remote environments.
- Health checks, startup/shutdown behavior, and observability hooks must be defined for container orchestration.
- Release readiness requires versioned container artifacts as the primary deployable output.

## Proposed V1 Shape

1. `v1` becomes an API/control-plane first system (GraphQL).
2. Orchestrator supervisor owns global monitoring and policy decisions.
3. Worker subsystem supports both local and remote executors.
4. Plugin contracts are formalized around the five agnostic slots with inward dependencies respected.
5. UI is a thin client over GraphQL + subscriptions.
6. Deployment baseline is container-first, with parity between local orchestration and remote runtime.

## Milestones

### Milestone 0 — Foundation (now)

- [x] `v1` module initialized.
- [x] Roadmap linked in root README.
- [ ] Define initial `gqlgen` schema (sessions, workers, workflows, events).
- [ ] Define container-first bootstrap artifacts for V1 (initial `Dockerfile` set and `docker compose` topology).

### Milestone 1 — Core contracts

- [ ] Define five-slot plugin interfaces in `v1` (`agent`, `scm`, `tracker`, `notifier`, `client`).
- [ ] Implement plugin registry + configuration loading.
- [ ] Add orchestrator supervisor domain/application skeleton.
- [ ] Define canonical taskboard domain + provider adapter contract (JSON first).
- [ ] Define runtime env/secrets contract for containerized API and worker components.

### Milestone 2 — Execution plane

- [ ] Implement local worker adapter.
- [ ] Implement remote worker adapter contract + dispatcher.
- [ ] Add SCM-backed work execution flow.
- [ ] Implement first board adapters: local JSON + external provider abstraction entrypoint.
- [ ] Run worker execution paths through containerized runtime profiles (local and remote parity).

### Milestone 3 — Realtime control plane

- [ ] Implement GraphQL queries/mutations.
- [ ] Implement GraphQL subscriptions for session/orchestrator events.
- [ ] Persist and expose session/event streams.

### Milestone 4 — Cross-platform UI

- [ ] Build UI with runtime GraphQL endpoint configuration.
- [ ] Implement dashboard zones (merge/respond/review/working/done).
- [ ] Implement session stream and action console.

## Immediate Next Actions

1. Create the initial `gqlgen` schema + resolver stubs in `v1`.
2. Define worker capability and lease contracts (local/remote parity).
3. Draft orchestrator supervisor state model and event taxonomy.
4. Author the first V1 container bootstrap (`Dockerfile` + `docker compose`) for API and worker startup parity.

## V1 Interface Stance (Explicit)

V1 does not support terminal-app operation as product UX.

- No tmux UX.
- No CLI UX.
- No terminal-first user interaction mode.
- Cross-platform client is the primary and immediate user surface for operating the system.

Terminal/process capabilities remain internal execution-plane mechanics only and are not exposed as end-user operating surfaces.

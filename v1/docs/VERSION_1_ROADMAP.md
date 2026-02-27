# Version 1 Roadmap (High-Level)

## Purpose

Define the high-level path to V1 release and the release-level outcomes that must be achieved.

Detailed execution plans belong in `docs/roadmap/*`.

## V1 Mission

Ship a GraphQL-first orchestrator platform with local + remote workers, real-time operational streams, tracker-agnostic planning, and a cross-platform client experience, all under strict DDD boundaries and container-first runtime parity.

## Non-Negotiable Constraints

- V1 is a ground-up rewrite in `v1/`.
- `mvp/` is inspiration only; no copy/migrate/reuse of source.
- Canonical slots: `agent`, `scm`, `tracker`, `notifier`, `client`.
- Dependency direction: `interface -> application -> domain`.
- End-user product UX is client-first (not terminal/tmux/CLI).
- Containerized runtime is the default deployable contract.

## Release Pillars

1. **Architecture Foundation**
   - Enforced placement rules and slot boundaries.
   - Stable composition roots and runtime contracts.

2. **Execution Plane**
   - Local and remote worker parity.
   - SCM-backed remote bootstrap and resumable checkpoints.

3. **Control Plane**
   - GraphQL schema as primary contract.
   - Queries, mutations, and subscriptions for orchestrator operations.

4. **Tracker-Agnostic Planning**
   - Canonical taskboard domain model.
   - Local JSON adapter + external provider extension boundary.

5. **Realtime Observability and Streams**
   - Session activity, agent output, and orchestrator decision streams.
   - Correlation IDs through execution and events.

6. **Cross-Platform Client**
   - Runtime-configurable GraphQL endpoint.
   - Operational dashboard + session detail/controls.

7. **Container-First Delivery**
   - API and worker container images.
   - Compose-based local parity and release-ready artifact flow.

## High-Level Delivery Phases

### Phase 0 — Foundation

- Finalize scope and placement governance.
- Establish baseline project layout and startup lifecycle contracts.

### Phase 1 — Core Runtime Contracts (SCM/Agent/Worker First)

- Define `agent` and `scm` slot interfaces with initial adapters.
- Define worker lease/capability and execution contracts.
- Validate SCM-backed execution baseline before supervisor policy depth.

### Phase 2 — Orchestrator + GraphQL Core

- Implement supervisor state/event taxonomy over real execution context.
- Establish GraphQL contract surface and core control actions.

### Phase 3 — Tracker + Realtime + Client Surface

- Deliver canonical tracker/taskboard ingestion flows.
- Deliver subscription/event stream pathways.
- Deliver first usable cross-platform client flows.

### Phase 4 — Release Hardening

- Container/runtime parity validation.
- Reliability, operability, and release criteria closure.

## V1 Release Exit Criteria

V1 is release-ready only when all are true:

1. GraphQL control plane supports core orchestrator workflows.
2. Local and remote worker modes both execute via shared orchestration contracts.
3. Realtime session/agent/orchestrator streams are available to clients.
4. Tracker flows operate on canonical internal model with adapter boundaries.
5. Cross-platform client can operate the system without terminal-first UX.
6. API and worker are containerized and validated under compose-based runtime profiles.
7. DDD boundary and placement constraints are maintained across implementation.

## Risks to Track

- Slot-boundary leakage under delivery pressure.
- Remote worker bootstrap complexity (SCM auth, checkout, resume fidelity).
- Stream volume/performance under concurrent sessions.
- Drift between API contract and client expectations.
- Host-native shortcuts eroding container-first parity.

## Immediate Priority Focus

1. Land SCM/agent/worker runtime contracts before supervisor policy depth.
2. Keep roadmap slices aligned to this dependency order.
3. Validate runtime parity early via containerized integration paths.

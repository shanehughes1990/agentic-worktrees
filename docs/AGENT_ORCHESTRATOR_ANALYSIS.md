# Agent Orchestrator Reference Analysis

This document captures architecture highlights extracted from `.docs/agent-orchestrator` and maps them to this repo’s V1 target shape.

## Scope and Separation

- This is reference analysis only.
- V1 implementation direction is defined in `docs/VERSION_1_ROADMAP.md`.
- V1 remains a ground-up rewrite; reference patterns are inspiration, not import/migration targets.

## Reference Signals Worth Carrying Forward

From the reference project, the strongest reusable ideas are:

- Plugin-oriented capability boundaries with registry-driven composition.
- Clear orchestration lifecycle management (task/session state transitions).
- Streaming-first operator visibility (events/output flow to UI/API).
- Worker-style execution abstraction where orchestration is decoupled from runtime location.

These signals align with our V1 goals for remote/local worker parity, GraphQL control-plane, and runtime event streaming.

## Mapping to Our V1 Agnostic Layers

Canonical V1 slots are:

- `agent`
- `scm`
- `tracker`
- `notifier`
- `client`

### 1) `agent`

Responsibility:

- AI execution adapter and session interaction boundary.
- Tool/use-cycle execution semantics and model-provider abstraction.
- Session introspection surfaces for orchestrator policy decisions.

### 2) `scm`

Responsibility:

- Source branch/commit context, PR lifecycle, CI/review metadata.
- Provider abstraction for GitHub now, alternatives later.
- Required support for remote worker bootstrap from `origin` using credentials.

### 3) `tracker`

Responsibility:

- Work-item and board/task lifecycle ingestion and synchronization.
- Canonical internal taskboard domain independent from external schemas.
- Provider adapters including local JSON board source and external systems.

Placement note:

- Board/task provider behavior is part of `tracker`; no separate board layer is used.

### 4) `notifier`

Responsibility:

- Human escalation and delivery fanout.
- Routing by attention/urgency policy from orchestrator decisions.

### 5) `client`

Responsibility:

- Cross-platform user control surface over GraphQL.
- Live status and stream rendering, including action invocation UX.
- Runtime-configured backend endpoint support.

## Runtime and Terminal Positioning (Not Canonical Slots)

The reference project includes runtime/terminal-style concerns, but in our V1 architecture these are **execution-plane capabilities**, not top-level agnostic slots.

- Runtime concerns belong to worker infrastructure and dispatch/lease orchestration.
- Terminal/PTY concerns stay internal to execution adapters and are not an end-user operating surface.
- End-user operations are performed through the cross-platform client only.
- Keeping these out of canonical slots preserves strict DDD boundaries and avoids mixing platform mechanics with domain/provider contracts.

## Architectural Fit With V1 Roadmap

This mapping is consistent with V1 non-negotiables in `docs/VERSION_1_ROADMAP.md`:

- Five-slot agnostic model.
- SCM-backed remote worker self-bootstrap from origin.
- Real-time stream surfaces.
- GraphQL-first control plane (`gqlgen`).
- Container-first deployment baseline.
- No terminal/tmux/CLI user operation mode.

## Practical Implication for Build Order

- Define slot contracts first (`agent`, `scm`, `tracker`, `notifier`, `client`).
- Implement worker/runtime contracts under execution plane (local + remote parity).
- Add tracker adapters in this order: local JSON board source, then external providers.
- Expose orchestration and stream state through GraphQL and client surfaces.

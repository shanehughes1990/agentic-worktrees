# Agent Orchestrator Reference Analysis (Informational)

## Purpose

Summarize useful architectural signals from reference projects and map them to V1 without implying source migration.

## Usage Rules

- This document is informational only.
- `mvp/` and other references are inspiration, not implementation source.
- V1 code and contracts must be authored natively in `v1/`.

## Key Signals Worth Reusing Conceptually

- Plugin-capability boundaries with explicit contracts.
- Orchestration lifecycle/state-machine discipline.
- Streaming-first operational visibility.
- Worker abstraction decoupled from runtime location.

## Mapping to V1 Canonical Slots

- `agent`
  - AI execution adapter and session introspection boundary.
- `scm`
  - Source/branch/PR/review integration and remote bootstrap context.
- `tracker`
  - Work-item + taskboard ingestion/sync under one canonical internal model.
- `notifier`
  - Human escalation and notification routing.
- `client`
  - Cross-platform GraphQL control surface for operators.

## Important V1 Interpretation

- Runtime/terminal process mechanics are execution-plane internals, not canonical top-level slots.
- End-user operation is client-driven; terminal tooling is internal/developer-only.
- All decisions must respect DDD boundaries and container-first runtime parity.

## Practical Outcome for Planning

- Define contracts first, adapters second, UI flows third.
- Keep orchestration logic provider-agnostic and contract-driven.
- Require correlation IDs and resumable checkpoints across long-running flows.

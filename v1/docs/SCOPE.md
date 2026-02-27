# V1 Scope Boundaries

## Purpose

Define what V1 includes, what V1 excludes, and the constraints all implementation work must follow.

## Rewrite Policy (Mandatory)

V1 is a ground-up rewrite contained in `v1/`.

- `mvp/` is reference and inspiration only.
- No source import, package reuse, file copy, or code migration from `mvp/` into `v1/`.
- V1 contracts, schemas, adapters, and flows are authored natively in V1.

## In Scope (V1)

- GraphQL-first control plane using `gqlgen`.
- Cross-platform client as the end-user operating surface.
- Local and remote worker execution with SCM-backed bootstrap.
- Slot model with canonical boundaries: `agent`, `scm`, `tracker`, `notifier`, `client`.
- Tracker-agnostic taskboard model with local JSON provider first.
- Realtime streams for session activity, agent output, and orchestrator decisions.
- Container-first runtime and deployment baseline.

## Out of Scope / Non-Goals (V1)

- Terminal-app product UX.
- tmux-based operator UX.
- CLI-first end-user workflows.
- Slot-first architecture rooted at `pkg/<slot>`.
- Feature work that bypasses GraphQL as the control-plane contract.

Developer-only command-line tooling for setup and testing is allowed but is not product UX.

## Scope Gates for New Work

A new item is in scope only if it satisfies all checks:

1. Fits one or more V1 in-scope capabilities above.
2. Preserves DDD boundaries and dependency direction.
3. Supports or does not block GraphQL-first control plane operation.
4. Does not introduce terminal-first user interaction as product behavior.
5. Can be deployed/run in containerized runtime profiles.

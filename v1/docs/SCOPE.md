# V1 Scope Boundaries

## Rewrite Policy

V1 is a ground-up rewrite contained in `v1/`.

- `mvp/` is reference/inspiration only.
- No source import, package reuse, file copy, or code migration from `mvp/` into `v1/`.
- V1 contracts, schemas, adapters, and flows are authored natively in V1.

## In Scope

- GraphQL-first control plane.
- Cross-platform client as the end-user operating surface.
- Local + remote worker execution parity with SCM-backed bootstrap.
- Slot model: `agent`, `scm`, `tracker`, `notifier`, `client`.
- Container-first deployment baseline.

## Non-Goals (V1)

- Terminal-app product UX.
- tmux-based operation.
- CLI-first user workflows.
- Slot-first architecture rooted at `pkg/<slot>`.

Developer-only command-line tooling for setup/test may exist, but it is not product UX.

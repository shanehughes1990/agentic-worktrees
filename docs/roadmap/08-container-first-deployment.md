# Slice 08 — Container-First Deployment

## Objective

Make containerized runtime the default and primary deployment contract for V1.

## Scope

- First-class images for API/control plane and worker runtime.
- `docker compose` topology for local development and integration parity.
- Env/secrets contract externalization for image portability.
- Health checks and lifecycle hooks for orchestrated runtime.
- Versioned container artifacts as release outputs.

## Acceptance Criteria

- Local integration runs from compose baseline, not host-only scripts.
- Same image set works in local and remote environments via config only.
- Worker execution paths validate in containerized profiles.

## Dependencies

- Worker execution plane.
- GraphQL control plane service startup.
- Observability and runtime lifecycle conventions.

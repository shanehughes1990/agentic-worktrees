# V1 Placement Rules

## Purpose

Define canonical package placement and dependency direction for all V1 implementation work.

## Canonical Structure

V1 is layer-first with slot boundaries mirrored inside layers.

- `internal/domain/<slot>`
- `internal/application/<slot>`
- `internal/infrastructure/<slot>`
- `internal/interface/*` (delivery surfaces only)
- `cmd/*` (composition roots only)

Canonical slots: `agent`, `scm`, `tracker`, `notifier`, `client`.

## Mandatory Dependency Direction

- `interface -> application -> domain`
- `infrastructure` implements ports/contracts owned by inner layers
- `domain` must not depend on `application`, `interface`, or `infrastructure`

## Layer Ownership

- `cmd/*`
  - Startup wiring and composition roots only.
  - No business orchestration.
- `internal/interface/*`
  - Transport, validation, mapping, and delivery handlers.
  - No business rules.
- `internal/application/*`
  - Use-case orchestration and process boundaries.
  - Coordinates domain + ports.
- `internal/domain/*`
  - Business meaning, invariants, value objects, aggregates, domain services.
- `internal/infrastructure/*`
  - Concrete adapters for queue, SCM, persistence, APIs, filesystem, runtime technology.

## Forbidden Placement

- Business logic in `cmd/*` or `internal/interface/*`.
- Direct `interface -> domain` bypass for orchestration.
- Infrastructure packages owning use-case orchestration.
- New slot-first roots (`pkg/agent`, `pkg/scm`, etc.) as primary architecture.

## Placement Decision Rule

When placement is ambiguous:

1. Put business meaning/invariants in `domain`.
2. Put orchestration in `application`.
3. Put adapters in `infrastructure`.
4. Keep transport concerns in `interface`.
5. Keep process wiring in `cmd`.

If a requested placement violates these rules, the change must be redirected before implementation.

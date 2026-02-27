# V1 Placement Rules

## Canonical Architecture Shape

Layer-first structure with slot boundaries mirrored inside layers:

- `internal/domain/<slot>`
- `internal/application/<slot>`
- `internal/infrastructure/<slot>`
- `internal/interface/*` (delivery only)

Slots: `agent`, `scm`, `tracker`, `notifier`, `client`.

## Dependency Direction (Mandatory)

- `interface -> application -> domain`
- `infrastructure` implements ports/contracts owned by inner layers.
- `domain` must not depend on `application`, `interface`, or `infrastructure`.

## Ownership Rules

- `cmd/*`: composition roots and startup wiring only.
- `interface/*`: input/output mapping, transport concerns, and delivery handlers only.
- `application/*`: use-case orchestration and process boundaries.
- `domain/*`: business meaning, invariants, and core contracts.
- `infrastructure/*`: concrete adapters for external systems and runtime technology.

## Forbidden Placement

- Business rules in `cmd/*` or `internal/interface/*`.
- Direct `interface -> domain` bypass for business flow orchestration.
- Infrastructure packages owning use-case orchestration.
- New slot-first roots such as `pkg/agent`, `pkg/scm`, etc. as primary architecture.

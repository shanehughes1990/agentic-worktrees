# Slice 00 — Project Layout and Scope Docs

## Objective

Establish the V1 foundation by defining scope documentation and creating the canonical repository structure inside `v1/` before feature implementation begins.

## Scope

- Define V1 scope boundaries and delivery framing docs.
- Define the canonical V1 folder/package layout aligned with DDD layering.
- Define initial conventions for package placement and dependency direction.
- Define where vertical-slice docs, architecture docs, and API schema docs live.

## Required Repository Structure (V1)

- `v1/cmd/`
  - Process entrypoints (API server and worker process, plus developer-only setup/test tooling where needed).
- `v1/internal/domain/`
  - Business entities, value objects, invariants, domain services.
- `v1/internal/application/`
  - Use-cases, orchestration flows, transaction/process boundaries.
- `v1/internal/infrastructure/`
  - Concrete adapters (SCM provider, queue, persistence, runtime adapters).
- `v1/internal/interface/`
  - Delivery surfaces (GraphQL schema/resolvers/server wiring and worker handlers).
- `v1/docs/`
  - V1-local technical docs (schema notes, ADR-like records, operational contracts).

## Acceptance Criteria

- `v1/` directory tree exists with DDD-aligned base folders.
- Scope documentation clearly states rewrite policy and non-goals.
- Placement rules are documented so new code lands in correct layers by default.
- Vertical-slice execution can proceed without structure ambiguity.

## Dependencies

- None. This slice is the mandatory first step for all remaining slices.

## Non-Negotiable Interface Policy (V1)

V1 is **not** a terminal application.

- No CLI-first UX in V1.
- No tmux-based UX in V1.
- No terminal-app workflows as a supported primary or fallback operator surface in V1.
- System usage is expected through the cross-platform client from first use.
- Any command-line utilities that may exist for developer setup/testing are non-user-facing and out of scope for runtime product UX.

This policy is mandatory and takes precedence over any legacy MVP operation patterns.

## Approved Structural Plan (V1)

This is the official V1 plan for architecture and package layout.

- Canonical capability boundaries (`agent`, `scm`, `tracker`, `notifier`, `client`) are defined as domain barriers first.
- Structure is layer-first, slot-within-layer (not slot-first package roots):
  - `v1/internal/domain/<slot>`
  - `v1/internal/application/<slot>`
  - `v1/internal/infrastructure/<slot>`
- `v1/internal/interface/*` remains delivery-only (GraphQL and worker delivery surfaces), with no business-rule ownership.
- `v1/cmd/*` remains composition-root only (wiring and startup), not business orchestration.
- `pkg/<slot>` is not the primary architecture pattern for V1.

Dependency direction is mandatory:

- `interface -> application -> domain`
- `infrastructure` implements ports/contracts owned by inner layers.

Any change proposal that conflicts with this plan should be treated as out of scope for V1 unless this document is explicitly revised.

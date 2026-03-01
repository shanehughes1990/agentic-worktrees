# Slice 00 — Project Layout, Governance, and Persistence Baseline

## Status

- Completed: **Completed**
- Reviewed At: **2026-03-01T01:25:40Z**

## Objective

Establish canonical V1 structure and governance so all implementation work follows strict DDD boundaries and a single durable persistence strategy.

## Task Checklist

- [x] Publish V1-local scope documentation (`SCOPE.md`).
- [x] Publish V1 placement/dependency rules (`PLACEMENT_RULES.md`).
- [x] Establish canonical `v1/internal/{domain,application,infrastructure,interface}` layout.
- [x] Document high-level vs detailed roadmap hierarchy (`VERSION_1_ROADMAP.md` vs `docs/roadmap/*`).
- [x] Define PostgreSQL as the durable system of record; keep Redis scoped to queue transport internals.

## Deliverables

- V1-local scope documentation (`SCOPE.md`) and placement rules (`PLACEMENT_RULES.md`).
- Canonical layer-first folder layout in `v1/internal/*`.
- Clear ownership rules for `cmd`, `interface`, `application`, `domain`, and `infrastructure`.
- Persistence governance rule: business state persists in Postgres, not transient runtime memory.

## In Scope

- Documentation and structure required to unblock all subsequent slices.
- Dependency direction and package ownership rules.
- Persistence placement rules and non-goal enforcement.

## Out of Scope

- Feature-specific business implementation.
- Provider-specific adapter depth.
- UI and control-plane feature breadth.

## Acceptance Criteria

- All V1 technical docs use consistent terminology and hierarchy.
- Canonical package placement is explicit and unambiguous.
- Architecture guardrails are documented and enforceable.
- Persistence direction is explicit: Postgres is authoritative for durable operational/business state.

## Dependencies

- None. This is the foundational slice.

## Exit Check

This slice is complete only when teams can place new code and persistence responsibilities without architecture disputes.

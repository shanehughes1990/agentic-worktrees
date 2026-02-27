# Slice 00 — Project Layout and Governance

## Objective

Establish canonical V1 structure and documentation governance so all implementation work follows consistent boundaries.

## Task Checklist

- [x] Publish V1-local scope documentation (`SCOPE.md`).
- [x] Publish V1 placement/dependency rules (`PLACEMENT_RULES.md`).
- [x] Establish canonical `v1/internal/{domain,application,infrastructure,interface}` layout.
- [x] Document high-level vs detailed roadmap hierarchy (`VERSION_1_ROADMAP.md` vs `docs/roadmap/*`).
- [x] Reorder roadmap slices to follow integration-first execution order.

## Deliverables

- V1-local scope documentation (`SCOPE.md`) and placement rules (`PLACEMENT_RULES.md`).
- Canonical layer-first folder layout in `v1/internal/*`.
- Clear ownership rules for `cmd`, `interface`, `application`, `domain`, and `infrastructure`.
- Documentation hierarchy that separates high-level roadmap from detailed slice plans.

## In Scope

- Documentation and structure required to unblock all subsequent slices.
- Dependency direction and package ownership rules.
- Rewrite policy and non-goal enforcement.

## Out of Scope

- Feature-specific business implementation.
- Provider-specific adapter depth.
- UI and control-plane feature breadth.

## Acceptance Criteria

- All V1 technical docs use consistent terminology and hierarchy.
- Canonical package placement is explicit and unambiguous.
- Architecture guardrails are documented and enforceable.
- Future slices can execute without structural ambiguity.

## Dependencies

- None. This is the foundational slice.

## Exit Check

This slice is complete only when teams can place new code without architecture disputes.

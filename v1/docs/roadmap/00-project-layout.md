# Slice 00 — Project Layout and Governance

## Objective

Establish canonical V1 structure and documentation governance so all implementation work follows consistent boundaries.

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

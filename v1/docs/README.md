# V1 Documentation Index

This folder contains V1-local technical and delivery documentation for the `v1/` rewrite.

## Document Roles

- `VERSION_1_ROADMAP.md`
  - High-level release roadmap and V1 release criteria.
  - Executive summary of what must be true at V1 release.
- `SCOPE.md`
  - Scope boundaries, non-goals, and rewrite constraints.
- `PLACEMENT_RULES.md`
  - Architecture and package placement rules for implementation.
- `AGENT_ORCHESTRATOR_ANALYSIS.md`
  - Reference analysis from external inspiration projects.
  - Informational only; not a source migration plan.

## Detailed Plans

Detailed implementation plans live in `roadmap/`.

- `roadmap/README.md` explains how slice docs are used.
- `roadmap/00-08-*.md` define detailed objectives, deliverables, and acceptance criteria by slice.

## Source of Truth Hierarchy

1. `SCOPE.md` and `PLACEMENT_RULES.md` define constraints.
2. `VERSION_1_ROADMAP.md` defines release-level outcomes.
3. `roadmap/*.md` define detailed execution plans under those constraints.

If a detailed slice conflicts with scope or placement rules, scope and placement rules win.

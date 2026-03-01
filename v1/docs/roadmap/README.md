# V1 Roadmap Slices (Detailed Plans)

## Purpose

This directory contains detailed execution plans for V1 delivery slices.

- `../VERSION_1_ROADMAP.md` is the high-level release roadmap.
- Files here are implementation-level plans, scope details, and acceptance criteria by slice.

## Slice Files

0. `00-project-layout.md`
1. `01-agent-scm-core.md`
2. `02-worker-execution-plane.md`
3. `03-orchestrator-supervisor.md`
4. `04-tracker-taskboard.md`
5. `05-realtime-streams.md`
6. `06-graphql-control-plane.md`
7. `07-client-experience.md`
8. `08-container-first-deployment.md`
9. `09-postgres-persistence-conversion.md`

## Execution Order

File numbering now matches recommended implementation order to maximize integration confidence at each step:

1. `00-project-layout.md`
2. `01-agent-scm-core.md`
3. `02-worker-execution-plane.md`
4. `03-orchestrator-supervisor.md`
5. `04-tracker-taskboard.md`
6. `05-realtime-streams.md`
7. `06-graphql-control-plane.md`
8. `07-client-experience.md`
9. `08-container-first-deployment.md`
10. `09-postgres-persistence-conversion.md`

This order ensures supervisor policy and control-plane behavior are validated against real SCM-backed execution flows.

## Common Rules for All Slices

- V1 rewrite policy is mandatory (no source migration from `mvp/`).
- Placement and dependency rules from `../PLACEMENT_RULES.md` are mandatory.
- User-facing operation is client-first; terminal-first UX is out of scope.
- Each slice must define clear acceptance criteria and dependencies.

## How to Use

1. Read the high-level roadmap first.
2. Execute slices in numbered order unless an explicit dependency update is approved.
3. Validate acceptance criteria before marking a slice complete.
4. Update docs when scope or sequencing changes.

# V1 Vertical Slice Roadmaps

This directory breaks V1 delivery into vertical slices.

Use `docs/VERSION_1_ROADMAP.md` as the canonical source of truth for global requirements and milestone intent.

## Slices

0. `00-project-layout.md`
1. `01-orchestrator-supervisor.md`
2. `02-agent-scm-core.md`
3. `03-tracker-taskboard.md`
4. `04-worker-execution-plane.md`
5. `05-realtime-streams.md`
6. `06-graphql-control-plane.md`
7. `07-client-experience.md`
8. `08-container-first-deployment.md`

## Shared Constraints

- V1 is a ground-up rewrite in `v1/`.
- No code import/reuse/migration from `mvp/`.
- DDD dependency direction remains mandatory: interface -> application -> domain, with infrastructure implementing ports.
- Runtime and terminal capabilities are execution-plane concerns, not canonical agnostic slots.

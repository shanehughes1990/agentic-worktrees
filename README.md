# agentic-worktrees

A fully autonomous multi-agentic code builder.

The core goal is to turn provided documentation into a micro task board, spawn parallel Copilot SDK agents to execute those tasks, and keep progressing until the board is complete.

## Vision

`agentic-worktrees` is designed to:

1. Ingest scoped documentation.
2. Derive small, actionable implementation tasks.
3. Execute tasks in parallel via Copilot SDK agents.
4. Track task state and mark tasks complete as work lands.
5. Continue orchestration until all tasks are finished.

## Execution Model

High-level workflow:

1. **Document Intake**: collect and normalize requirement/context documents.
2. **Micro-Task Planning**: decompose documentation into bounded tasks with clear done criteria.
3. **Parallel Agent Dispatch**: assign tasks to isolated Copilot SDK agents running concurrently.
4. **Task Lifecycle Tracking**: move tasks through states (`not-started`, `in-progress`, `completed`, `blocked`).
5. **Autonomous Completion Loop**: re-plan remaining work and dispatch again until complete.

## Architecture Boundaries (DDD)

This repository follows strict Domain-Driven Design layering:

- `internal/interface`: admission surfaces (CLI/MCP/HTTP/worker handlers), input validation, and mapping.
- `internal/application`: use-case orchestration and process boundaries.
- `internal/domain`: business invariants, entities, value objects, and domain services.
- `internal/infrastructure`: adapters for persistence, queues, SDK integrations, and observability.

Dependency direction is enforced as:

- `interface -> application -> domain`
- `infrastructure` implements ports required by inner layers.

## Repository Layout

- `docs/`: project documentation.
- `internal/`: DDD layers (`application`, `domain`, `infrastructure`, `interface`).
- `go.mod`: Go module and language version.

## Current Status

This repository is currently scaffolded for architecture-first development.

Implemented now:

- DDD folder layout.
- Project/release metadata configuration.

Planned next (aligned to this project goal):

- micro task board generation from documentation,
- parallel Copilot SDK agent orchestration,
- task state transitions and completion tracking,
- end-to-end autonomous completion loop.

## Build Metadata

- Go version: `1.25.3`
- Release project/binary name: `agentic-worktrees`
- Expected CLI entrypoint for builds: `./cmd/agentic-worktrees`

Note: the Go module path currently uses `github.com/goeasycare/agentic-workflows`; naming alignment can be handled in a follow-up change.

## First Implementation Milestones

1. Define task board domain model and state machine.
2. Implement application use-case for documentation-to-task decomposition.
3. Add infrastructure adapter for parallel Copilot SDK agent execution.
4. Implement orchestration loop that converges on full board completion.
5. Expose interface entrypoint for autonomous run execution.

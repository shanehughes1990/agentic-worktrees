# Project Documentation

This folder defines the full product and engineering scope for `agentic-worktrees`.

## Document Map

- [01 - Scope and Product Definition](./01-scope-and-product-definition.md)
- [02 - Functional and Non-Functional Requirements](./02-requirements.md)
- [03 - Architecture](./03-architecture.md)
- [04 - Workflow and State Machine](./04-workflow-and-state-machine.md)
- [05 - Git Strategy and Conflict Handling](./05-git-strategy-and-conflicts.md)
- [06 - Kanban Model and Scheduling](./06-kanban-and-scheduling.md)
- [07 - Copilot ADK Integration](./07-copilot-adk-integration.md)
- [08 - Realtime Observability](./08-realtime-observability.md)
- [09 - Containerization and Runtime](./09-containerization.md)
- [10 - Failure Modes and Recovery](./10-failure-modes-and-recovery.md)
- [11 - Security and Safety Controls](./11-security-and-safety.md)
- [12 - Testing, SLOs, and Release Readiness](./12-testing-slos-and-release.md)
- [13 - Delivery Plan](./13-delivery-plan.md)
- [14 - Persistence and Realtime Abstractions](./14-persistence-and-realtime-abstractions.md)
- [15 - Git Worktree Task Lifecycle](./15-git-worktree-task-lifecycle.md)
- [16 - Kanban Ingestion and Board Seeding Pipeline](./16-kanban-ingestion-and-seeding.md)
- [17 - Local Development Infrastructure](./17-local-development-infrastructure.md)

## Current Goal

Build a resilient Go-based orchestration system whose execution core is durable background processing with Asynq. The first interface is a `urfave/cli/v3` CLI, while core orchestration remains interface-agnostic.

## Design Principles

- Asynq-first durable execution and retries
- Interface-agnostic orchestration services
- Deterministic task/worktree lifecycle
- Idempotent handlers and checkpointed side effects
- Swappable infrastructure via stable interfaces
- Safe-by-default Git automation
- Full realtime visibility of all running agents
- Container-ready from day one

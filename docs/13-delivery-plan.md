# 13 - Delivery Plan

## Phase 0 - Foundations

- project skeleton and package boundaries
- config system and logging baseline
- durable state store abstraction
- CLI bootstrapping with `urfave/cli/v3`

## Phase 1 - Durable Queue Core (Asynq)

- Asynq client/server integration
- Redis connectivity and health checks
- queue definitions, priorities, and retry policies
- idempotent task handler scaffolding

## Phase 2 - Core Workflow Loop

- task model and dependency graph
- planner + enqueue service
- worktree manager
- deterministic state machine + checkpoints

## Phase 3 - Git and PR Automation

- fetch/rebase/push orchestration
- PR create/update integration
- merge safety gates
- conflict policy framework

## Phase 4 - Copilot ADK Runtime

- runtime adapter contract implementation
- event streaming and heartbeats
- run cancellation and retry controls

## Phase 5 - Realtime Insights

- CLI status views and tailing
- metrics and alert rules

## Phase 6 - Container Readiness

- container image and runtime hardening
- health/readiness probes
- persistence and restart recovery validation

## Phase 7 - Reliability Hardening

- chaos drills and failure-mode closure
- dead-letter workflows
- SLO validation and tuning

## Phase 8 - Optional Interface Adapters (Post-MVP)

- tmux supervision adapter
- additional operator interfaces (TUI/Web/API)

## Milestone Exit Criteria

- end-to-end automation cycle proven on parallel task board
- repeatable rebase/merge behavior on latest origin each iteration
- clear operator recovery paths for all documented failure classes

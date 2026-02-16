# 12 - Testing, SLOs, and Release Readiness

## Test Strategy

### Unit Tests

- planner ordering, fairness, and dependency gating
- Asynq task payload validation and routing rules
- handler idempotency and checkpoint guards
- preflight dependency/version/auth validator behavior
- model policy validator (default `gpt-5.3-codex`, override allowlist, fallback rules)
- schema version parser and compatibility-window validator
- migration normalizer behavior for supported older schema versions
- state machine transition validity
- git strategy decision logic
- retry/backoff policy behavior

### Integration Tests

- real git repo fixture with worktree lifecycle
- Redis + Asynq integration with real queue execution
- PR/rebase/merge workflow using mocked remote APIs where needed
- checkpoint resume after forced process kill
- conflict policy matrix behavior
- admission-control blocking when `git`/`gh` checks fail
- admission-control blocking when `gh` auth is invalid
- task execution uses `gpt-5.3-codex` when no override exists
- unsupported PRD model override falls back to `gpt-5.3-codex` with warning event
- ingestion succeeds for supported legacy schema versions via normalization
- admission-control blocks unsupported schema versions until migrated

### End-to-End Tests

- full automation cycle on synthetic task board
- parallel tasks with dependency chains
- containerized run validation

### Chaos/Failure Tests

- network interruptions
- state store restart/unavailability
- runtime heartbeat loss
- disk pressure scenarios

## SLO Proposals (Initial)

- run completion success rate (non-terminal): >= 95%
- successful resume after crash: >= 99%
- queue enqueue-to-start p95: <= 3s (under target load)
- dead-letter rate per 1k jobs: <= agreed threshold per environment
- mean conflict detection latency: <= 5s
- mean time to visibility of run start event: <= 2s

## Release Gates

- all critical-path integration tests green
- failure drill checklist passed
- security policy checks enabled
- observability dashboards and alerts configured
- rollback and recovery runbook validated
- preflight admission checks verified in target environment before enabling intake
- model policy gate verified: default and override-fallback behavior validated
- schema compatibility gate verified: backward-read support and migration path validated

## Definition of Done (Feature Level)

A feature is done when:

- behavior is documented
- tests cover success + failure paths
- metrics/events emitted for key transitions
- operator controls and recovery behavior are defined

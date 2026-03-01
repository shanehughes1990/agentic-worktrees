# Slice 07 — Cross-Platform Client Experience

## Objective

Ship a cross-platform client that operates the system through GraphQL and presents durable operational state sourced from Postgres-backed control-plane models.

## Task Checklist

- [x] Lock client runtime: Flutter desktop-first.
- [x] Lock state management approach: Riverpod.
- [x] Lock typed GraphQL client/codegen stack.
- [ ] Implement runtime-configurable GraphQL endpoint setup flow.
- [ ] Implement multi-session status board and attention zones.
- [ ] Implement session detail stream view with replay-from-offset support.
- [ ] Implement control actions wired to GraphQL mutations.
- [ ] Add reconnect/resilience behavior for stream disruptions.
- [ ] Surface historical execution/dead-letter/supervisor history from persisted data.
- [ ] Add cross-platform packaging + smoke test coverage.

## Deliverables

- Runtime-configurable backend endpoint management.
- Multi-session status board with attention zones.
- Session detail view with live stream output + persisted history replay.
- Operator-oriented interaction model for escalation and intervention.

## In Scope

- Client flows required for day-1 V1 operation.
- Integration with GraphQL queries/mutations/subscriptions.
- Resilient reconnect behavior using persisted replay data.
- Platform-specific UI variants (different physical UI per target OS) sharing common backend/domain logic.

## Out of Scope

- Advanced visual polish not needed for operational readiness.
- Terminal-based fallback UX.

## Acceptance Criteria

- Client runs across supported OS targets.
- Operators can monitor and control workflows from one interface.
- Session activity and decision streams are visible in near-real-time and can recover from persisted replay history.
- Platform-specific UI layers can evolve independently without duplicating backend/domain logic.

## Dependencies

- Slice 06 control plane.
- Slice 05 stream layer.

## Client Technical Direction (Locked)

### Runtime + Platform Strategy

- **Framework:** Flutter (desktop-first).
- **Targets:** macOS, Windows, Linux in first release lane.
- **UI strategy:** physically different UI compositions per platform (not only responsive tweaks).

### Architecture Boundary Rule

- Keep platform UI in separate presentation modules.
- Keep shared domain/use-case/network/state logic in common modules.
- Platform layers must depend inward on shared modules; shared modules must not import platform UI code.

### State Management

- **Primary:** Riverpod (`Notifier`/`AsyncNotifier` for app + feature state).
- **Reason:** compile-safe DI, testable state graph, strong async handling for GraphQL query/subscription flows.

### GraphQL Typed Client + Codegen

- **Client stack:** `ferry` + `gql_http_link` + `gql_websocket_link`.
- **Typed codegen:** `gql_build` + `build_runner` generated operation/request/response types from `.graphql` documents.
- **Cache:** normalized GraphQL cache through Ferry store.

### Baseline App Structure

- `client/flutter/lib/shared/` — domain models, use-cases, repositories, GraphQL client setup, app-wide Riverpod providers.
- `client/flutter/lib/features/` — feature-level presentation + controllers (platform-agnostic view models/state).
- `client/flutter/lib/platform/windows/` — Windows-specific screens/layouts/compositions.
- `client/flutter/lib/platform/macos/` — macOS-specific screens/layouts/compositions.
- `client/flutter/lib/platform/linux/` — Linux-specific screens/layouts/compositions.
- `client/flutter/lib/app/` — app bootstrap, routing, platform dispatcher.

## MVP Milestones (Client Start Plan)

### Week 1 — Client Shell + Connectivity

- Initialize Flutter desktop app with Riverpod and Ferry foundations.
- Build app shell, route layout, and environment-safe endpoint configuration flow.
- Add connection status, auth/token input (if used), and GraphQL health probe.
- Wire existing operations only:
  - `Query.scmSupportedOperations`
  - `Query.supervisorDecisionHistory`
  - `Subscription.supervisorDecisionHistoryStream`
  - `Mutation.enqueueScmWorkflow`
- Exit: client can connect to API, run a query, receive a live subscription payload, and dispatch one SCM mutation.

### Week 2 — Control Plane P0 for Operator Board

- Land the minimum GraphQL read models required for client dashboard:
  - `Query.sessions`
  - `Query.session(runID)`
  - `Query.workers`
  - `Query.workflowJobs(runID, taskID)`
- Land minimum command mutations required for issue-intake governance path:
  - `Mutation.enqueueIngestionWorkflow(input)`
  - `Mutation.approveIssueIntake(input)`
- Implement first platform-specific board variants (macOS/Windows/Linux).
- Exit: operator can trigger ingestion, inspect session/job state, and approve issue intake from client.

### Week 3 — Session Detail + Historical Views

- Add session detail screen with supervisor history timeline and execution state panels.
- Add persisted historical views:
  - execution history
  - dead-letter history
  - supervisor decision history
- Add pagination/filter controls on list and history screens.
- Exit: operators can debug and recover a session from persisted historical state without terminal access.

### Week 4 — Realtime Streams + Replay

- Integrate stream channels from Slice 05:
  - `Subscription.sessionActivityStream`
  - `Subscription.workflowExecutionStream`
  - `Subscription.agentOutputStream`
  - `Subscription.supervisorDecisionHistoryStream` (already available)
- Implement reconnect behavior with replay cursor recovery.
- Exit: stream disconnect/reconnect recovers without data loss for the active session view.

### Week 5 — Packaging + Smoke Validation

- Cross-platform packaging and release artifact generation.
- Add smoke tests for:
  - startup + endpoint configuration
  - query/mutation/subscription happy paths
  - reconnect + replay behavior
- Exit: installable client artifacts pass smoke checks on supported targets.

## GraphQL Contract-First Backlog (Exact Operations)

### Available Now (can be consumed immediately)

- `Query.scmSupportedOperations`
- `Query.supervisorDecisionHistory(correlation)`
- `Subscription.supervisorDecisionHistoryStream(correlation, intervalMS)`
- `Mutation.enqueueScmWorkflow(input)`

### P0 Required Before Full Client Operations

- `Query.sessions`
- `Query.session(runID)`
- `Query.workers`
- `Query.workflowJobs(runID, taskID)`
- `Mutation.enqueueIngestionWorkflow(input)`
- `Mutation.approveIssueIntake(input)`

### P1 Required For Operational UX Completion

- `Query.executionHistory(correlation, page)`
- `Query.deadLetters(page, filter)`
- `Subscription.sessionActivityStream(correlation, fromOffset)`
- `Subscription.workflowExecutionStream(correlation, fromOffset)`
- `Subscription.agentOutputStream(correlation, fromOffset)`

## Start Gate Decision

Client implementation should start now with Week 1 scope and existing operations, while Week 2 P0 control-plane operations are developed in parallel. Full operational readiness still depends on Slice 05 replayable streams and Slice 06 read-model/query completion.

## Exit Check

This slice is complete when core orchestration operations can be run from the client alone with resilient replayable operational visibility.

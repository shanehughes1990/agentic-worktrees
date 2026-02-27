# Agent Orchestrator Analysis

This document captures findings from the local reference project at `.docs/agent-orchestrator` and keeps those findings separate from the V1 implementation roadmap.

## Scope of Analysis

Reference analyzed:

- `.docs/agent-orchestrator`

Primary focus:

- Plugin slot architecture and contract boundaries
- Core orchestrator stability points
- Session lifecycle and reaction model
- Real-time state propagation patterns
- Mapping to V1 slot strategy

## Key Architecture Signals

### 1) Plugin-first orchestration model

Observed in reference:

- Strong slot-based model with swappable implementations (`Runtime`, `Agent`, `Workspace`, `Tracker`, `SCM`, `Notifier`, `Terminal`).
- Core services orchestrate flow while plugins encapsulate provider/runtime specifics.

Implication:

- This pattern scales orchestration across providers without rewriting core logic.

### 2) Stable core orchestration services

Observed in reference:

- Session manager coordinates session CRUD and plugin usage.
- Lifecycle manager runs status polling, transition detection, reaction triggering, and escalation.

Implication:

- Core orchestration remains stable while plugin implementations vary.

### 3) Orchestrator-brain behavior is explicit

Observed in reference:

- Dedicated orchestrator prompt and supervisor behavior for coordinating worker sessions.
- Monitoring and intervention are treated as first-class behavior.

Implication:

- “Coordinator over workers” is an explicit architectural role, not an emergent side effect.

### 4) Real-time UX through event snapshots and polling

Observed in reference:

- Event stream endpoint (`/api/events`) uses SSE semantics with periodic snapshots.
- Session detail/dashboard endpoints refresh and enrich state with SCM/tracker data.

Implication:

- Real-time operator UX is feasible with incremental snapshots even before full push-native event buses.

### 5) SCM and tracker abstractions are critical

Observed in reference:

- SCM contract centralizes PR/CI/review state.
- Tracker contract centralizes issue/task context.

Implication:

- Provider neutrality depends on strict SCM/tracker contracts.

## Plugin Slot Responsibilities in Reference Repo

### Runtime

Responsibility:

- Owns execution substrate lifecycle for sessions (for example tmux/process implementations).
- Creates session runtime environments from workspace + launch config.
- Destroys/cleans runtime environments when sessions terminate.
- Provides runtime control primitives used by orchestration:
  - liveness checks
  - output capture
  - message injection
  - attach metadata for operator access

Key inputs:

- Session runtime configuration (session id, cwd/workspace path, launch command, env).

Key outputs:

- Runtime handle (id/type/data) and execution telemetry primitives (alive/output/attach info).

What it does not own:

- Workflow policy decisions, issue semantics, SCM/PR logic, or notification policy.

### Agent

Responsibility:

- Owns AI-tool adapter behavior (Claude/Codex/Aider/etc).
- Produces agent-specific launch command + environment contract.
- Implements activity/state detection from agent-native artifacts.
- Extracts agent session metadata for orchestration UX:
  - summary
  - session id
  - cost/usage where available
- Optionally provides restore/resume command behavior and workspace hook setup.

Key inputs:

- Session context + project config + runtime handle/workspace path.

Key outputs:

- Agent launch contract and interpreted activity/session information.

What it does not own:

- Runtime lifecycle itself, tracker issue truth, or SCM review/CI state.

### Workspace

Responsibility:

- Owns per-session isolated code context strategy (worktree/clone).
- Creates workspace from project + branch/session inputs.
- Restores workspace for resumed sessions.
- Destroys workspace on cleanup.
- Runs post-create preparation hooks where needed.

Key inputs:

- Project repository configuration, session id, branch naming decisions.

Key outputs:

- Workspace info (path, branch, session linkage) consumed by runtime/agent layers.

What it does not own:

- Agent reasoning, lifecycle policy, PR/CI state, or notification behavior.

### Tracker

Responsibility:

- Owns issue/work-item provider integration.
- Fetches issue/task details and lifecycle state.
- Generates issue-driven prompt context for worker sessions.
- Provides branch naming/context hints derived from issue identity.

Key inputs:

- Provider credentials/config + issue identifiers + project mapping.

Key outputs:

- Canonicalized issue/task context for session spawning and orchestration.

What it does not own:

- PR state machine, CI checks, review decisions, merge semantics.

### SCM

Responsibility:

- Owns source-control platform truth for PR lifecycle.
- Detects/loads PR state for active sessions.
- Exposes CI status/checks and review decision data.
- Exposes mergeability and blockers.
- Executes merge/close operations.

Key inputs:

- Repository/project identity + branch/session association + provider credentials.

Key outputs:

- Canonical PR/CI/review/mergeability signals used by lifecycle manager reactions.

What it does not own:

- Session scheduling policy, escalation policy, agent launch behavior.

### Notifier

Responsibility:

- Owns event delivery to humans/channels (desktop/slack/webhook/etc).
- Emits informational, warning, action-required, and urgent notifications.
- Optionally supports actionable notifications.

Key inputs:

- Orchestrator events and priority/context payloads.

Key outputs:

- Delivered messages/alerts to configured channels.

What it does not own:

- Event classification rules and reaction trigger policy.

### Terminal

Responsibility:

- Owns operator interaction surface behavior (open/attach flows).
- Bridges runtime attach metadata into usable human workflows.
- Supports opening one/all sessions based on orchestration state.

Key inputs:

- Session metadata + runtime attach information.

Key outputs:

- Human-accessible session interaction endpoint/window/tab.

What it does not own:

- Automated orchestration, tracker/SCM state transitions, or policy decisions.

## Mapping to V1 Slot Strategy

Reference has 7 visible slots above, while V1 will standardize around 6 canonical slots:

- `agent`, `scm`, `tracker`, `notifier`, `client`, `kanban`

Practical mapping:

- Reference `Runtime` and `Workspace` concerns are absorbed into V1 execution/worker and infrastructure design.
- Reference `Terminal` concern is absorbed into V1 `client` responsibility.
- V1 introduces `kanban` as a first-class slot not present as an explicit slot in the reference.

## Relevant Reference Files

Core contracts and plugin model:

- `.docs/agent-orchestrator/packages/core/src/types.ts`
- `.docs/agent-orchestrator/packages/core/src/plugin-registry.ts`

Core orchestration stability points:

- `.docs/agent-orchestrator/packages/core/src/session-manager.ts`
- `.docs/agent-orchestrator/packages/core/src/lifecycle-manager.ts`
- `.docs/agent-orchestrator/packages/core/src/orchestrator-prompt.ts`

Dashboard/API/service composition:

- `.docs/agent-orchestrator/packages/web/src/lib/services.ts`
- `.docs/agent-orchestrator/packages/web/src/app/api/events/route.ts`
- `.docs/agent-orchestrator/packages/web/src/app/api/sessions/route.ts`

Plugin implementations:

- `.docs/agent-orchestrator/packages/plugins/*`

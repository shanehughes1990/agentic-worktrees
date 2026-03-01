# Slice 03 — Orchestrator Supervisor

## Status

- Completed: **Completed**
- Reviewed At: **2026-02-28T00:00:00Z**

## Objective

Implement the central orchestration supervisor as an **agentic gatekeeper** for change quality and merge readiness.

The end goal for this slice is a supervisor that:

- Watches SCM and workflow signals continuously.
- Evaluates whether task-agent output is sufficient to merge.
- Decides multiple layers of merge readiness (checks, conflicts, review quality, policy compliance).
- Kicks back incomplete or unsafe work to task agents with explicit remediation actions.
- Coordinates rework loops for merge conflicts, review comments, and quality gaps.
- Triggers new task-agent workflows from incoming issues as part of the full operational cycle.

## AO-Informed Supervisor Patterns (Implementation Baseline)

Based on `.docs/agent-orchestrator` review, Slice 03 implementation must follow these patterns:

- **Deterministic transition engine** with explicit rule priority and typed reason codes.
- **Correlation-first event model** keyed by `run_id/task_id/job_id` for all decisions and transitions.
- **Append-only supervisor decision history** persisted in Postgres for replay, audit, and diagnostics.
- **Signal-driven evaluation loop** consuming execution, SCM, checkpoint, tracker, and issue-intake signals.
- **Escalation/reaction semantics** encoded as explicit actions (`retry`, `block`, `escalate`, `continue`, `request_rework`, `merge`, `refuse`), not inferred behavior.
- **Strict DDD placement**: interface submits signals, application orchestrates, domain decides, infrastructure persists/dispatches.

## Responsibility Boundary

- **Task agents** own task execution and PR construction against the instructed source branch.
- **Supervisor** owns merge governance and quality control:
  - Accept/reject merge readiness.
  - Route corrective work back to task agents.
  - Gate final merge execution.
  - Initiate task workflows from incoming issues.

## Tracker Intake Sources

Supervisor issue-intake orchestration in this slice is explicitly sourced from tracker ingestion pathways for:

- `github_issues` (GitHub Issues tracker source), and
- `local_json` (local taskboard source).

Additional tracker providers may exist as boundaries, but they are not required for this slice’s supervisor issue-intake completion target.

## Task Checklist

- [x] Define supervisor domain model (states, transitions, invariants) including merge-gate and rework-loop states.
- [x] Define supervisor event taxonomy with typed reason codes and action codes for merge/refuse/rework/escalation outcomes.
- [x] Implement deterministic policy priority order for conflicting rule matches across merge gates.
- [x] Implement application-layer policy engine for multi-layer merge decisions.
- [x] Implement correlation-ID propagation across supervisor decisions and rework loops.
- [x] Persist append-only supervisor decision history in Postgres (`supervisor_events`).
- [x] Implement decision query interfaces for `run_id/task_id/job_id` history retrieval.
- [x] Add deterministic transition tests over real worker/SCM/tracker/issue signal fixtures.
- [x] Expose supervisor decision history for query/subscription layers.
- [x] Implement kickback routing for merge conflicts, unresolved review comments, and policy violations back to task agents.
- [x] Implement final merge/refusal orchestration path where supervisor is the gatekeeper.
- [x] Implement issue-intake to task-agent kickoff flow under supervisor policy control.

## Deliverables

- Supervisor state model for workflows, merge gates, and rework loops.
- Decision policy model for attention zones, merge layers, and action transitions.
- Event taxonomy for supervisor decisions and lifecycle transitions.
- Postgres-backed append-only supervisor decision/event history.
- Correlation-ID propagation contract across supervisor flows.
- Query surface for historical decision replay by run/task/job.
- Supervisor-governed rework dispatch contract for task-agent remediation.
- Supervisor-governed issue-to-task kickoff contract.

## In Scope

- Domain/application contracts for supervisor state and decisions.
- Deterministic transition handling, merge readiness policy, and error classification.
- Event emission contracts consumed by stream/control-plane layers.
- Durable persistence of supervisor history for replay/audit.
- AO-aligned escalation/reaction and policy-evaluation semantics.
- Supervisor-triggered rework dispatch to task agents.
- Supervisor-triggered issue intake orchestration for task kickoff via tracker `github_issues` and `local_json` sources.

## Out of Scope

- Full provider adapter implementations beyond contracts needed by this slice.
- Final UI rendering logic.
- Full stream replay storage implementation (`stream_events`) beyond required supervisor persistence hooks.

## Acceptance Criteria

- Transition rules are deterministic and testable.
- Supervisor decisions can be queried/traced by correlation IDs.
- Policies are provider-agnostic and do not depend on concrete SDKs.
- Supervisor history survives process restarts via Postgres persistence.
- Policy decisions are explainable via reason codes + action records.
- Supervisor can gate merge outcomes with explicit merge/refuse/rework decisions.
- Supervisor can route failed merge gates (conflicts/comments/policy gaps) back to task agents with deterministic remediation actions.
- Supervisor can start task workflows from issue-intake signals under policy control.
- Issue-intake kickoff is demonstrably wired from tracker `github_issues` and `local_json` source pathways.

## Dependencies

- Slice 00 complete.
- Slice 01 complete.
- Slice 02 baseline complete.
- Slice 04 tracker signal contracts available for attention-zone and planning evaluation.
- Stream contract alignment from slice 05.

## Exit Check

This slice is complete when supervisor decisions can be explained, replayed, and audited from persisted event history, and the supervisor acts as the operational gatekeeper for:

- merge approval/refusal,
- task-agent rework routing,
- and issue-to-task kickoff,

with AO-informed deterministic policy semantics implemented under strict DDD boundaries.

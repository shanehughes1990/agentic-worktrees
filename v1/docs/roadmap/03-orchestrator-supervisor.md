# Slice 03 — Orchestrator Supervisor

## Objective

Implement the central orchestration supervisor responsible for policy decisions, global lifecycle tracking, and escalation triggers, with durable decision/event persistence in Postgres.

## AO-Informed Supervisor Patterns (Implementation Baseline)

Based on `.docs/agent-orchestrator` review, Slice 03 implementation must follow these patterns:

- **Deterministic transition engine** with explicit rule priority and typed reason codes.
- **Correlation-first event model** keyed by `run_id/task_id/job_id` for all decisions and transitions.
- **Append-only supervisor decision history** persisted in Postgres for replay, audit, and diagnostics.
- **Signal-driven evaluation loop** consuming execution, SCM, checkpoint, and tracker signals.
- **Escalation/reaction semantics** encoded as explicit actions (retry, block, escalate, continue), not inferred behavior.
- **Strict DDD placement**: interface submits signals, application orchestrates, domain decides, infrastructure persists/dispatches.

## Task Checklist

- [ ] Define supervisor domain model (states, transitions, invariants).
- [ ] Define supervisor event taxonomy with typed reason codes and action codes.
- [ ] Implement deterministic policy priority order for conflicting rule matches.
- [ ] Implement application-layer policy engine for attention-zone transitions.
- [ ] Implement correlation-ID propagation across supervisor decisions.
- [ ] Persist append-only supervisor decision history in Postgres (`supervisor_events`).
- [ ] Implement decision query interfaces for `run_id/task_id/job_id` history retrieval.
- [ ] Add deterministic transition tests over real worker/SCM/tracker signal fixtures.
- [ ] Expose supervisor decision history for query/subscription layers.

## Deliverables

- Supervisor state model for sessions/workflows.
- Decision policy model for attention zones and action transitions.
- Event taxonomy for supervisor decisions and lifecycle transitions.
- Postgres-backed append-only supervisor decision/event history.
- Correlation-ID propagation contract across supervisor flows.
- Query surface for historical decision replay by run/task/job.

## In Scope

- Domain/application contracts for supervisor state and decisions.
- Deterministic transition handling and error classification.
- Event emission contracts consumed by stream/control-plane layers.
- Durable persistence of supervisor history for replay/audit.
- AO-aligned escalation/reaction and policy-evaluation semantics.

## Out of Scope

- Full provider adapter implementations.
- Final UI rendering logic.
- Full stream replay storage implementation (`stream_events`) beyond required supervisor persistence hooks.

## Acceptance Criteria

- Transition rules are deterministic and testable.
- Supervisor decisions can be queried/traced by correlation IDs.
- Policies are provider-agnostic and do not depend on concrete SDKs.
- Supervisor history survives process restarts via Postgres persistence.
- Policy decisions are explainable via reason codes + action records.

## Dependencies

- Slice 00 complete.
- Slice 01 complete.
- Slice 02 baseline complete.
- Slice 04 tracker signal contracts available for attention-zone evaluation.
- Stream contract alignment from slice 05.

## Exit Check

This slice is complete when supervisor decisions can be explained, replayed, and audited from persisted event history, with AO-informed deterministic policy semantics implemented under strict DDD boundaries.
